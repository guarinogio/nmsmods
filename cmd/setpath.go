package cmd

import (
	"errors"
	"fmt"
	"strings"

	"nmsmods/internal/app"
	"nmsmods/internal/nms"
	"nmsmods/internal/steam"

	"github.com/spf13/cobra"
)

var setPathCmd = &cobra.Command{
	Use:   "set-path <path> | --auto",
	Short: "Set the No Man's Sky installation path (Steam library folder)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()

		auto, _ := cmd.Flags().GetBool("auto")
		newPath := ""
		if len(args) > 0 {
			newPath = args[0]
		}
		if auto {
			if newPath != "" {
				return fmt.Errorf("--auto cannot be used with an explicit path")
			}
			paths, err := steam.GuessNMSPaths()
			if err != nil {
				return err
			}
			valid := []string{}
			for _, cand := range paths {
				g, err := nms.ValidateGamePath(cand)
				if err == nil {
					valid = append(valid, g.Path)
				}
			}
			// de-dupe
			seen := map[string]struct{}{}
			uniq := []string{}
			for _, v := range valid {
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				uniq = append(uniq, v)
			}
			if len(uniq) == 0 {
				return fmt.Errorf("could not auto-detect No Man's Sky install. Use: nmsmods set-path <path>")
			}
			if len(uniq) > 1 {
				msg := "multiple No Man's Sky installs detected. Please choose one with: nmsmods set-path <path>\n"
				for _, v := range uniq {
					msg += " - " + v + "\n"
				}
				return errors.New(strings.TrimSpace(msg))
			}
			newPath = uniq[0]
		}
		if newPath == "" {
			return fmt.Errorf("missing path (or use --auto)")
		}

		return withStateLock(p, func() error {
			// Validate now so config doesn't get junk
			g, err := nms.ValidateGamePath(newPath)
			if err != nil {
				return fmt.Errorf("invalid game path: %w", err)
			}

			cfg, err := app.LoadConfig(p.Config)
			if err != nil {
				return err
			}
			// Persist canonical path (after symlink resolution).
			canonical := g.Path
			cfg.GamePath = canonical
			if err := app.SaveConfig(p.Config, cfg); err != nil {
				return err
			}

			if canonical != newPath {
				fmt.Println("Game path set to:", canonical)
				fmt.Println("(canonicalized from:", newPath+")")
			} else {
				fmt.Println("Game path set to:", canonical)
			}
			return nil
		})
	},
}

func init() {
	setPathCmd.Flags().Bool("auto", false, "Auto-detect the No Man's Sky install path from Steam libraries")
}
