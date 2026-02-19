package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var dlID string

func copyFile(src, dst string) (int64, error) {
	in, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return 0, err
	}

	tmp := dst + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return 0, err
	}

	n, err := io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return n, err
	}
	if cerr != nil {
		_ = os.Remove(tmp)
		return n, cerr
	}
	return n, os.Rename(tmp, dst)
}

var downloadCmd = &cobra.Command{
	Use:   "download <url-or-local-zip>",
	Short: "Download a mod ZIP from a URL, or import a local ZIP into ~/.nmsmods/downloads",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		src := args[0]

		// If src is a local file, import it
		if st, err := os.Stat(src); err == nil && !st.IsDir() {
			base := filepath.Base(src)
			if !isZipFile(base) {
				return fmt.Errorf("local file must be a .zip: %s", src)
			}

			id := dlID
			if id == "" {
				// derive id from filename
				name := strings.TrimSuffix(base, filepath.Ext(base))
				id = mods.SlugFromURL(name) // slug helper is URL-based, but fine for plain strings too
			}

			destAbs := filepath.Join(p.Downloads, base)
			fmt.Println("ID:", id)
			fmt.Println("Importing local ZIP to:", destAbs)

			if _, err := copyFile(src, destAbs); err != nil {
				return err
			}

			stt, err := app.LoadState(p.State)
			if err != nil {
				return err
			}
			relZip := filepath.ToSlash(filepath.Join("downloads", base))
			entry := stt.Mods[id]
			entry.URL = "file://" + src
			entry.ZIP = relZip
			entry.DownloadedAt = app.NowRFC3339()
			stt.Mods[id] = entry

			if err := app.SaveState(p.State, stt); err != nil {
				return err
			}

			fmt.Println("Done.")
			return nil
		}

		// Otherwise treat as URL
		rawURL := src

		id := dlID
		if id == "" {
			id = mods.SlugFromURL(rawURL)
		}

		noQuery := strings.SplitN(rawURL, "?", 2)[0]
		base := path.Base(noQuery)
		if !isZipFile(base) {
			base = base + ".zip"
		}
		destAbs := filepath.Join(p.Downloads, base)

		fmt.Println("ID:", id)
		fmt.Println("Saving to:", destAbs)
		if _, err := mods.DownloadToFile(rawURL, destAbs); err != nil {
			return err
		}

		stt, err := app.LoadState(p.State)
		if err != nil {
			return err
		}

		relZip := filepath.ToSlash(filepath.Join("downloads", base))
		entry := stt.Mods[id]
		entry.URL = rawURL
		entry.ZIP = relZip
		entry.DownloadedAt = app.NowRFC3339()
		stt.Mods[id] = entry

		if err := app.SaveState(p.State, stt); err != nil {
			return err
		}

		fmt.Println("Done.")
		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVar(&dlID, "id", "", "Override mod id (slug)")
}
