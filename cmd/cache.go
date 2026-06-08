package cmd

import (
	"fmt"
	"os"

	"github.com/ordinlabs/ssh-autocomplete/internal/sshparser"
	"github.com/spf13/cobra"
)

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the host name cache",
	}

	cmd.AddCommand(newCacheClearCmd())

	return cmd
}

func newCacheClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear the cached host names",
		Long:  "Removes the cached host list so the next completion will re-parse your SSH config.",
		RunE: func(cmd *cobra.Command, args []string) error {
			parser := sshparser.NewHostParser()
			cachePath := parser.CacheFilePath()

			if _, err := os.Stat(cachePath); os.IsNotExist(err) {
				fmt.Fprintln(cmd.OutOrStdout(), "No cache file found.")
				return nil
			}

			if err := os.Remove(cachePath); err != nil {
				return fmt.Errorf("failed to remove cache: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Cache cleared.")
			return nil
		},
	}
}
