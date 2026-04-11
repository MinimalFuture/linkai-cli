package database

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

type TablesOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	Code    string
}

type TableItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TablesResult struct {
	List []TableItem `json:"list"`
}

func NewCmdDatabaseTables(f *cmdutil.Factory, runF func(*TablesOptions) error) *cobra.Command {
	opts := &TablesOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "tables <code>",
		Short: "List tables in a database",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "db:read",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			if runF != nil {
				return runF(opts)
			}
			return tablesRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")

	return cmd
}

func tablesRun(opts *TablesOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("code", opts.Code)

	resp, err := client.Get(opts.Ctx, "/api/cli/database/tables", params)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	var result TablesResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	if len(result.List) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No tables found.")
		return nil
	}

	headers := []string{"NAME", "DESCRIPTION"}
	rows := make([][]string, 0, len(result.List))
	for _, t := range result.List {
		desc := []rune(t.Description)
		if len(desc) > 50 {
			desc = append(desc[:50], []rune("...")...)
		}
		rows = append(rows, []string{t.Name, string(desc)})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)

	return nil
}
