package cmd

import (
	"fmt"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var nexusLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored Nexus API key/token from config",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, cfg, err := nexusPathsConfig()
		if err != nil {
			return err
		}
		if cfg.Nexus.APIKey == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "No Nexus API key stored.")
			return nil
		}
		cfg.Nexus.APIKey = ""
		if err := app.SaveConfig(p.Config, cfg); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Removed Nexus API key from config.")
		return nil
	},
}

func init() {
	nexusCmd.AddCommand(nexusLogoutCmd)
}
