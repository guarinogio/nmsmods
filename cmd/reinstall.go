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
	Short: "Reinstall a mod in the active profile (replaces the stored folder and redeploys)",
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
				return fmt.Errorf("no zip recorded for %s", id)
			}
			zipAbs := joinPathFromState(p.Root, me.ZIP)

			folderGuess, err := mods.ProposedInstallFolderFromZip(zipAbs, id)
			if err != nil {
				return err
			}

			if me.Installations == nil {
				me.Installations = map[string]app.ProfileInstall{}
			}
			pi := me.Installations[profile]
			destFolder := pi.Folder
			if destFolder == "" {
				destFolder = folderGuess
			}
			destFolder, _ = mods.ResolveFolderCollision(id, destFolder, profile, st)

			storeAbs := filepath.Join(app.ProfileModsDir(p, profile), destFolder)
			deployAbs := filepath.Join(game.ModsDir, destFolder)

			if reinstallDryRun {
				fmt.Println("[dry-run] Would reinstall:")
				fmt.Println("  profile:", profile)
				fmt.Println("  id:     ", id)
				fmt.Println("  zip:    ", zipAbs)
				fmt.Println("  folder: ", destFolder)
				fmt.Println("  store:  ", storeAbs)
				fmt.Println("  deploy: ", deployAbs)
				fmt.Println("  action:  REPLACE store + REDEPLOY")
				return nil
			}

			if fileExists(storeAbs) {
				if reinstallNoOverwrite {
					return fmt.Errorf("store exists: %s (use without --no-overwrite to replace it)", storeAbs)
				}
				fmt.Println("Removing existing store:", storeAbs)
				_ = os.RemoveAll(storeAbs)
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
			folder, _ = mods.ResolveFolderCollision(id, folder, profile, st)
			storeAbs = filepath.Join(app.ProfileModsDir(p, profile), folder)

			fmt.Println("Installing into profile store:", storeAbs)
			if err := mods.CopyDir(srcPath, storeAbs); err != nil {
				return err
			}

			// Redeploy if enabled (default to enabled).
			enabled := true
			if pi.Installed {
				enabled = pi.Enabled
			}
			pi.Installed = true
			pi.Enabled = enabled
			pi.Folder = folder
			pi.Store = filepath.ToSlash(filepath.Join("profiles", profile, "mods", folder))
			pi.InstalledAt = app.NowRFC3339()

			if enabled {
				deployed, err := mods.Deploy(storeAbs, game.ModsDir, folder)
				if err != nil {
					return err
				}
				pi.DeployedPath = deployed
			}

			ok, verr := mods.HasRelevantFiles(storeAbs)
			if verr != nil || !ok {
				fmt.Println("Warning: installed folder contains no .EXML/.MBIN files (or health check failed)")
				me.Health = "warning"
			} else {
				me.Health = "ok"
			}

			me.Installations[profile] = pi
			// Legacy best-effort
			me.Folder = folder
			me.Installed = enabled
			me.InstalledAt = pi.InstalledAt
			me.InstalledPath = pi.DeployedPath

			st.Mods[id] = me
			if err := app.SaveState(p.State, st); err != nil {
				return err
			}

			fmt.Println("Reinstalled:", id, "(profile:", profile+")")
			return nil
		})
	},
}

func init() {
	reinstallCmd.Flags().BoolVar(&reinstallDryRun, "dry-run", false, "Print what would happen without making changes")
	reinstallCmd.Flags().BoolVar(&reinstallNoOverwrite, "no-overwrite", false, "Do not overwrite if destination in the profile store already exists")
}
