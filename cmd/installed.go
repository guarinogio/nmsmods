
package cmd

import (
	"fmt"

	"nmsmods/internal/nms"

	"github.com/spf13/cobra"
)

var installedCmd = &cobra.Command{
	Use:   "installed",
	Short: "List installed mod folders under <NMS>/GAMEDATA/MODS",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		_, game, err := requireGame(p)
		if err != nil {
			return err
		}
		modsList, err := nms.ListInstalledModFolders(game)
		if err != nil {
			return err
		}
		if len(modsList) == 0 {
			fmt.Println("(none)")
			return nil
		}
		for _, m := range modsList {
			fmt.Println(m)
		}
		return nil
	},
}
