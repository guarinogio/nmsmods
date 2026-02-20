package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var dryRunUninstall bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <id-or-index-or-folder>",
	Short: "Uninstall a mod from the active profile (and undeploy from the game)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		target := args[0]

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

			treatedAsTracked := false
			var trackedID string

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
				pi, ok := me.Installations[profile]
				if !ok || !pi.Installed || pi.Folder == "" {
					return fmt.Errorf("mod %s is not installed in profile %q", trackedID, profile)
				}

				storeAbs := filepath.Join(p.Root, filepath.FromSlash(pi.Store))
				deployDest := filepath.Join(game.ModsDir, pi.Folder)

				if dryRunUninstall {
					fmt.Println("[dry-run] Would uninstall from profile:")
					fmt.Println("  profile:", profile)
					fmt.Println("  id:     ", trackedID)
					fmt.Println("  folder: ", pi.Folder)
					fmt.Println("  store:  ", storeAbs)
					fmt.Println("  deploy: ", deployDest)
					fmt.Println("  action:  REMOVE store + undeploy; set installed=false in state.json for this profile")
					return nil
				}

				// Undeploy first (so game state is clean even if store removal fails).
				if err := mods.Undeploy(game.ModsDir, pi.Folder, trackedID, profile); err != nil {
					return err
				}

				if pi.Store != "" {
					if _, err := os.Stat(storeAbs); err == nil {
						if err := os.RemoveAll(storeAbs); err != nil {
							return err
						}
					}
				}

				pi.Installed = false
				pi.Enabled = false
				pi.DeployedPath = ""
				pi.InstalledAt = ""
				me.Installations[profile] = pi

				// Legacy fields best-effort.
				me.Installed = false
				me.InstalledAt = ""
				me.InstalledPath = ""
				me.Health = ""

				st.Mods[trackedID] = me
				if err := app.SaveState(p.State, st); err != nil {
					return err
				}

				fmt.Println("Uninstalled:", trackedID, "(profile:", profile+")")
				return nil
			}

			// Otherwise treat as folder name directly (DANGEROUS - not tracked).
			// Still validate it's a single segment to avoid accidental traversal.
			folder, err := mods.SanitizeFolderName(target, target)
			if err != nil {
				return err
			}
			dest, err := mods.SafeJoinUnder(game.ModsDir, folder)
			if err != nil {
				return err
			}

			if dryRunUninstall {
				fmt.Println("[dry-run] Would uninstall by folder (untracked):")
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
			fmt.Println("Removed (untracked):", dest)
			return nil
		})
	},
}

func init() {
	uninstallCmd.Flags().BoolVar(&dryRunUninstall, "dry-run", false, "Print what would happen without making changes")
}
