package app

import (
	"github.com/spf13/cobra"

	"github.com/yjr/linkai-cli/internal/cmdutil"
)

// NewCmdApp creates the app command with subcommands.
func NewCmdApp(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage applications",
	}

	cmd.AddCommand(NewCmdAppList(f, nil))

	return cmd
}
