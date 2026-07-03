// Package update implements `linkai update`: check for and install the latest
// CLI version. It detects how the CLI was installed and delegates the upgrade to
// the matching package manager or the install script, then refreshes the agent
// skill. Use --check to only report availability, and --json for agents.
package update

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	"github.com/MinimalFuture/linkai-cli/internal/selfupdate"
	skillpkg "github.com/MinimalFuture/linkai-cli/internal/skill"
)

// Options holds inputs for the update command.
type Options struct {
	Factory *cmdutil.Factory
	Version string // current CLI version (injected from cmd package)
	JSON    bool
	Check   bool
	Force   bool
}

// NewCmdUpdate creates the `update` command. currentVersion is the running
// binary's version string, resolved by the root command.
func NewCmdUpdate(f *cmdutil.Factory, currentVersion string) *cobra.Command {
	opts := &Options{Factory: f, Version: currentVersion}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the CLI to the latest version",
		Long: `Update the linkai CLI to the latest published version.

The install method is detected automatically:
  - npm install     → npm install -g linkai-cli@<version>
  - Homebrew        → brew upgrade --cask linkai
  - go install      → go install ...@<version>
  - script/manual   → re-runs the install script (macOS/Linux)

Use --check to only check for updates without installing.
Use --json for structured output (for agents and scripts).`,
		Example: `  linkai update            # update to the latest version
  linkai update --check    # only report whether an update is available
  linkai update --json     # machine-readable output`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output as JSON")
	cmd.Flags().BoolVar(&opts.Check, "check", false, "only check for updates, do not install")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "reinstall even if already up to date")
	return cmd
}

func run(ctx context.Context, opts *Options) error {
	io := opts.Factory.IOStreams
	cur := opts.Version

	latest, err := selfupdate.FetchLatest(ctx)
	if err != nil {
		return output.ErrNetwork("failed to check the latest version: %v", err)
	}

	upToDate := !selfupdate.IsNewer(latest, cur)
	detect := selfupdate.Detect()

	// --check: report only, never install.
	if opts.Check {
		return reportCheck(opts, cur, latest, upToDate, detect)
	}

	// Already current and not forced: nothing to do.
	if upToDate && !opts.Force {
		return reportUpToDate(opts, cur, latest)
	}

	// Windows script installs can't be auto-updated from Go (no POSIX shell to
	// pipe the installer into); guide the user instead.
	if detect.Method == selfupdate.MethodScript && selfupdate.IsWindows() {
		return reportManualWindows(opts, cur, latest)
	}

	if !opts.JSON {
		fmt.Fprintf(io.ErrOut, "Updating linkai %s -> %s via %s ...\n", cur, latest, detect.Method)
	}

	var cmdIO selfupdate.CommandIO
	if !opts.JSON {
		// Stream the package-manager output live to stderr in human mode.
		cmdIO.Stream = io.ErrOut
	}

	res := selfupdate.Upgrade(ctx, detect.Method, latest, cmdIO)
	if res.Err != nil {
		return reportFailure(opts, cur, latest, detect, res)
	}

	// Refresh the embedded skill so it matches the new binary. Best-effort:
	// a skill sync failure should not fail the whole update.
	skillsErr := refreshSkill()

	return reportUpdated(opts, cur, latest, detect, skillsErr)
}

func refreshSkill() error {
	_, err := skillpkg.Install("")
	return err
}
