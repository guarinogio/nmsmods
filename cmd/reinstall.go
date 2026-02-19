package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var reinstallDryRun bool
var reinstallNoOverwrite bool

var reinstallCmd = &cobra.Command{
	Use:   "reinstall <id-or-index>",
	Short: "Uninstall (if present) and install again (overwrites by default)",
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
				return fmt.Errorf("no zip recorded for %s", id)
			}
			zipAbs := joinPathFromState(p.Root, me.ZIP)

			folderGuess, err := mods.ProposedInstallFolderFromZip(zipAbs, id)
			if err != nil {
				return err
			}

			destFolder := me.Folder
			if destFolder == "" {
				destFolder = folderGuess
			}
			dest := filepath.Join(game.ModsDir, destFolder)

			if reinstallDryRun {
				fmt.Println("[dry-run] Would reinstall:")
				fmt.Println("  id:      ", id)
				fmt.Println("  zip:     ", zipAbs)
				fmt.Println("  folder:  ", destFolder)
				fmt.Println("  dest:    ", dest)
				fmt.Println("  action:  REMOVE dest (if exists) + INSTALL")
				return nil
			}

			// Remove existing folder if present
			if _, err := os.Stat(dest); err == nil {
				if reinstallNoOverwrite {
					return fmt.Errorf("destination exists: %s (use without --no-overwrite to replace it)", dest)
				}
				fmt.Println("Removing existing:", dest)
				if err := os.RemoveAll(dest); err != nil {
					return err
				}
			}

			// Install
			stageDir := filepath.Join(p.Staging, id)
			_ = os.RemoveAll(stageDir)
			if err := os.MkdirAll(stageDir, 0o755); err != nil {
				return err
			}

			fmt.Println("Extracting to:", stageDir)
			if err := mods.ExtractZip(zipAbs, stageDir); err != nil {
				return err
			}

			folder, srcPath, err := mods.ChooseInstallFolder(stageDir, id)
			if err != nil {
				return err
			}
			dest = filepath.Join(game.ModsDir, folder)

			if _, err := os.Stat(dest); err == nil {
				if reinstallNoOverwrite {
					return fmt.Errorf("destination exists: %s (use without --no-overwrite to replace it)", dest)
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

			st.Mods[id] = me
			if err := app.SaveState(p.State, st); err != nil {
				return err
			}

			fmt.Println("Reinstalled to:", dest)
			return nil
		})
	},
}

func init() {
	reinstallCmd.Flags().BoolVar(&reinstallDryRun, "dry-run", false, "Print what would happen without making changes")
	reinstallCmd.Flags().BoolVar(&reinstallNoOverwrite, "no-overwrite", false, "Do not overwrite if destination folder already exists")
}
