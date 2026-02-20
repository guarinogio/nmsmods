package cmd

import (
	"fmt"
	"path/filepath"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable <id-or-index>",
	Short: "Disable a mod in the active profile (keeps it in the profile store, removes it from the game)",
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
			if !pi.Enabled {
				fmt.Println("Already disabled:", id)
				return nil
			}

			if err := mods.Undeploy(game.ModsDir, pi.Folder, id, profile); err != nil {
				return err
			}
			pi.Enabled = false
			pi.DeployedPath = ""
			me.Installations[profile] = pi

			// Legacy: mark not installed in game, but keep store in profile.
			me.Installed = false
			me.InstalledPath = ""

			st.Mods[id] = me
			if err := app.SaveState(p.State, st); err != nil {
				return err
			}

			fmt.Println("Disabled:", id)
			fmt.Println("Store:", filepath.Join(p.Root, filepath.FromSlash(pi.Store)))
			return nil
		})
	},
}
