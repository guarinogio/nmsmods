package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"
	"nmsmods/internal/nms"

	"github.com/spf13/cobra"
)

var doctorJSON bool
var doctorAutoSetPath bool

type doctorReport struct {
	OK                  bool     `json:"ok"`
	DataDir             string   `json:"data_dir"`
	Downloads           string   `json:"downloads"`
	Staging             string   `json:"staging"`
	ConfiguredGamePath  string   `json:"game_path,omitempty"`
	ModsDir             string   `json:"mods_dir,omitempty"`
	DetectedGamePaths   []string `json:"detected_game_paths,omitempty"`
	InstalledModFolders []string `json:"installed_mod_folders,omitempty"`
	ManagedModFolders   []string `json:"managed_mod_folders,omitempty"`
	ExternalModFolders  []string `json:"external_mod_folders,omitempty"`
	TrackedDownloads    int      `json:"tracked_downloads"`
	Issues              []string `json:"issues,omitempty"`
}

func listInstalledFolders(modsDir string) ([]string, error) {
	ents, err := os.ReadDir(modsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var folders []string
	for _, e := range ents {
		if e.IsDir() {
			folders = append(folders, e.Name())
		}
	}
	sort.Strings(folders)
	return folders, nil
}

func splitManagedFolders(modsDir string, folders []string) (managed []string, external []string) {
	for _, f := range folders {
		p := filepath.Join(modsDir, f)
		if _, err := mods.ReadManagedMarker(p); err == nil {
			managed = append(managed, f)
		} else {
			external = append(external, f)
		}
	}
	sort.Strings(managed)
	sort.Strings(external)
	return managed, external
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check setup and show current configuration/state",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()

		// Lock not strictly required (read-only), but doctor can opt-in to auto-set-path.
		run := func() error {
			rep := doctorReport{
				OK:        true,
				DataDir:   p.Root,
				Downloads: p.Downloads,
				Staging:   p.Staging,
			}

			cfg, err := loadConfig(p)
			if err != nil {
				rep.OK = false
				rep.Issues = append(rep.Issues, fmt.Sprintf("failed to read config: %v", err))
			} else {
				rep.ConfiguredGamePath = cfg.GamePath
			}

			// Detect possible game paths (do not persist by default).
			guesses, gerr := detectGamePaths()
			if gerr == nil && len(guesses) > 0 {
				rep.DetectedGamePaths = guesses
			}

			// If no game path configured, either fail or auto-set if requested.
			if rep.ConfiguredGamePath == "" {
				if doctorAutoSetPath && len(rep.DetectedGamePaths) > 0 {
					first := rep.DetectedGamePaths[0]
					if _, verr := nms.ValidateGamePath(first); verr != nil {
						rep.OK = false
						rep.Issues = append(rep.Issues, fmt.Sprintf("auto-set candidate invalid: %s (%v)", first, verr))
					} else {
						if err := app.SaveConfig(p.Config, app.Config{GamePath: first}); err != nil {
							rep.OK = false
							rep.Issues = append(rep.Issues, fmt.Sprintf("failed to auto-set game path: %v", err))
						} else {
							rep.ConfiguredGamePath = first
						}
					}
				} else {
					rep.OK = false
					rep.Issues = append(rep.Issues, "game path not set (run: nmsmods set-path <path> or nmsmods where)")
				}
			}

			// Validate configured path if present
			if rep.ConfiguredGamePath != "" {
				game, verr := nms.ValidateGamePath(rep.ConfiguredGamePath)
				if verr != nil {
					rep.OK = false
					rep.Issues = append(rep.Issues, fmt.Sprintf("invalid game path: %v", verr))
				} else {
					// Use Game.ModsDir from validator (no Root field in your Game struct)
					rep.ModsDir = game.ModsDir

					if err := nms.EnsureModsDir(game); err != nil {
						rep.OK = false
						rep.Issues = append(rep.Issues, fmt.Sprintf("mods dir not ready: %v", err))
					} else {
						folders, ferr := listInstalledFolders(game.ModsDir)
						if ferr != nil {
							rep.OK = false
							rep.Issues = append(rep.Issues, fmt.Sprintf("failed to list installed mods: %v", ferr))
						} else {
							rep.InstalledModFolders = folders
							rep.ManagedModFolders, rep.ExternalModFolders = splitManagedFolders(game.ModsDir, folders)
						}
					}
				}
			}

			// State read
			st, serr := app.LoadState(p.State)
			if serr != nil {
				rep.OK = false
				rep.Issues = append(rep.Issues, fmt.Sprintf("failed to read state: %v", serr))
			} else {
				rep.TrackedDownloads = len(st.Mods)
			}

			if doctorJSON {
				b, _ := json.MarshalIndent(rep, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(b))
			} else {
				fmt.Println("Data dir:", rep.DataDir)
				fmt.Println("Downloads:", rep.Downloads)
				fmt.Println("Staging:", rep.Staging)
				fmt.Println("Game path:", rep.ConfiguredGamePath)
				fmt.Println("Mods dir:", rep.ModsDir)
				if len(rep.DetectedGamePaths) > 0 {
					fmt.Println("Detected game paths:")
					for _, gp := range rep.DetectedGamePaths {
						fmt.Println(" -", gp)
					}
				}
				fmt.Println("Tracked downloads:", rep.TrackedDownloads)
				if len(rep.InstalledModFolders) > 0 {
					fmt.Println("Installed mod folders:", len(rep.InstalledModFolders))
					if len(rep.ManagedModFolders) > 0 {
						fmt.Println(" - managed by nmsmods:", len(rep.ManagedModFolders))
						for _, f := range rep.ManagedModFolders {
							fmt.Println("   -", f)
						}
					}
					if len(rep.ExternalModFolders) > 0 {
						fmt.Println(" - external/unmanaged:", len(rep.ExternalModFolders))
						for _, f := range rep.ExternalModFolders {
							fmt.Println("   -", f)
						}
					}
				}
				if len(rep.Issues) > 0 {
					fmt.Println("Issues:")
					for _, is := range rep.Issues {
						fmt.Println(" -", is)
					}
				}
			}

			if !rep.OK {
				return fmt.Errorf("doctor found issues")
			}
			return nil
		}

		if doctorAutoSetPath {
			return withStateLock(p, run)
		}
		return run()
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output in JSON format")
	doctorCmd.Flags().BoolVar(&doctorAutoSetPath, "auto-set-path", false, "Auto-detect and persist game path if not configured")
}
