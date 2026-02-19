package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var downloadsJSON bool

type downloadRow struct {
	Index       int    `json:"index"`
	ID          string `json:"id"`
	Installed   bool   `json:"installed"`
	ZipPath     string `json:"zip_path"`
	URL         string `json:"url,omitempty"`
	Folder      string `json:"folder,omitempty"`
	DownloadedAt string `json:"downloaded_at,omitempty"`
}

var downloadsCmd = &cobra.Command{
	Use:   "downloads",
	Short: "List downloaded mods tracked in state.json (with numeric indices)",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		st, err := app.LoadState(p.State)
		if err != nil {
			return err
		}

		if len(st.Mods) == 0 {
			if downloadsJSON {
				fmt.Println("[]")
				return nil
			}
			fmt.Println("(none)")
			return nil
		}

		ids := sortedModIDs(st)

		if downloadsJSON {
			out := make([]downloadRow, 0, len(ids))
			for i, id := range ids {
				me := st.Mods[id]
				zipAbs := ""
				if me.ZIP != "" {
					zipAbs = filepath.Join(p.Root, filepath.FromSlash(me.ZIP))
				}
				out = append(out, downloadRow{
					Index:       i + 1,
					ID:          id,
					Installed:   me.Installed,
					ZipPath:     zipAbs,
					URL:         me.URL,
					Folder:      me.Folder,
					DownloadedAt: me.DownloadedAt,
				})
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		for i, id := range ids {
			me := st.Mods[id]
			zipAbs := ""
			if me.ZIP != "" {
				zipAbs = filepath.Join(p.Root, filepath.FromSlash(me.ZIP))
			}
			fmt.Printf("[%d] %s\tinstalled=%v\tzip=%s\n", i+1, id, me.Installed, zipAbs)
		}

		return nil
	},
}

func init() {
	downloadsCmd.Flags().BoolVar(&downloadsJSON, "json", false, "Output in JSON format")
}
