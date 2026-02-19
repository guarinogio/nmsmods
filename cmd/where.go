
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var whereCmd = &cobra.Command{
	Use:   "where",
	Short: "Show detected/configured No Man's Sky path",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		cfg, err := loadConfigAndMaybeGuess(p)
		if err != nil {
			return err
		}
		if cfg.GamePath == "" {
			fmt.Println("(not set)")
			fmt.Println("Use: nmsmods set-path <path>")
			return nil
		}
		fmt.Println(cfg.GamePath)
		return nil
	},
}
