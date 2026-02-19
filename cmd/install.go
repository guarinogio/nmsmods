package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var forceInstall bool

var installCmd = &cobra.Command{
	Use:   "install <id-or-index>",
	Short: "Install a downloaded mod into <NMS>/GAMEDATA/MODS",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()

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
			return fmt.Errorf("no zip recorded for %s. Use: nmsmods download <url> [--id %s]", id, id)
		}
		zipAbs := joinPathFromState(p.Root, me.ZIP)
		if _, err := os.Stat(zipAbs); err != nil {
			return fmt.Errorf("zip not found: %s", zipAbs)
		}

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
		dest := filepath.Join(game.ModsDir, folder)

		if _, err := os.Stat(dest); err == nil {
			if !forceInstall {
				return fmt.Errorf("destination exists: %s (use --force to overwrite)", dest)
			}
			fmt.Println("Overwriting:", dest)
			_ = os.RemoveAll(dest)
		}

		fmt.Println("Installing folder:", folder)
		if err := mods.CopyDir(srcPath, dest); err != nil {
			return err
		}

		me.Folder = folder
		me.Installed = true
		st.Mods[id] = me
		if err := app.SaveState(p.State, st); err != nil {
			return err
		}

		fmt.Println("Installed to:", dest)
		return nil
	},
}

func init() {
	installCmd.Flags().BoolVar(&forceInstall, "force", false, "Overwrite if destination folder already exists")
}
