package workflow

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

	Name string
	Desc string
}

func NewCmdWorkflowCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a blank workflow",
		Long: "Create a blank workflow. Only --name is required. Orchestration " +
			"(nodes and edges) is done in the console — the returned link opens " +
			"the workflow editor.",
		Annotations: map[string]string{
			permission.RequiredKey: permission.WorkflowCreate.String(),
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
	cmd.Flags().StringVar(&opts.Name, "name", "", "workflow name (required)")
	cmd.Flags().StringVar(&opts.Desc, "desc", "", "workflow description")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func createRun(opts *CreateOptions) error {
	if err := validate.RejectControlChars("name", opts.Name); err != nil {
		return err
	}

	body := map[string]interface{}{"name": opts.Name}
	if opts.Desc != "" {
		body["description"] = opts.Desc
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/workflow/create",
			Body:   body,
		})
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Post(opts.Ctx, "/cli/workflow/create", body)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	var result struct {
		Code  string                 `json:"code"`
		Name  string                 `json:"name"`
		Links map[string]interface{} `json:"links,omitempty"`
	}
	if err := resp.Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintf(opts.Factory.IOStreams.Out, "Workflow created: %s (code: %s)\n", result.Name, result.Code)
	output.PrintLinks(opts.Factory.IOStreams.Out, result.Links)
	return nil
}
