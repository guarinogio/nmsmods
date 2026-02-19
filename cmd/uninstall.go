package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var dryRunUninstall bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <id-or-index-or-folder>",
	Short: "Remove a mod folder from <NMS>/GAMEDATA/MODS (tracked by id/index if possible)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		target := args[0]

		_, game, err := requireGame(p)
		if err != nil {
			return err
		}

		st, err := app.LoadState(p.State)
		if err != nil {
			return err
		}

		// Try interpret as tracked id/index first.
		treatedAsTracked := false
		var trackedID string
		var folder string

		if _, aerr := strconv.Atoi(target); aerr == nil {
			if id, rerr := resolveModArg(target, st); rerr == nil {
				treatedAsTracked = true
				trackedID = id
			}
		} else {
			if _, ok := st.Mods[target]; ok {
				treatedAsTracked = true
				trackedID = target
			}
		}

		if treatedAsTracked {
			me := st.Mods[trackedID]
			if me.Folder == "" {
				return fmt.Errorf("tracked mod %s has no installed folder recorded", trackedID)
			}
			folder = me.Folder
			dest := filepath.Join(game.ModsDir, folder)

			if dryRunUninstall {
				fmt.Println("[dry-run] Would uninstall tracked mod:")
				fmt.Println("  id:     ", trackedID)
				fmt.Println("  folder: ", folder)
				fmt.Println("  dest:   ", dest)
				fmt.Println("  action:  REMOVE folder; set installed=false in state.json")
				return nil
			}

			if _, err := os.Stat(dest); os.IsNotExist(err) {
				return fmt.Errorf("not found: %s", dest)
			}
			if err := os.RemoveAll(dest); err != nil {
				return err
			}

			me.Installed = false
			st.Mods[trackedID] = me
			_ = app.SaveState(p.State, st)

			fmt.Println("Removed:", dest)
			return nil
		}

		// Otherwise treat as folder name directly.
		dest := filepath.Join(game.ModsDir, target)

		if dryRunUninstall {
			fmt.Println("[dry-run] Would uninstall by folder:")
			fmt.Println("  folder: ", target)
			fmt.Println("  dest:   ", dest)
			fmt.Println("  action:  REMOVE folder")
			return nil
		}

		if _, err := os.Stat(dest); os.IsNotExist(err) {
			return fmt.Errorf("not found: %s", dest)
		}
		if err := os.RemoveAll(dest); err != nil {
			return err
		}
		fmt.Println("Removed:", dest)
		return nil
	},
}

func init() {
	uninstallCmd.Flags().BoolVar(&dryRunUninstall, "dry-run", false, "Print what would happen without making changes")
}
