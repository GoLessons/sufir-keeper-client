package cli

import (
	"github.com/spf13/cobra"
)

func AttachCompletion(root *cobra.Command) {
	c := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Сгенерировать скрипт автодополнения для оболочки",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return root.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			default:
				_, _ = cmd.OutOrStdout().Write([]byte("поддерживаются: bash, zsh, fish, powershell\n"))
				return nil
			}
		},
	}
	root.AddCommand(c)
}
