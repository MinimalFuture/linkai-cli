package auth

import (
	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
)

// NewCmdAuth creates the auth command with subcommands.
func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication management",
	}

	cmd.AddCommand(NewCmdAuthLogin(f, nil))
	cmd.AddCommand(NewCmdAuthLogout(f, nil))
	cmd.AddCommand(NewCmdAuthStatus(f, nil))

	return cmd
}
