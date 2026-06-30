package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type DeleteOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	DryRun  bool
	Code    string
	Force   bool
}

func NewCmdAppDelete(f *cmdutil.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <code>",
		Short: "Delete an application",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.AppDelete.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			if runF != nil {
				return runF(opts)
			}
			return deleteRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Force, "force", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")

	return cmd
}

func deleteRun(opts *DeleteOptions) error {
	f := opts.Factory

	if opts.DryRun {
		return output.PrintDryRun(f.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/api/cli/app/delete",
			Body:   map[string]string{"code": opts.Code},
		})
	}

	if !opts.Force && f.IOStreams.IsStdinTerminal {
		fmt.Fprintf(f.IOStreams.ErrOut, "Delete application %q? This cannot be undone. [y/N] ", opts.Code)
		var confirm string
		_, _ = fmt.Fscan(f.IOStreams.In, &confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Fprintln(f.IOStreams.ErrOut, "Aborted.")
			return nil
		}
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	_, err = client.Post(opts.Ctx, "/api/cli/app/delete", map[string]string{
		"code": opts.Code,
	})
	if err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}

	fmt.Fprintf(f.IOStreams.Out, "Application %q deleted.\n", opts.Code)
	return nil
}
