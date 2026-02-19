
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"nmsmods/internal/app"
	"nmsmods/internal/nms"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run basic checks (paths, downloads, installed mods)",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		cfg, err := loadConfigAndMaybeGuess(p)
		if err != nil {
			return err
		}
		fmt.Println("Data dir:", p.Root)
		fmt.Println("Downloads:", p.Downloads)
		fmt.Println("Staging:", p.Staging)

		if cfg.GamePath == "" {
			fmt.Println("Game path: (not set)")
			fmt.Println("Tip: nmsmods set-path <path>")
		} else {
			fmt.Println("Game path:", cfg.GamePath)
			game, gerr := nms.ValidateGamePath(cfg.GamePath)
			if gerr != nil {
				fmt.Println("Game path validation: ERROR:", gerr)
			} else {
				_ = nms.EnsureModsDir(game)
				fmt.Println("Mods dir:", game.ModsDir)
				installed, _ := nms.ListInstalledModFolders(game)
				fmt.Printf("Installed mod folders: %d\n", len(installed))
				for _, m := range installed {
					fmt.Println(" -", m)
				}
			}
		}

		// downloads from state
		st, err := app.LoadState(p.State)
		if err != nil {
			return err
		}
		fmt.Printf("Tracked downloads: %d\n", len(st.Mods))
		for id, me := range st.Mods {
			zipAbs := filepath.Join(p.Root, filepath.FromSlash(me.ZIP))
			zipStatus := "missing"
			if me.ZIP != "" {
				if _, err := os.Stat(zipAbs); err == nil {
					zipStatus = "present"
				}
			} else {
				zipStatus = "(none)"
			}
			fmt.Printf(" - %s | installed=%v | zip=%s\n", id, me.Installed, zipStatus)
		}
		return nil
	},
}
