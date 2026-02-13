package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for zeroui.

The completion script can be loaded to provide auto-completion for zeroui commands,
subcommands, and flags in your shell.

Supported shells:
  - bash
  - zsh
  - fish

To load completions:

Bash:

  $ source <(zeroui completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ zeroui completion bash > /etc/bash_completion.d/zeroui

  # macOS:
  $ zeroui completion bash > $(brew --prefix)/etc/bash_completion.d/zeroui

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ zeroui completion zsh > "${fpath[1]}/_zeroui"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ zeroui completion fish | source

  # To load completions for each session, execute once:
  $ zeroui completion fish > ~/.config/fish/completions/zeroui.fish
`,
		Example: `  # Bash
  source <(zeroui completion bash)

  # Zsh
  zeroui completion zsh > "${fpath[1]}/_zeroui"

  # Fish
  zeroui completion fish | source`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]

			switch shell {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
			}
		},
	}
}
