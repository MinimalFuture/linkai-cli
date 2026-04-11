package plugin

import (
	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

func NewCmdPlugin(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage and execute plugins",
	}

	cmd.AddCommand(NewCmdPluginList(f, nil))
	cmd.AddCommand(NewCmdPluginDetail(f, nil))
	cmd.AddCommand(NewCmdPluginExec(f, nil))

	return cmd
}
