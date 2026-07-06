package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
	"github.com/MinimalFuture/linkai-cli/internal/validate"
)

// tableField mirrors the backend field shape: {name, type, comment}.
type tableField struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Comment string `json:"comment,omitempty"`
}

type CreateTableOptions struct {
	Factory   *cmdutil.Factory
	Ctx       context.Context
	JSON      bool
	DryRun    bool
	Code      string
	Name      string
	Desc      string
	Fields    []string // repeated --field name:type[:comment]
	FieldsRaw string   // --fields-json (agent-friendly)
}

func NewCmdDatabaseCreateTable(f *cmdutil.Factory, runF func(*CreateTableOptions) error) *cobra.Command {
	opts := &CreateTableOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create-table <db_code>",
		Short: "Create a table in a builtin database",
		Long: "Create a table in a builtin database. Define columns with repeated " +
			"--field name:type[:comment], or pass --fields-json for a JSON array. " +
			"Types: text, text1024, longtext, number, decimal, datetime. An auto-" +
			"increment `id` primary key is added by the platform.",
		Example: `  linkai database create-table qD7EwFNj --name orders \
    --field order_no:text:order number --field amount:decimal --field created:datetime

  linkai database create-table qD7EwFNj --name orders \
    --fields-json '[{"name":"order_no","type":"text"},{"name":"amount","type":"decimal"}]' --json`,
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.DBWrite.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			if runF != nil {
				return runF(opts)
			}
			return createTableRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVar(&opts.Name, "name", "", "table name (required; no spaces)")
	cmd.Flags().StringVar(&opts.Desc, "description", "", "table description")
	cmd.Flags().StringArrayVar(&opts.Fields, "field", nil, "column as name:type[:comment] (repeatable)")
	cmd.Flags().StringVar(&opts.FieldsRaw, "fields-json", "", "columns as a JSON array of {name,type,comment}")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func createTableRun(opts *CreateTableOptions) error {
	if err := validate.RejectControlChars("name", opts.Name); err != nil {
		return err
	}
	if err := validate.RejectControlChars("description", opts.Desc); err != nil {
		return err
	}

	fields, err := parseFields(opts)
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return output.ErrValidation("%s", "at least one column is required: use --field name:type[:comment] or --fields-json")
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"code":        opts.Code,
		"name":        opts.Name,
		"description": opts.Desc,
		"fields":      fields,
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/database/create-table",
			Body:   body,
		})
	}

	resp, err := client.Post(opts.Ctx, "/cli/database/create-table", body)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	var result map[string]interface{}
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintf(opts.Factory.IOStreams.Out, "Table created: %s (database: %s)\n", opts.Name, opts.Code)
	return nil
}

// parseFields builds the field list from either --fields-json (takes
// precedence) or repeated --field name:type[:comment] flags.
func parseFields(opts *CreateTableOptions) ([]tableField, error) {
	if opts.FieldsRaw != "" {
		var fields []tableField
		if err := json.Unmarshal([]byte(opts.FieldsRaw), &fields); err != nil {
			return nil, output.ErrValidation("invalid --fields-json: %v", err)
		}
		for i, f := range fields {
			if strings.TrimSpace(f.Name) == "" || strings.TrimSpace(f.Type) == "" {
				return nil, output.ErrValidation("--fields-json[%d]: name and type are required", i)
			}
		}
		return fields, nil
	}

	fields := make([]tableField, 0, len(opts.Fields))
	for _, raw := range opts.Fields {
		// Split into at most 3 parts so a comment may itself contain ':'.
		parts := strings.SplitN(raw, ":", 3)
		if len(parts) < 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, output.ErrValidation("invalid --field %q: expected name:type[:comment]", raw)
		}
		f := tableField{Name: strings.TrimSpace(parts[0]), Type: strings.TrimSpace(parts[1])}
		if len(parts) == 3 {
			f.Comment = strings.TrimSpace(parts[2])
		}
		fields = append(fields, f)
	}
	return fields, nil
}
