package knowledge

import (
	"github.com/spf13/cobra"

	"github.com/yjr/linkai-cli/internal/cmdutil"
)

// NewCmdKnowledge creates the knowledge command with subcommands.
func NewCmdKnowledge(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Manage knowledge bases",
	}

	cmd.AddCommand(NewCmdKnowledgeCreate(f, nil))
	cmd.AddCommand(NewCmdKnowledgeList(f, nil))
	cmd.AddCommand(NewCmdKnowledgeFiles(f, nil))
	cmd.AddCommand(NewCmdKnowledgeDelete(f, nil))
	cmd.AddCommand(NewCmdKnowledgeSearch(f, nil))

	return cmd
}
