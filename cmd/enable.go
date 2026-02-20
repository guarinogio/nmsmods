package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable <id-or-index>",
	Short: "Enable a mod in the active profile (deploys it to the game MODS directory)",
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
			pi, ok := me.Installations[profile]
			if !ok || !pi.Installed || pi.Folder == "" {
				return fmt.Errorf("mod %s is not installed in profile %q", id, profile)
			}
			if pi.Enabled {
				fmt.Println("Already enabled:", id)
				return nil
			}

			storeAbs := filepath.Join(p.Root, filepath.FromSlash(pi.Store))
			if _, err := os.Stat(storeAbs); err != nil {
				return fmt.Errorf("stored mod folder not found: %s", storeAbs)
			}

			deployed, err := mods.Deploy(storeAbs, game.ModsDir, pi.Folder, id, profile)
			if err != nil {
				return err
			}

			pi.Enabled = true
			pi.DeployedPath = deployed
			me.Installations[profile] = pi

			// Legacy best-effort.
			me.Installed = true
			me.InstalledPath = deployed
			me.InstalledAt = pi.InstalledAt
			me.Folder = pi.Folder

			st.Mods[id] = me
			if err := app.SaveState(p.State, st); err != nil {
				return err
			}

			fmt.Println("Enabled:", id)
			fmt.Println("Deployed to:", deployed)
			return nil
		})
	},
}
