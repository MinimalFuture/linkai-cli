package knowledge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type DataDeleteOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	DryRun  bool
	Force   bool

	Code   string
	FileID string
	ID     string
}

func NewCmdKnowledgeDataDelete(f *cmdutil.Factory, runF func(*DataDeleteOptions) error) *cobra.Command {
	opts := &DataDeleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <code>",
		Short: "Delete a data entry from a knowledge base file",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.KnowledgeDelete.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			if runF != nil {
				return runF(opts)
			}
			return dataDeleteRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "skip confirmation prompt")
	cmd.Flags().StringVar(&opts.FileID, "file-id", "", "file the entry belongs to (required)")
	cmd.Flags().StringVar(&opts.ID, "id", "", "data entry id to delete (required)")
	_ = cmd.MarkFlagRequired("file-id")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func dataDeleteRun(opts *DataDeleteOptions) error {
	f := opts.Factory

	body := map[string]string{
		"code":    opts.Code,
		"file_id": opts.FileID,
		"id":      opts.ID,
	}

	if opts.DryRun {
		return output.PrintDryRun(f.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/api/cli/knowledge/data/delete",
			Body:   body,
		})
	}

	if !opts.Force && f.IOStreams.IsStdinTerminal {
		fmt.Fprintf(f.IOStreams.ErrOut, "Delete data entry %q from file %q? This cannot be undone. [y/N] ", opts.ID, opts.FileID)
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

	_, err = client.Post(opts.Ctx, "/api/cli/knowledge/data/delete", body)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge data: %w", err)
	}

	fmt.Fprintf(f.IOStreams.Out, "Data entry %q deleted.\n", opts.ID)
	return nil
}
