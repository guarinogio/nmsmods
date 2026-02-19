package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

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

		// If numeric index or known id, uninstall tracked mod (folder comes from state).
		if _, aerr := strconv.Atoi(target); aerr == nil || st.Mods[target].URL != "" || st.Mods[target].ZIP != "" || st.Mods[target].Folder != "" {
			id, rerr := resolveModArg(target, st)
			if rerr == nil {
				me := st.Mods[id]
				if me.Folder == "" {
					return fmt.Errorf("tracked mod %s has no installed folder recorded", id)
				}
				dest := filepath.Join(game.ModsDir, me.Folder)
				if _, err := os.Stat(dest); os.IsNotExist(err) {
					return fmt.Errorf("not found: %s", dest)
				}
				if err := os.RemoveAll(dest); err != nil {
					return err
				}
				me.Installed = false
				st.Mods[id] = me
				_ = app.SaveState(p.State, st)
				fmt.Println("Removed:", dest)
				return nil
			}
			// if resolve failed, fallthrough to folder uninstall
		}

		// Otherwise interpret as folder name directly.
		dest := filepath.Join(game.ModsDir, target)
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
