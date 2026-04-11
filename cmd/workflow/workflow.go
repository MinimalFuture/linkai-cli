package workflow

import (
	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

func NewCmdWorkflow(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage and run workflows",
	}

	cmd.AddCommand(NewCmdWorkflowList(f, nil))
	cmd.AddCommand(NewCmdWorkflowRun(f, nil))

	return cmd
}
