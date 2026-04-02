package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	authCmd "github.com/yjr/linkai-cli/cmd/auth"
	larkauth "github.com/yjr/linkai-cli/internal/auth"
	"github.com/yjr/linkai-cli/internal/cmdutil"
)

var version = "dev"

func Execute() int {
	f := cmdutil.NewDefault()

	rootCmd := &cobra.Command{
		Use:     "linkai",
		Short:   "LinkAI CLI - Command line tool for the LinkAI platform",
		Version: version,
	}
	rootCmd.SilenceErrors = true

	// PersistentPreRunE runs before every subcommand.
	// It silences usage on error and enforces scope-based permission checks.
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		requiredScope, hasScope := cmd.Annotations[cmdutil.RequiredScopeKey]
		if !hasScope {
			return nil
		}

		token := larkauth.GetStoredToken()
		if token == nil {
			return errors.New("not logged in: run 'linkai auth login'")
		}
		return cmdutil.CheckScope(token, requiredScope)
	}

	rootCmd.AddCommand(authCmd.NewCmdAuth(f))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(f.IOStreams.ErrOut, "Error:", err)
		return 1
	}
	return 0
}
