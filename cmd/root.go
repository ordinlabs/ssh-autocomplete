package cmd

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh-autocomplete",
		Short: "SSH host autocompletion helper",
		Long:  "Parses SSH configuration files and provides host name autocompletion for bash, zsh, and PowerShell.",
	}

	cmd.AddCommand(newGenerateCmd())
	cmd.AddCommand(newSetupCmd())
	cmd.AddCommand(newUninstallCmd())

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}
