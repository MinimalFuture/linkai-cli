package database

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/MinimalFuture/linkai-cli/internal/auth"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
	"github.com/MinimalFuture/linkai-cli/internal/validate"
)

type ExecOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	DryRun  bool
	Code    string
	SQL     string
}

type ExecResult struct {
	Status       string                   `json:"status"`
	Data         []map[string]interface{} `json:"data,omitempty"`
	Total        int                      `json:"total,omitempty"`
	AffectedRows int                      `json:"affected_rows,omitempty"`
	Error        string                   `json:"error,omitempty"`
}

func NewCmdDatabaseExec(f *cmdutil.Factory, runF func(*ExecOptions) error) *cobra.Command {
	opts := &ExecOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "exec <code> <sql>",
		Short: "Execute SQL against a database",
		Long: `Execute a SQL statement against a database.

SELECT queries require the db:read permission.
Write operations (INSERT/UPDATE/DELETE) additionally require db:write,
checked at runtime once the SQL has been classified.`,
		Args: cobra.ExactArgs(2),
		Annotations: map[string]string{
			permission.RequiredKey: permission.DBRead.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			opts.SQL = args[1]
			if runF != nil {
				return runF(opts)
			}
			return execRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")

	return cmd
}

// isDangerousSQL returns true for DDL operations that should not be allowed via CLI.
func isDangerousSQL(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	for _, prefix := range []string{"DROP", "TRUNCATE", "ALTER"} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

// isMutatingSQL returns true when the SQL is a write operation (not SELECT).
func isMutatingSQL(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	for _, prefix := range []string{"INSERT", "UPDATE", "DELETE"} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

func execRun(opts *ExecOptions) error {
	// Validate SQL input for control characters
	if err := validate.RejectControlChars("sql", opts.SQL); err != nil {
		return output.ErrValidation("%v", err)
	}

	// Reject dangerous DDL operations
	if isDangerousSQL(opts.SQL) {
		return output.ErrValidation("dangerous SQL operation detected: DROP/TRUNCATE/ALTER are not allowed via CLI")
	}

	// Client-side permission pre-check for mutating SQL so the user gets the
	// normal re-login hint before the request is even sent.
	if isMutatingSQL(opts.SQL) {
		token := auth.GetStoredToken()
		if err := permission.Check(token, permission.DBWrite); err != nil {
			return err
		}
	}

	body := map[string]string{
		"code": opts.Code,
		"sql":  opts.SQL,
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/database/exec",
			Body:   body,
		})
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Post(opts.Ctx, "/cli/database/exec", body)
	if err != nil {
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	var result ExecResult
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	if result.Status == "FAILED" {
		return fmt.Errorf("SQL error: %s", result.Error)
	}

	if result.Data != nil {
		// SELECT result
		if len(result.Data) == 0 {
			fmt.Fprintln(opts.Factory.IOStreams.Out, "No rows returned.")
			return nil
		}
		// Collect column names in stable order from first row
		firstRow := result.Data[0]
		cols := make([]string, 0, len(firstRow))
		for k := range firstRow {
			cols = append(cols, k)
		}
		sort.Strings(cols)

		rows := make([][]string, 0, len(result.Data))
		for _, row := range result.Data {
			cells := make([]string, 0, len(cols))
			for _, col := range cols {
				v := row[col]
				if v == nil {
					cells = append(cells, "NULL")
				} else {
					cells = append(cells, fmt.Sprintf("%v", v))
				}
			}
			rows = append(rows, cells)
		}
		output.PrintTable(opts.Factory.IOStreams.Out, cols, rows)
		fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "\n%d row(s)\n", result.Total)
	} else {
		// DML result
		fmt.Fprintf(opts.Factory.IOStreams.Out, "OK  affected rows: %d\n", result.AffectedRows)
	}

	return nil
}
