package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var nexusLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save Nexus API key/token to config",
	Long:  "Save Nexus API key/token to config (stored in config.json with 0600 permissions).",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, cfg, err := nexusPathsConfig()
		if err != nil {
			return err
		}

		// If user provided via flag, use it. Otherwise prompt on stdin.
		key, _ := cmd.Flags().GetString("api-key")
		key = strings.TrimSpace(key)
		if key == "" {
			fmt.Fprint(cmd.ErrOrStderr(), "Enter Nexus API key: ")
			r := bufio.NewReader(os.Stdin)
			line, err := r.ReadString('\n')
			if err != nil {
				line = strings.TrimSpace(line)
				if line == "" {
					return fmt.Errorf("failed to read api key from stdin: %w", err)
				}
				key = line
			} else {
				key = strings.TrimSpace(line)
			}
		}

		if key == "" {
			return fmt.Errorf("empty api key")
		}

		cfg.Nexus.APIKey = key
		if err := app.SaveConfig(p.Config, cfg); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Saved Nexus API key to: %s\n", p.Config)
		return nil
	},
}

func init() {
	nexusLoginCmd.Flags().String("api-key", "", "Nexus API key/token (if omitted, you will be prompted)")
	nexusCmd.AddCommand(nexusLoginCmd)
}
