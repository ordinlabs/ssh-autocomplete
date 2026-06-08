package cmd

import (
	"fmt"

	"github.com/ordinlabs/ssh-autocomplete/internal/sshparser"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var noCache bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SSH host names",
		Long:  "Parses SSH configuration files and outputs non-wildcard host names, one per line. Results are cached for 10 seconds unless --no-cache is specified.",
		RunE: func(cmd *cobra.Command, args []string) error {
			parser := sshparser.NewHostParser()

			useCache := !noCache
			hosts, err := parser.GetHosts(useCache)
			if err != nil {
				return fmt.Errorf("failed to parse SSH hosts: %w", err)
			}

			for _, host := range hosts {
				fmt.Fprintln(cmd.OutOrStdout(), host)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "bypass cache and regenerate host names")

	return cmd
}
