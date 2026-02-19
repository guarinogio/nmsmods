package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var noOverwrite bool
var dryRunInstall bool

var installCmd = &cobra.Command{
	Use:   "install <id-or-index>",
	Short: "Install a downloaded mod into the active profile, then deploy to <NMS>/GAMEDATA/MODS",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()

		return withStateLock(p, func() error {
			cfg, game, err := requireGame(p)
			if err != nil {
				return err
			}
			profile, err := ensureActiveProfileDirs(p, cfg)
			if err != nil {
				return err
			}

			st, err := app.LoadState(p.State)
			if err != nil {
				return err
			}

			id, err := resolveModArg(args[0], st)
			if err != nil {
				return err
			}

			me := st.Mods[id]
			if me.ZIP == "" {
				return fmt.Errorf("no zip recorded for %s. Use: nmsmods download <url-or-zip> [--id %s]", id, id)
			}
			zipAbs := joinPathFromState(p.Root, me.ZIP)
			if _, err := os.Stat(zipAbs); err != nil {
				return fmt.Errorf("zip not found: %s", zipAbs)
			}

			// Predict folder name without extracting (useful for --dry-run).
			folder, err := mods.ProposedInstallFolderFromZip(zipAbs, id)
			if err != nil {
				return err
			}

			if me.Installations == nil {
				me.Installations = map[string]app.ProfileInstall{}
			}
			pi := me.Installations[profile]

			// Avoid clobbering another mod folder within this profile.
			folder, collided := mods.ResolveFolderCollision(id, folder, profile, st)

			storeDir := app.ProfileModsDir(p, profile)
			storePath := filepath.Join(storeDir, folder)
			storeExists := fileExists(storePath)

			if dryRunInstall {
				fmt.Println("[dry-run] Would install to profile:")
				fmt.Println("  profile:", profile)
				fmt.Println("  id:     ", id)
				fmt.Println("  zip:    ", zipAbs)
				fmt.Println("  folder: ", folder)
				fmt.Println("  store:  ", storePath)
				fmt.Println("  deploy: ", filepath.Join(game.ModsDir, folder))
				if collided {
					fmt.Println("  note:    collision avoided (another mod uses same folder in this profile)")
				}
				if storeExists {
					if noOverwrite {
						fmt.Println("  action:  SKIP (store exists and --no-overwrite set)")
					} else {
						fmt.Println("  action:  REPLACE (store exists; overwrite is default)")
					}
				} else {
					fmt.Println("  action:  INSTALL")
				}
				return nil
			}

			// Extract -> choose folder -> copy into profile store -> deploy into game MODS
			stageDir := filepath.Join(p.Staging, id)
			_ = os.RemoveAll(stageDir)
			if err := os.MkdirAll(stageDir, 0o755); err != nil {
				return err
			}

			fmt.Println("Extracting to:", stageDir)
			if err := mods.ExtractZip(zipAbs, stageDir); err != nil {
				return err
			}

			// Choose folder based on extracted layout (authoritative)
			folder, srcPath, err := mods.ChooseInstallFolder(stageDir, id)
			if err != nil {
				return err
			}
			folder, _ = mods.ResolveFolderCollision(id, folder, profile, st)

			storePath = filepath.Join(storeDir, folder)

			if fileExists(storePath) {
				if noOverwrite {
					return fmt.Errorf("destination exists in profile store: %s (run without --no-overwrite to replace it)", storePath)
				}
				fmt.Println("Replacing existing profile install:", storePath)
				if err := os.RemoveAll(storePath); err != nil {
					return err
				}
			}

			fmt.Println("Installing into profile store:", storePath)
			if err := mods.CopyDir(srcPath, storePath); err != nil {
				return err
			}

			ok, verr := mods.HasRelevantFiles(storePath)
			if verr != nil || !ok {
				fmt.Println("Warning: installed folder contains no .EXML/.MBIN files (or health check failed)")
				me.Health = "warning"
			} else {
				me.Health = "ok"
			}

			// Enabled by default: deploy to game
			deployed, err := mods.Deploy(storePath, game.ModsDir, folder)
			if err != nil {
				return err
			}

			pi.Installed = true
			pi.Enabled = true
			pi.Folder = folder
			pi.Store = filepath.ToSlash(filepath.Join("profiles", profile, "mods", folder))
			pi.DeployedPath = deployed
			pi.InstalledAt = app.NowRFC3339()
			me.Installations[profile] = pi

			// Keep legacy fields in sync for users that rely on them (best effort).
			me.Folder = folder
			me.Installed = true
			me.InstalledPath = deployed
			me.InstalledAt = pi.InstalledAt

			if me.DisplayName == "" {
				me.DisplayName = id
			}

			st.Mods[id] = me
			if err := app.SaveState(p.State, st); err != nil {
				return err
			}

			fmt.Println("Installed in profile:", profile)
			fmt.Println("Deployed to:", deployed)
			return nil
		})
	},
}

func init() {
	installCmd.Flags().BoolVar(&noOverwrite, "no-overwrite", false, "Do not overwrite if destination already exists in the profile store")
	installCmd.Flags().BoolVar(&dryRunInstall, "dry-run", false, "Print what would happen without making changes")
}
