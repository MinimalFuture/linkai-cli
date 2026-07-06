package knowledge

import (
	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

// NewCmdKnowledgeFile groups operations on whole files inside a knowledge base.
func NewCmdKnowledgeFile(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Manage whole files in a knowledge base",
	}

	cmd.AddCommand(NewCmdKnowledgeFileDelete(f, nil))

	return cmd
}
