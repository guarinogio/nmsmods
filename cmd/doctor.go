package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"nmsmods/internal/app"
	"nmsmods/internal/nms"

	"github.com/spf13/cobra"
)

var doctorJSON bool

type doctorReport struct {
	OK                bool     `json:"ok"`
	DataDir           string   `json:"data_dir"`
	Downloads         string   `json:"downloads"`
	Staging           string   `json:"staging"`
	GamePath          string   `json:"game_path,omitempty"`
	ModsDir           string   `json:"mods_dir,omitempty"`
	InstalledModFolders []string `json:"installed_mod_folders,omitempty"`
	TrackedDownloads  int      `json:"tracked_downloads"`
	Issues            []string `json:"issues,omitempty"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check setup and show current configuration/state",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()

		rep := doctorReport{
			OK:        true,
			DataDir:   p.Root,
			Downloads: p.Downloads,
			Staging:   p.Staging,
		}

		cfg, err := loadConfigAndMaybeGuess(p)
		if err != nil {
			rep.OK = false
			rep.Issues = append(rep.Issues, "failed to load config: "+err.Error())
		} else {
			rep.GamePath = cfg.GamePath
			if cfg.GamePath == "" {
				rep.OK = false
				rep.Issues = append(rep.Issues, "game path not set (run: nmsmods set-path <path>)")
			} else {
				game, gerr := nms.ValidateGamePath(cfg.GamePath)
				if gerr != nil {
					rep.OK = false
					rep.Issues = append(rep.Issues, "invalid game path: "+gerr.Error())
				} else {
					rep.ModsDir = game.ModsDir
					if err := nms.EnsureModsDir(game); err != nil {
						rep.OK = false
						rep.Issues = append(rep.Issues, "failed to ensure mods dir: "+err.Error())
					} else {
						modsList, lerr := nms.ListInstalledModFolders(game)
						if lerr != nil {
							rep.OK = false
							rep.Issues = append(rep.Issues, "failed to list installed mods: "+lerr.Error())
						} else {
							rep.InstalledModFolders = modsList
						}
					}
				}
			}
		}

		st, serr := app.LoadState(p.State)
		if serr != nil {
			rep.OK = false
			rep.Issues = append(rep.Issues, "failed to load state: "+serr.Error())
		} else {
			rep.TrackedDownloads = len(st.Mods)
		}

		if doctorJSON {
			b, _ := json.MarshalIndent(rep, "", "  ")
			fmt.Println(string(b))
			if !rep.OK {
				return fmt.Errorf("doctor found issues")
			}
			return nil
		}

		fmt.Println("Data dir:", rep.DataDir)
		fmt.Println("Downloads:", rep.Downloads)
		fmt.Println("Staging:", rep.Staging)
		fmt.Println("Game path:", rep.GamePath)
		if rep.ModsDir != "" {
			fmt.Println("Mods dir:", rep.ModsDir)
		}
		if rep.InstalledModFolders != nil {
			fmt.Printf("Installed mod folders: %d\n", len(rep.InstalledModFolders))
			for _, m := range rep.InstalledModFolders {
				fmt.Println(" -", m)
			}
		}
		fmt.Println("Tracked downloads:", rep.TrackedDownloads)

		if !rep.OK {
			fmt.Println("Issues:")
			for _, i := range rep.Issues {
				fmt.Println(" -", i)
			}
			// Return error so exit code is non-zero
			return fmt.Errorf("doctor found issues")
		}

		// Also warn if downloads dir is not writable (common)
		if f, err := os.CreateTemp(rep.Downloads, ".writecheck-*"); err != nil {
			fmt.Println("Warning: downloads dir not writable:", rep.Downloads, err.Error())
		} else {
			_ = os.Remove(f.Name())
		}

		// sanity: state paths exist
		_ = os.MkdirAll(filepath.Dir(p.State), 0o755)

		return nil
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output in JSON format")
}
