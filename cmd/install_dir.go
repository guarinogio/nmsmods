package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var installDirID string
var installDirNoOverwrite bool
var installDirDryRun bool

var installDirCmd = &cobra.Command{
	Use:   "install-dir <path>",
	Short: "Install a local extracted mod directory into <NMS>/GAMEDATA/MODS (overwrites by default)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		src := args[0]

		return withStateLock(p, func() error {
			_, game, err := requireGame(p)
			if err != nil {
				return err
			}

			fi, err := os.Stat(src)
			if err != nil {
				return err
			}
			if !fi.IsDir() {
				return fmt.Errorf("not a directory: %s", src)
			}

			id := installDirID
			if id == "" {
				id = mods.SlugFromURL(filepath.Base(src))
			}

			state, err := app.LoadState(p.State)
			if err != nil {
				return err
			}

			folder, copyFrom, err := mods.ChooseInstallFolder(src, id)
			if err != nil {
				return err
			}
			folder, collided := mods.ResolveFolderCollision(id, folder, state)
			dest := filepath.Join(game.ModsDir, folder)

			if installDirDryRun {
				fmt.Println("[dry-run] Would install directory:")
				fmt.Println("  id:     ", id)
				fmt.Println("  src:    ", src)
				fmt.Println("  folder: ", folder)
				fmt.Println("  from:   ", copyFrom)
				fmt.Println("  dest:   ", dest)
				if collided {
					fmt.Println("  note:    collision avoided (another mod uses same folder)")
				}
				if _, err := os.Stat(dest); err == nil {
					if installDirNoOverwrite {
						fmt.Println("  action:  SKIP (destination exists and --no-overwrite set)")
					} else {
						fmt.Println("  action:  REPLACE (destination exists; overwrite is default)")
					}
				} else {
					fmt.Println("  action:  INSTALL")
				}
				return nil
			}

			if _, err := os.Stat(dest); err == nil {
				if installDirNoOverwrite {
					return fmt.Errorf("destination exists: %s (run without --no-overwrite to replace it)", dest)
				}
				fmt.Println("Replacing existing install:", dest)
				if err := os.RemoveAll(dest); err != nil {
					return err
				}
			}

			fmt.Println("Installing folder:", folder)
			if err := mods.CopyDir(copyFrom, dest); err != nil {
				return err
			}

			me := state.Mods[id]
			me.Folder = folder
			me.Installed = true
			me.InstalledPath = dest
			me.InstalledAt = app.NowRFC3339()
			if me.URL == "" {
				me.URL = "dir://" + src
			}
			if me.Source == "" {
				me.Source = "local"
			}
			if me.DisplayName == "" {
				me.DisplayName = id
			}

			ok, verr := mods.HasRelevantFiles(dest)
			if verr != nil || !ok {
				fmt.Println("Warning: installed folder contains no .EXML/.MBIN files (or health check failed)")
				me.Health = "warning"
			} else {
				me.Health = "ok"
			}

			state.Mods[id] = me
			if err := app.SaveState(p.State, state); err != nil {
				return err
			}

			fmt.Println("Installed to:", dest)
			return nil
		})
	},
}

func init() {
	installDirCmd.Flags().StringVar(&installDirID, "id", "", "Override mod id (slug)")
	installDirCmd.Flags().BoolVar(&installDirNoOverwrite, "no-overwrite", false, "Do not overwrite if destination folder already exists")
	installDirCmd.Flags().BoolVar(&installDirDryRun, "dry-run", false, "Print what would happen without making changes")
}
