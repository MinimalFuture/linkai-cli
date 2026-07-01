package knowledge

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
	Code    string
	Name    string
	Desc    string

	// nameSet / descSet track whether the flag was explicitly provided so we
	// only send fields the user actually wants to change.
	nameSet bool
	descSet bool
}

func NewCmdKnowledgeUpdate(f *cmdutil.Factory, runF func(*UpdateOptions) error) *cobra.Command {
	opts := &UpdateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "update <code>",
		Short: "Update a knowledge base's name or description",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.KnowledgeUpdate.String(),
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
	cmd.Flags().StringVar(&opts.Name, "name", "", "new knowledge base name")
	cmd.Flags().StringVar(&opts.Desc, "desc", "", "new knowledge base description")

	return cmd
}

func updateRun(opts *UpdateOptions) error {
	if !opts.nameSet && !opts.descSet {
		return output.ErrValidation("nothing to update: provide --name and/or --desc")
	}
	if opts.nameSet {
		if err := validate.RejectControlChars("name", opts.Name); err != nil {
			return err
		}
	}
	if opts.descSet {
		if err := validate.RejectControlChars("desc", opts.Desc); err != nil {
			return err
		}
	}

	body := map[string]interface{}{"code": opts.Code}
	if opts.nameSet {
		body["name"] = opts.Name
	}
	if opts.descSet {
		body["desc"] = opts.Desc
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/knowledge/update",
			Body:   body,
		})
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Post(opts.Ctx, "/cli/knowledge/update", body)
	if err != nil {
		return fmt.Errorf("failed to update knowledge base: %w", err)
	}

	if opts.JSON {
		var result map[string]interface{}
		if err := resp.Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintf(opts.Factory.IOStreams.Out, "Knowledge base %q updated.\n", opts.Code)
	return nil
}
