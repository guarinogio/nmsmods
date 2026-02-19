package cmd

import (
	"encoding/json"
	"fmt"

	"nmsmods/internal/nms"

	"github.com/spf13/cobra"
)

var installedJSON bool

type installedRow struct {
	Folder string `json:"folder"`
}

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

		if installedJSON {
			out := make([]installedRow, 0, len(modsList))
			for _, m := range modsList {
				out = append(out, installedRow{Folder: m})
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
			return nil
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

func init() {
	installedCmd.Flags().BoolVar(&installedJSON, "json", false, "Output in JSON format")
}
