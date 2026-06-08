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

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSetupCmd())
	cmd.AddCommand(newUninstallCmd())
	cmd.AddCommand(newCacheCmd())

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}
