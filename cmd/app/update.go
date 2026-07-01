package app

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

	Code         string
	Name         string
	Desc         string
	Prompt       string
	Introduction string

	nameSet   bool
	descSet   bool
	promptSet bool
	introSet  bool
}

func NewCmdAppUpdate(f *cmdutil.Factory, runF func(*UpdateOptions) error) *cobra.Command {
	opts := &UpdateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "update <code>",
		Short: "Update an application (only provided fields change)",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.AppUpdate.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			opts.nameSet = cmd.Flags().Changed("name")
			opts.descSet = cmd.Flags().Changed("desc")
			opts.promptSet = cmd.Flags().Changed("prompt")
			opts.introSet = cmd.Flags().Changed("intro")
			if runF != nil {
				return runF(opts)
			}
			return updateRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVar(&opts.Name, "name", "", "new application name")
	cmd.Flags().StringVar(&opts.Desc, "desc", "", "new description")
	cmd.Flags().StringVar(&opts.Prompt, "prompt", "", "new system prompt / persona")
	cmd.Flags().StringVar(&opts.Introduction, "intro", "", "new opening message")

	return cmd
}

func updateRun(opts *UpdateOptions) error {
	if !opts.nameSet && !opts.descSet && !opts.promptSet && !opts.introSet {
		return output.ErrValidation("nothing to update: provide at least one of --name/--desc/--prompt/--intro")
	}
	if opts.nameSet {
		if err := validate.RejectControlChars("name", opts.Name); err != nil {
			return err
		}
	}

	body := map[string]interface{}{"code": opts.Code}
	if opts.nameSet {
		body["name"] = opts.Name
	}
	if opts.descSet {
		body["description"] = opts.Desc
	}
	if opts.promptSet {
		body["prompt"] = opts.Prompt
	}
	if opts.introSet {
		body["introduction"] = opts.Introduction
	}

	if opts.DryRun {
		return output.PrintDryRun(opts.Factory.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/app/update",
			Body:   body,
		})
	}

	client, err := opts.Factory.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Post(opts.Ctx, "/cli/app/update", body)
	if err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}

	if opts.JSON {
		var result map[string]interface{}
		if err := resp.Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		return output.PrintJSON(opts.Factory.IOStreams.Out, result)
	}

	fmt.Fprintf(opts.Factory.IOStreams.Out, "Application %q updated.\n", opts.Code)
	return nil
}
