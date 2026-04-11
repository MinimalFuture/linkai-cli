package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	accountCmd  "github.com/yjr/linkai-cli/cmd/account"
	appCmd      "github.com/yjr/linkai-cli/cmd/app"
	audioCmd    "github.com/yjr/linkai-cli/cmd/audio"
	authCmd     "github.com/yjr/linkai-cli/cmd/auth"
	databaseCmd "github.com/yjr/linkai-cli/cmd/database"
	imageCmd    "github.com/yjr/linkai-cli/cmd/image"
	knowledgeCmd "github.com/yjr/linkai-cli/cmd/knowledge"
	modelCmd    "github.com/yjr/linkai-cli/cmd/model"
	videoCmd    "github.com/yjr/linkai-cli/cmd/video"
	"github.com/yjr/linkai-cli/internal/auth"
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

		token := auth.GetStoredToken()
		if token == nil {
			return errors.New("not logged in: run 'linkai auth login'")
		}
		return cmdutil.CheckScope(token, requiredScope)
	}

	rootCmd.AddCommand(accountCmd.NewCmdAccount(f))
	rootCmd.AddCommand(appCmd.NewCmdApp(f))
	rootCmd.AddCommand(audioCmd.NewCmdAudio(f))
	rootCmd.AddCommand(authCmd.NewCmdAuth(f))
	rootCmd.AddCommand(databaseCmd.NewCmdDatabase(f))
	rootCmd.AddCommand(imageCmd.NewCmdImage(f))
	rootCmd.AddCommand(knowledgeCmd.NewCmdKnowledge(f))
	rootCmd.AddCommand(modelCmd.NewCmdModel(f))
	rootCmd.AddCommand(videoCmd.NewCmdVideo(f))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(f.IOStreams.ErrOut, "Error:", err)
		return 1
	}
	return 0
}
