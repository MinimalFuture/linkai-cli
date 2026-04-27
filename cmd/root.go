package cmd

import (
	"github.com/spf13/cobra"

	accountCmd    "github.com/MinimalFuture/linkai-cli/cmd/account"
	appCmd        "github.com/MinimalFuture/linkai-cli/cmd/app"
	audioCmd      "github.com/MinimalFuture/linkai-cli/cmd/audio"
	authCmd       "github.com/MinimalFuture/linkai-cli/cmd/auth"
	chatCmd       "github.com/MinimalFuture/linkai-cli/cmd/chat"
	completionCmd "github.com/MinimalFuture/linkai-cli/cmd/completion"
	databaseCmd   "github.com/MinimalFuture/linkai-cli/cmd/database"
	imageCmd      "github.com/MinimalFuture/linkai-cli/cmd/image"
	knowledgeCmd  "github.com/MinimalFuture/linkai-cli/cmd/knowledge"
	modelCmd      "github.com/MinimalFuture/linkai-cli/cmd/model"
	pluginCmd     "github.com/MinimalFuture/linkai-cli/cmd/plugin"
	scoreCmd      "github.com/MinimalFuture/linkai-cli/cmd/score"
	videoCmd      "github.com/MinimalFuture/linkai-cli/cmd/video"
	workflowCmd   "github.com/MinimalFuture/linkai-cli/cmd/workflow"
	"github.com/MinimalFuture/linkai-cli/internal/auth"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
)

// version and buildDate are set via ldflags at build time.
// Example: go build -ldflags "-X github.com/MinimalFuture/linkai-cli/cmd.version=v1.0.0 -X github.com/MinimalFuture/linkai-cli/cmd.buildDate=2026-04-12"
var (
	version   = "dev"
	buildDate = ""
)

func Execute() int {
	f := cmdutil.NewDefault()

	versionStr := version
	if buildDate != "" {
		versionStr += " (" + buildDate + ")"
	}

	rootCmd := &cobra.Command{
		Use:     "linkai",
		Short:   "LinkAI CLI - Command line tool for the LinkAI platform",
		Version: versionStr,
	}
	rootCmd.SilenceErrors = true

	// PersistentPreRunE runs before every subcommand.
	// It silences usage on error and enforces declared permission checks.
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		required, ok := cmd.Annotations[permission.RequiredKey]
		if !ok {
			return nil
		}

		token := auth.GetStoredToken()
		if token == nil {
			return output.ErrAuth("not logged in: run 'linkai auth login'")
		}
		return permission.Check(token, permission.Permission(required))
	}

	rootCmd.AddCommand(completionCmd.NewCmdCompletion())
	rootCmd.AddCommand(accountCmd.NewCmdAccount(f))
	rootCmd.AddCommand(appCmd.NewCmdApp(f))
	rootCmd.AddCommand(audioCmd.NewCmdAudio(f))
	rootCmd.AddCommand(authCmd.NewCmdAuth(f))
	rootCmd.AddCommand(databaseCmd.NewCmdDatabase(f))
	rootCmd.AddCommand(imageCmd.NewCmdImage(f))
	rootCmd.AddCommand(knowledgeCmd.NewCmdKnowledge(f))
	rootCmd.AddCommand(modelCmd.NewCmdModel(f))
	rootCmd.AddCommand(videoCmd.NewCmdVideo(f))
	rootCmd.AddCommand(chatCmd.NewCmdChat(f, nil))
	rootCmd.AddCommand(pluginCmd.NewCmdPlugin(f))
	rootCmd.AddCommand(scoreCmd.NewCmdScore(f))
	rootCmd.AddCommand(workflowCmd.NewCmdWorkflow(f))

	if err := rootCmd.Execute(); err != nil {
		output.PrintError(f.IOStreams.ErrOut, f.IOStreams.IsTerminal, err.Error())
		return output.ExitCodeFrom(err)
	}
	return 0
}
