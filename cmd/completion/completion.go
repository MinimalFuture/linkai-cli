package completion

import (
	"os"

	"github.com/spf13/cobra"
)

// NewCmdCompletion creates the completion command for shell autocompletion.
func NewCmdCompletion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell autocompletion script",
		Long: `Generate shell autocompletion script for linkai.

To load completions:

Bash:
  $ source <(linkai completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ linkai completion bash > /etc/bash_completion.d/linkai
  # macOS:
  $ linkai completion bash > $(brew --prefix)/etc/bash_completion.d/linkai

Zsh:
  $ source <(linkai completion zsh)
  # To load completions for each session, execute once:
  $ linkai completion zsh > "${fpath[1]}/_linkai"

Fish:
  $ linkai completion fish | source
  # To load completions for each session, execute once:
  $ linkai completion fish > ~/.config/fish/completions/linkai.fish

PowerShell:
  PS> linkai completion powershell | Out-String | Invoke-Expression
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}
