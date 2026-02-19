package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var downloadID string

var downloadCmd = &cobra.Command{
	Use:   "download <url-or-zip>",
	Short: "Download a mod ZIP from a URL, or import a local ZIP file into downloads/",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		input := args[0]

		return withStateLock(p, func() error {
			st, err := app.LoadState(p.State)
			if err != nil {
				return err
			}

			id := strings.TrimSpace(downloadID)
			if id == "" {
				id = mods.SlugFromURL(input)
			}

			// ensure state map exists
			if st.Mods == nil {
				st.Mods = map[string]app.ModEntry{}
			}

			me := st.Mods[id]
			if me.DisplayName == "" {
				me.DisplayName = id
			}

			// Local ZIP import
			if fileExists(input) {
				if !isZipFile(input) {
					return fmt.Errorf("not a zip file: %s", input)
				}

				base := filepath.Base(input)
				dest := filepath.Join(p.Downloads, base)

				// Copy into managed downloads
				if err := copyFile(input, dest); err != nil {
					return err
				}

				me.URL = "file://" + input
				me.Source = "local"
				me.ZIP = filepath.ToSlash(filepath.Join("downloads", base))
				me.DownloadedAt = app.NowRFC3339()

				st.Mods[id] = me
				if err := app.SaveState(p.State, st); err != nil {
					return err
				}

				fmt.Println("Imported to:", dest)
				return nil
			}

			// Remote URL download
			url := input
			outName := id + ".zip"
			dest := filepath.Join(p.Downloads, outName)

			fmt.Println("Downloading to:", dest)
			if err := mods.DownloadURLToFile(url, dest); err != nil {
				return err
			}

			me.URL = url
			me.Source = "url"
			me.ZIP = filepath.ToSlash(filepath.Join("downloads", outName))
			me.DownloadedAt = app.NowRFC3339()

			st.Mods[id] = me
			if err := app.SaveState(p.State, st); err != nil {
				return err
			}

			fmt.Println("Downloaded:", dest)
			return nil
		})
	},
}

func init() {
	downloadCmd.Flags().StringVar(&downloadID, "id", "", "Override mod id (slug)")
}

// copyFile copies src -> dst (overwrites).
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.ReadFrom(in)
	return err
}
