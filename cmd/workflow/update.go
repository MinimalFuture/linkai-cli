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

type UpdateOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	JSON    bool
	DryRun  bool

	Code string
	Name string
	Desc string

	nameSet bool
	descSet bool
}

func NewCmdWorkflowUpdate(f *cmdutil.Factory, runF func(*UpdateOptions) error) *cobra.Command {
	opts := &UpdateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "update <code>",
		Short: "Update a workflow's name or description",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.WorkflowUpdate.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			opts.nameSet = cmd.Flags().Changed("name")
			opts.descSet = cmd.Flags().Changed("desc")
			if runF != nil {
				return runF(opts)
			}
			return updateRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVar(&opts.Name, "name", "", "new workflow name")
	cmd.Flags().StringVar(&opts.Desc, "desc", "", "new workflow description")

	return cmd
}

func updateRun(opts *UpdateOptions) error {
	if !opts.nameSet && !opts.descSet {
		return output.ErrValidation("nothing to update: provide --name and/or --desc")
	}

	body := map[string]interface{}{"code": opts.Code}
	if opts.nameSet {
		if err := validate.RejectControlChars("name", opts.Name); err != nil {
			return err
		}
		body["name"] = opts.Name
	}
	if opts.descSet {
		body["description"] = opts.Desc
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/workflow/update",
			Body:   body,
		})
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	if _, err := client.Post(opts.Ctx, "/cli/workflow/update", body); err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	fmt.Fprintf(opts.Factory.IOStreams.Out, "Workflow %q updated.\n", opts.Code)
	return nil
}
