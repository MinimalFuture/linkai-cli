package workflow

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
)

type ListOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
}

type WorkflowItem struct {
	AppCode     string `json:"appCode"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewCmdWorkflowList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflows",
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "workflow:read",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")

	return cmd
}

func listRun(opts *ListOptions) error {
	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(opts.Ctx, "/api/cli/workflow/list", nil)
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	var items []WorkflowItem
	if err := resp.Decode(&items); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if opts.JSON {
		return output.PrintJSON(opts.Factory.IOStreams.Out, items)
	}

	if len(items) == 0 {
		fmt.Fprintln(opts.Factory.IOStreams.Out, "No workflows found.")
		return nil
	}

	headers := []string{"CODE", "NAME", "DESCRIPTION"}
	rows := make([][]string, 0, len(items))
	for _, wf := range items {
		desc := []rune(wf.Description)
		if len(desc) > 40 {
			desc = append(desc[:40], []rune("...")...)
		}
		rows = append(rows, []string{wf.AppCode, wf.Name, string(desc)})
	}
	output.PrintTable(opts.Factory.IOStreams.Out, headers, rows)

	return nil
}
