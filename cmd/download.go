// cmd/download.go
package cmd

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var dlID string

var downloadCmd = &cobra.Command{
	Use:   "download <url>",
	Short: "Download a mod ZIP from a direct URL into ~/.nmsmods/downloads",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		rawURL := args[0]

		id := dlID
		if id == "" {
			id = mods.SlugFromURL(rawURL)
		}

		// Name zip by URL basename without query
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

		st, err := app.LoadState(p.State)
		if err != nil {
			return err
		}

		relZip := filepath.ToSlash(filepath.Join("downloads", base))
		entry := st.Mods[id]
		entry.URL = rawURL
		entry.ZIP = relZip
		entry.DownloadedAt = app.NowRFC3339()
		st.Mods[id] = entry

		if err := app.SaveState(p.State, st); err != nil {
			return err
		}

		fmt.Println("Done.")
		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVar(&dlID, "id", "", "Override mod id (slug)")
}
