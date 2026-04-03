package knowledge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yjr/linkai-cli/internal/cmdutil"
)

type DeleteOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	Code    string
	Force   bool
}

func NewCmdKnowledgeDelete(f *cmdutil.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <code>",
		Short: "Delete a knowledge base",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			cmdutil.RequiredScopeKey: "knowledge:delete",
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

	return cmd
}

func deleteRun(opts *DeleteOptions) error {
	f := opts.Factory

	if !opts.Force && f.IOStreams.IsStdinTerminal {
		fmt.Fprintf(f.IOStreams.ErrOut, "Delete knowledge base %q? This cannot be undone. [y/N] ", opts.Code)
		var confirm string
		fmt.Fscan(f.IOStreams.In, &confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Fprintln(f.IOStreams.ErrOut, "Aborted.")
			return nil
		}
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	_, err = client.Post(opts.Ctx, "/api/cli/knowledge/delete", map[string]string{
		"code": opts.Code,
	})
	if err != nil {
		return fmt.Errorf("failed to delete knowledge base: %w", err)
	}

	fmt.Fprintf(f.IOStreams.Out, "Knowledge base %q deleted.\n", opts.Code)
	return nil
}
