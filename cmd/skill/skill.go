// Package skill implements the `linkai skill` command group, which installs and
// reads the agent-readable skill content embedded in the CLI binary. Agents
// discover skills by scanning their skills directory, so `skill install` copies
// the embedded docs there; `list` and `read` expose the same content directly.
package skill

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MinimalFuture/linkai-cli/internal/cmdutil"
	"github.com/MinimalFuture/linkai-cli/internal/output"
	skillpkg "github.com/MinimalFuture/linkai-cli/internal/skill"
)

func NewCmdSkill(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Install or read the agent skill bundled with the CLI",
		Long: "The linkai skill (SKILL.md + reference docs) is embedded in the CLI binary so " +
			"it always matches the CLI version. `skill install` copies it into the skill " +
			"directories that agents (Claude Code, Cursor, Codex, …) scan, so they can drive " +
			"the CLI out of the box.",
	}
	cmd.AddCommand(newInstallCmd(f), newListCmd(f), newReadCmd(f))
	return cmd
}

func newInstallCmd(f *cmdutil.Factory) *cobra.Command {
	var dir string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the embedded skill into agent skill directories",
		Long: "Copies the embedded skill into every detected agent skill directory under your " +
			"home (~/.claude/skills, ~/.cursor/skills, ~/.codex/skills, …), or into an explicit " +
			"path given with --dir. Existing copies are replaced so the skill stays in sync " +
			"with this CLI version.",
		Example: `  linkai skill install                  # into all detected agent homes
  linkai skill install --dir ./skills   # into a specific directory`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				results []skillpkg.InstallResult
				err     error
			)
			if dir != "" {
				results, err = skillpkg.InstallToDir(dir)
			} else {
				results, err = skillpkg.Install("")
			}
			if err != nil {
				return output.Errorf("skill install failed: %v", err)
			}

			if asJSON {
				return output.PrintJSON(f.IOStreams.Out, map[string]interface{}{
					"ok":        true,
					"installed": results,
					"count":     len(results),
				})
			}

			if len(results) == 0 {
				fmt.Fprintln(f.IOStreams.ErrOut, "No agent skill directories detected; nothing installed. Use --dir to target one explicitly.")
				return nil
			}
			for _, r := range results {
				output.PrintSuccess(f.IOStreams.Out, f.IOStreams.IsTerminal, r.Dest)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "install into this directory instead of auto-detected agent homes")
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List skills embedded in the CLI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			skills, err := skillpkg.List()
			if err != nil {
				return output.Errorf("%v", err)
			}
			if asJSON {
				return output.PrintJSON(f.IOStreams.Out, map[string]interface{}{
					"ok":     true,
					"skills": skills,
					"count":  len(skills),
				})
			}
			rows := make([][]string, 0, len(skills))
			for _, s := range skills {
				rows = append(rows, []string{s.Name, s.Description})
			}
			output.PrintTable(f.IOStreams.Out, []string{"NAME", "DESCRIPTION"}, rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}

func newReadCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <name> [path]",
		Short: "Print a skill's SKILL.md, or a reference file under it",
		Example: `  linkai skill read linkai-cli                       # the skill's SKILL.md
  linkai skill read linkai-cli references/chat.md    # a reference file`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			relpath := ""
			if len(args) == 2 {
				relpath = args[1]
			}
			content, _, err := skillpkg.Read(name, relpath)
			if err != nil {
				return output.Errorf("%v", err)
			}
			_, err = f.IOStreams.Out.Write(content)
			return err
		},
	}
	return cmd
}
