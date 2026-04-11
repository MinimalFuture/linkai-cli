package account

import (
	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

// NewCmdAccount creates the account command with subcommands.
func NewCmdAccount(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage account information",
	}

	cmd.AddCommand(NewCmdAccountInfo(f, nil))

	return cmd
}
