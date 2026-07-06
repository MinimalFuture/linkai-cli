package knowledge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

type FileDeleteOptions struct {
	Factory *cmdutil.Factory
	Ctx     context.Context
	DryRun  bool
	Force   bool

	Code   string
	FileID string
}

func NewCmdKnowledgeFileDelete(f *cmdutil.Factory, runF func(*FileDeleteOptions) error) *cobra.Command {
	opts := &FileDeleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <kb_code>",
		Short: "Delete a whole file from a knowledge base",
		Long: "Delete an entire file (and all its entries) from a knowledge base, identified " +
			"by its file id. Get file ids from `linkai knowledge files <kb_code>`. This differs " +
			"from `knowledge data delete`, which removes a single entry inside a file.",
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			permission.RequiredKey: permission.KnowledgeDelete.String(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			opts.Code = args[0]
			if runF != nil {
				return runF(opts)
			}
			return fileDeleteRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print request without executing")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "skip confirmation prompt")
	cmd.Flags().StringVar(&opts.FileID, "file-id", "", "id of the file to delete (required)")
	_ = cmd.MarkFlagRequired("file-id")

	return cmd
}

func fileDeleteRun(opts *FileDeleteOptions) error {
	f := opts.Factory

	body := map[string]string{
		"code":    opts.Code,
		"file_id": opts.FileID,
	}

	if opts.DryRun {
		return output.PrintDryRun(f.IOStreams.Out, output.DryRunInfo{
			Method: "POST",
			URL:    "/cli/knowledge/file/delete",
			Body:   body,
		})
	}

	if !opts.Force && f.IOStreams.IsStdinTerminal {
		fmt.Fprintf(f.IOStreams.ErrOut, "Delete file %q and all its entries? This cannot be undone. [y/N] ", opts.FileID)
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

	_, err = client.Post(opts.Ctx, "/cli/knowledge/file/delete", body)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	fmt.Fprintf(f.IOStreams.Out, "File %q deleted.\n", opts.FileID)
	return nil
}
