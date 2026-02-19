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
	Short: "Install a downloaded mod into <NMS>/GAMEDATA/MODS (overwrites by default)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()

		return withStateLock(p, func() error {
			_, game, err := requireGame(p)
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

			// Avoid clobbering another installed mod folder.
			folder, collided := mods.ResolveFolderCollision(id, folder, st)
			dest := filepath.Join(game.ModsDir, folder)

			_, destErr := os.Stat(dest)
			destExists := destErr == nil

			if dryRunInstall {
				fmt.Println("[dry-run] Would install:")
				fmt.Println("  id:     ", id)
				fmt.Println("  zip:    ", zipAbs)
				fmt.Println("  folder: ", folder)
				fmt.Println("  dest:   ", dest)
				if collided {
					fmt.Println("  note:    collision avoided (another mod uses same folder)")
				}
				if destExists {
					if noOverwrite {
						fmt.Println("  action:  SKIP (destination exists and --no-overwrite set)")
					} else {
						fmt.Println("  action:  REPLACE (destination exists; overwrite is default)")
					}
				} else {
					fmt.Println("  action:  INSTALL")
				}
				return nil
			}

			// Real install path: extract -> choose folder -> copy
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
			folder, _ = mods.ResolveFolderCollision(id, folder, st)
			dest = filepath.Join(game.ModsDir, folder)

			if _, err := os.Stat(dest); err == nil {
				if noOverwrite {
					return fmt.Errorf("destination exists: %s (run without --no-overwrite to replace it)", dest)
				}
				fmt.Println("Replacing existing install:", dest)
				if err := os.RemoveAll(dest); err != nil {
					return err
				}
			}

			fmt.Println("Installing folder:", folder)
			if err := mods.CopyDir(srcPath, dest); err != nil {
				return err
			}

			me.Folder = folder
			me.Installed = true
			me.InstalledPath = dest
			me.InstalledAt = app.NowRFC3339()

			ok, verr := mods.HasRelevantFiles(dest)
			if verr != nil || !ok {
				fmt.Println("Warning: installed folder contains no .EXML/.MBIN files (or health check failed)")
				me.Health = "warning"
			} else {
				me.Health = "ok"
			}

			if me.DisplayName == "" {
				me.DisplayName = id
			}
			st.Mods[id] = me
			if err := app.SaveState(p.State, st); err != nil {
				return err
			}

			fmt.Println("Installed to:", dest)
			return nil
		})
	},
}

func init() {
	installCmd.Flags().BoolVar(&noOverwrite, "no-overwrite", false, "Do not overwrite if destination folder already exists")
	installCmd.Flags().BoolVar(&dryRunInstall, "dry-run", false, "Print what would happen without making changes")
}
