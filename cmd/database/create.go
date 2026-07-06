package database

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
	"github.com/MinimalFuture/linkai-cli/internal/validate"
)

type CreateOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	DryRun  bool
	Name    string
	Desc    string
}

func NewCmdDatabaseCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a builtin (platform-hosted) database",
		Long: "Create a builtin database — a platform-hosted store, with no external " +
			"connection required. Add tables and data from the console afterwards; " +
			"query it with `linkai database exec`.",
		Annotations: map[string]string{
			permission.RequiredKey: permission.DBWrite.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVar(&opts.Name, "name", "", "database name (required)")
	cmd.Flags().StringVar(&opts.Desc, "description", "", "database description")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func createRun(opts *CreateOptions) error {
	if err := validate.RejectControlChars("name", opts.Name); err != nil {
		return err
	}
	if err := validate.RejectControlChars("description", opts.Desc); err != nil {
		return err
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"name":        opts.Name,
		"description": opts.Desc,
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/database/create",
			Body:   body,
		})
	}

	resp, err := client.Post(opts.Ctx, "/cli/database/create", body)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	var result map[string]interface{}
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	code, _ := result["code"].(string)
	fmt.Fprintf(opts.Factory.IOStreams.Out, "Database created: %s (code: %s)\n", opts.Name, code)
	output.PrintLinks(opts.Factory.IOStreams.Out, result["links"])
	return nil
}
