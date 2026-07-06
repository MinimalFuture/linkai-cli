package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"

	accountCmd    "github.com/MinimalFuture/linkai-cli/cmd/account"
	appCmd        "github.com/MinimalFuture/linkai-cli/cmd/app"
	audioCmd      "github.com/MinimalFuture/linkai-cli/cmd/audio"
	authCmd       "github.com/MinimalFuture/linkai-cli/cmd/auth"
	chatCmd       "github.com/MinimalFuture/linkai-cli/cmd/chat"
	databaseCmd   "github.com/MinimalFuture/linkai-cli/cmd/database"
	imageCmd      "github.com/MinimalFuture/linkai-cli/cmd/image"
	knowledgeCmd  "github.com/MinimalFuture/linkai-cli/cmd/knowledge"
	modelCmd      "github.com/MinimalFuture/linkai-cli/cmd/model"
	pluginCmd     "github.com/MinimalFuture/linkai-cli/cmd/plugin"
	skillCmd      "github.com/MinimalFuture/linkai-cli/cmd/skill"
	updateCmd     "github.com/MinimalFuture/linkai-cli/cmd/update"
	videoCmd      "github.com/MinimalFuture/linkai-cli/cmd/video"
	workflowCmd   "github.com/MinimalFuture/linkai-cli/cmd/workflow"
	"github.com/MinimalFuture/linkai-cli/internal/auth"
	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/permission"
	"github.com/MinimalFuture/linkai-cli/internal/selfupdate"
)

// version and buildDate are set via ldflags at build time.
// Example: go build -ldflags "-X github.com/MinimalFuture/linkai-cli/cmd.version=v1.0.0 -X github.com/MinimalFuture/linkai-cli/cmd.buildDate=2026-04-12"
var (
	version   = "dev"
	buildDate = ""
)

func Execute() int {
	// Preserve our explicit registration order in `--help` (group commands by
	// concern in a deliberate order) instead of cobra's default alphabetical sort.
	cobra.EnableCommandSorting = false

	// hasUngrouped reports whether any visible (non-hidden) command lacks a
	// GroupID. The usage template uses it to render the "Additional Commands"
	// heading only when there is something to list (our help command is hidden).
	cobra.AddTemplateFunc("hasUngrouped", func(cmds []*cobra.Command) bool {
		for _, c := range cmds {
			if c.GroupID == "" && c.IsAvailableCommand() {
				return true
			}
		}
		return false
	})

	f := cmdutil.NewDefault()

	// Kick off a throttled, best-effort update check in the background so the
	// latest version is cached for the next run without adding latency here.
	// NotifyIfDue (at the end) only reads that cache.
	go selfupdate.RefreshLatestInBackground(resolveVersion())

	versionStr := resolveVersion()
	if buildDate != "" {
		versionStr += " (" + buildDate + ")"
	}

	rootCmd := &cobra.Command{
		Use:     "linkai",
		Short:   "LinkAI CLI - Command line interface for the LinkAI platform",
		Version: versionStr,
	}
	rootCmd.SilenceErrors = true
	// Hide cobra's auto-generated `completion` command; we don't ship shell
	// completion as a user-facing feature.
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// Keep the `help` command functional (`linkai help`, `linkai help account`)
	// but hide it from the command listing — the -h/--help flag is the primary
	// entry point and we don't want a near-empty "Additional Commands" section.
	// cobra's default usage template force-shows any command literally named
	// "help", so a custom template (below) is required to actually hide it.
	rootCmd.SetHelpCommand(newHelpCmd(rootCmd))
	rootCmd.SetUsageTemplate(usageTemplate)

	// Accept --json on EVERY command. The agent skill tells agents to always
	// append --json, so a command without a meaningful JSON mode (e.g. logout)
	// must still tolerate the flag instead of erroring with "unknown flag".
	// This persistent flag is the universal fallback; commands that declare
	// their own local --json shadow it and drive real JSON output. It is hidden
	// so it doesn't clutter every command's help.
	rootCmd.PersistentFlags().Bool("json", false, "output in JSON format")
	_ = rootCmd.PersistentFlags().MarkHidden("json")

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

	// Command groups organize `linkai --help` into meaningful sections instead
	// of one long alphabetical list. Each command is assigned a GroupID below.
	const (
		groupResources = "resources"
		groupAI        = "ai"
		groupAccount   = "account"
	)
	rootCmd.AddGroup(
		&cobra.Group{ID: groupResources, Title: "Resource management:"},
		&cobra.Group{ID: groupAI, Title: "AI capabilities:"},
		&cobra.Group{ID: groupAccount, Title: "Account & auth:"},
	)

	// 1. Core resource management
	addToGroup(rootCmd, groupResources,
		appCmd.NewCmdApp(f),
		knowledgeCmd.NewCmdKnowledge(f),
		databaseCmd.NewCmdDatabase(f),
		workflowCmd.NewCmdWorkflow(f),
		pluginCmd.NewCmdPlugin(f),
		modelCmd.NewCmdModel(f),
	)

	// 2. AI capability invocation
	addToGroup(rootCmd, groupAI,
		chatCmd.NewCmdChat(f, nil),
		imageCmd.NewCmdImage(f),
		videoCmd.NewCmdVideo(f),
		audioCmd.NewCmdAudio(f),
	)

	// 3. Account, recharge and authentication
	addToGroup(rootCmd, groupAccount,
		accountCmd.NewCmdAccount(f),
		authCmd.NewCmdAuth(f),
	)

	// Utility commands (ungrouped, shown under "Additional Commands"): the skill
	// bundle installer/reader and the self-updater need no login, so they stay
	// out of the groups above.
	rootCmd.AddCommand(skillCmd.NewCmdSkill(f))
	rootCmd.AddCommand(updateCmd.NewCmdUpdate(f, resolveVersion()))

	if err := rootCmd.Execute(); err != nil {
		output.PrintError(f.IOStreams.ErrOut, f.IOStreams.IsTerminal, err.Error())
		return output.ExitCodeFrom(err)
	}
	maybeNotifyUpdate(f)
	return 0
}

// maybeNotifyUpdate prints a one-line hint to stderr when a newer CLI version is
// available. The notice goes to stderr (never stdout), so it does not corrupt
// JSON parsed by agents; it is throttled to at most once per day and reads only
// the locally cached latest version (no network on this path).
func maybeNotifyUpdate(f *cmdutil.Factory) {
	if latest := selfupdate.NotifyIfDue(resolveVersion()); latest != "" {
		output.PrintUpdateNotice(f.IOStreams.ErrOut, resolveVersion(), latest)
	}
}

// usageTemplate mirrors cobra's default usage template with one change: it does
// NOT force-display a command literally named "help". This lets a Hidden help
// command stay functional while being omitted from the listing, and avoids a
// dangling empty "Additional Commands" heading.
const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if hasUngrouped $cmds}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// newHelpCmd builds a `help` command that behaves like cobra's default one
// (prints help for the target command, or root help when no args) but is
// hidden from the command listing.
func newHelpCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:    "help [command]",
		Short:  "Help about any command",
		Hidden: true,
		Run: func(c *cobra.Command, args []string) {
			target, _, err := root.Find(args)
			if target == nil || err != nil {
				c.Printf("Unknown help topic %#q\n", args)
				_ = root.Help()
				return
			}
			_ = target.Help()
		},
	}
}

// addToGroup assigns each command to the given help group and registers it.
func addToGroup(root *cobra.Command, groupID string, cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.GroupID = groupID
		root.AddCommand(c)
	}
}

// resolveVersion prefers the ldflags-injected version. When the binary is
// built without ldflags (notably `go install`), it falls back to the module
// version recorded in the build info so the user sees something meaningful
// instead of "dev".
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return version
}
