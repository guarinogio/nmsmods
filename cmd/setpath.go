
package cmd

import (
	"fmt"

	"nmsmods/internal/app"
	"nmsmods/internal/nms"

	"github.com/spf13/cobra"
)

var setPathCmd = &cobra.Command{
	Use:   "set-path <path>",
	Short: "Set No Man's Sky installation path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		path := args[0]
		if _, err := nms.ValidateGamePath(path); err != nil {
			return err
		}
		cfg := app.Config{GamePath: path}
		if err := app.SaveConfig(p.Config, cfg); err != nil {
			return err
		}
		fmt.Println("Saved game path:", path)
		return nil
	},
}
