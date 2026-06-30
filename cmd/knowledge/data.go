package knowledge

import (
	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

// NewCmdKnowledgeData groups operations on individual data entries (chunks/QA)
// inside a knowledge base file.
func NewCmdKnowledgeData(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data",
		Short: "Manage data entries inside a knowledge base",
	}

	cmd.AddCommand(NewCmdKnowledgeDataDelete(f, nil))

	return cmd
}
