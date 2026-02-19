package cmd

import (
	"fmt"

	"nmsmods/internal/app"
	"nmsmods/internal/nms"

	"github.com/spf13/cobra"
)

var setPathCmd = &cobra.Command{
	Use:   "set-path <path>",
	Short: "Set the No Man's Sky installation path (Steam library folder)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		newPath := args[0]

		return withStateLock(p, func() error {
			// Validate now so config doesn't get junk
			if _, err := nms.ValidateGamePath(newPath); err != nil {
				return fmt.Errorf("invalid game path: %w", err)
			}

			cfg, err := app.LoadConfig(p.Config)
			if err != nil {
				return err
			}
			cfg.GamePath = newPath
			if err := app.SaveConfig(p.Config, cfg); err != nil {
				return err
			}

			fmt.Println("Game path set to:", newPath)
			return nil
		})
	},
}
