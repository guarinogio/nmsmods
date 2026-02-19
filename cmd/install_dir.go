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
	Short: "Install an extracted mod directory into the active profile, then deploy to the game",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		src := args[0]

		return withStateLock(p, func() error {
			cfg, game, err := requireGame(p)
			if err != nil {
				return err
			}
			profile, err := ensureActiveProfileDirs(p, cfg)
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
			folder, collided := mods.ResolveFolderCollision(id, folder, profile, state)

			storeDir := app.ProfileModsDir(p, profile)
			storePath := filepath.Join(storeDir, folder)
			deployPath := filepath.Join(game.ModsDir, folder)

			if installDirDryRun {
				fmt.Println("[dry-run] Would install directory:")
				fmt.Println("  profile:", profile)
				fmt.Println("  id:     ", id)
				fmt.Println("  src:    ", src)
				fmt.Println("  folder: ", folder)
				fmt.Println("  from:   ", copyFrom)
				fmt.Println("  store:  ", storePath)
				fmt.Println("  deploy: ", deployPath)
				if collided {
					fmt.Println("  note:    collision avoided (another mod uses same folder in this profile)")
				}
				if fileExists(storePath) {
					if installDirNoOverwrite {
						fmt.Println("  action:  SKIP (store exists and --no-overwrite set)")
					} else {
						fmt.Println("  action:  REPLACE (store exists; overwrite is default)")
					}
				} else {
					fmt.Println("  action:  INSTALL")
				}
				return nil
			}

			if fileExists(storePath) {
				if installDirNoOverwrite {
					return fmt.Errorf("destination exists in profile store: %s (run without --no-overwrite to replace it)", storePath)
				}
				fmt.Println("Replacing existing profile install:", storePath)
				if err := os.RemoveAll(storePath); err != nil {
					return err
				}
			}

			fmt.Println("Installing into profile store:", storePath)
			if err := mods.CopyDir(copyFrom, storePath); err != nil {
				return err
			}

			deployed, err := mods.Deploy(storePath, game.ModsDir, folder)
			if err != nil {
				return err
			}

			me := state.Mods[id]
			if me.Installations == nil {
				me.Installations = map[string]app.ProfileInstall{}
			}
			pi := me.Installations[profile]
			pi.Installed = true
			pi.Enabled = true
			pi.Folder = folder
			pi.Store = filepath.ToSlash(filepath.Join("profiles", profile, "mods", folder))
			pi.DeployedPath = deployed
			pi.InstalledAt = app.NowRFC3339()
			me.Installations[profile] = pi

			// Backfill metadata
			if me.URL == "" {
				me.URL = "dir://" + src
			}
			if me.Source == "" {
				me.Source = "local"
			}
			if me.DisplayName == "" {
				me.DisplayName = id
			}

			ok, verr := mods.HasRelevantFiles(storePath)
			if verr != nil || !ok {
				fmt.Println("Warning: installed folder contains no .EXML/.MBIN files (or health check failed)")
				me.Health = "warning"
			} else {
				me.Health = "ok"
			}

			// Legacy best-effort.
			me.Folder = folder
			me.Installed = true
			me.InstalledPath = deployed
			me.InstalledAt = pi.InstalledAt

			state.Mods[id] = me
			if err := app.SaveState(p.State, state); err != nil {
				return err
			}

			fmt.Println("Installed in profile:", profile)
			fmt.Println("Deployed to:", deployed)
			return nil
		})
	},
}

func init() {
	installDirCmd.Flags().StringVar(&installDirID, "id", "", "Override mod id (slug)")
	installDirCmd.Flags().BoolVar(&installDirNoOverwrite, "no-overwrite", false, "Do not overwrite if destination already exists in the profile store")
	installDirCmd.Flags().BoolVar(&installDirDryRun, "dry-run", false, "Print what would happen without making changes")
}
