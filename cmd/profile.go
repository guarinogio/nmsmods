package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage mod profiles (separate stored sets of installed mods)",
}

var profileStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the active profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		cfg, err := loadConfig(p)
		if err != nil {
			return err
		}
		prof := app.ActiveProfile(cfg)
		fmt.Println("Active profile:", prof)
		fmt.Println("Profile store:", app.ProfileModsDir(p, prof))
		return nil
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List existing profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		cfg, err := loadConfig(p)
		if err != nil {
			return err
		}
		active := app.ActiveProfile(cfg)

		entries, err := os.ReadDir(p.Profiles)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		names := []string{"default"}
		for _, e := range entries {
			if e.IsDir() {
				n := e.Name()
				if n != "default" {
					names = append(names, n)
				}
			}
		}
		sort.Strings(names)
		for _, n := range names {
			mark := ""
			if n == active {
				mark = " *"
			}
			fmt.Println(n + mark)
		}
		return nil
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch active profile and deploy its enabled mods into the game",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		name := args[0]
		if err := app.ValidateProfileName(name); err != nil {
			return err
		}

		return withStateLock(p, func() error {
			cfg, game, err := requireGame(p)
			if err != nil {
				return err
			}
			cfg.ActiveProfile = name
			if err := app.SaveConfig(p.Config, *cfg); err != nil {
				return err
			}
			if err := app.EnsureProfileDirs(p, name); err != nil {
				return err
			}
			if err := deployActiveProfile(p, cfg, game.ModsDir); err != nil {
				return err
			}
			fmt.Println("Switched to profile:", name)
			return nil
		})
	},
}

var profileDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy enabled mods from the active profile into the game",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		return withStateLock(p, func() error {
			cfg, game, err := requireGame(p)
			if err != nil {
				return err
			}
			_, err = ensureActiveProfileDirs(p, cfg)
			if err != nil {
				return err
			}
			if err := deployActiveProfile(p, cfg, game.ModsDir); err != nil {
				return err
			}
			fmt.Println("Deployed active profile:", app.ActiveProfile(*cfg))
			return nil
		})
	},
}

func deployActiveProfile(p *app.Paths, cfg *app.Config, modsDir string) error {
	st, err := app.LoadState(p.State)
	if err != nil {
		return err
	}
	active := app.ActiveProfile(*cfg)

	// 1) Undeploy anything we previously deployed (any profile), but only folders we track.
	for id, me := range st.Mods {
		changed := false
		for prof, pi := range me.Installations {
			if pi.Enabled && pi.Folder != "" {
				if err := mods.Undeploy(modsDir, pi.Folder, id, prof); err != nil {
					return fmt.Errorf("failed to undeploy %s (%s): %w", id, prof, err)
				}
				pi.DeployedPath = ""
				// Keep enabled state; we'll redeploy active ones below.
				me.Installations[prof] = pi
				changed = true
			}
		}
		if changed {
			st.Mods[id] = me
		}
	}

	// 2) Deploy active profile enabled mods.
	for id, me := range st.Mods {
		pi, ok := me.Installations[active]
		if !ok || !pi.Installed || !pi.Enabled || pi.Folder == "" || pi.Store == "" {
			continue
		}
		storeAbs := filepath.Join(p.Root, filepath.FromSlash(pi.Store))
		if _, err := os.Stat(storeAbs); err != nil {
			continue
		}
		deployed, err := mods.Deploy(storeAbs, modsDir, pi.Folder, id, active)
		if err != nil {
			return err
		}
		pi.DeployedPath = deployed
		me.Installations[active] = pi
		st.Mods[id] = me
	}

	return app.SaveState(p.State, st)
}

func init() {
	profileCmd.AddCommand(profileStatusCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileUseCmd)
	profileCmd.AddCommand(profileDeployCmd)
}
