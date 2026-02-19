package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var downloadsJSON bool

// downloadRow is the stable JSON shape for `nmsmods downloads --json`.
type downloadRow struct {
	Index int    `json:"index"`
	ID    string `json:"id"`

	// Backward-compatible fields
	Installed    bool   `json:"installed"`
	ZipPath      string `json:"zip_path"`
	URL          string `json:"url,omitempty"`
	Folder       string `json:"folder,omitempty"`
	DownloadedAt string `json:"downloaded_at,omitempty"`

	// Newer state fields
	Source        string         `json:"source,omitempty"`
	DisplayName   string         `json:"display_name,omitempty"`
	Nexus         *app.NexusInfo `json:"nexus,omitempty"`
	InstalledAt   string         `json:"installed_at,omitempty"`
	InstalledPath string         `json:"installed_path,omitempty"`
	Health        string         `json:"health,omitempty"`
	SHA256        string         `json:"sha256,omitempty"`

	// Convenience
	ZipRel string `json:"zip_rel,omitempty"`
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
				fmt.Fprintln(cmd.OutOrStdout(), "[]")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "(none)")
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

					Source:        me.Source,
					DisplayName:   me.DisplayName,
					Nexus:         me.Nexus,
					InstalledAt:   me.InstalledAt,
					InstalledPath: me.InstalledPath,
					Health:        me.Health,
					SHA256:        me.SHA256,

					ZipRel: me.ZIP,
				})
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		}

		for i, id := range ids {
			me := st.Mods[id]
			zipAbs := ""
			if me.ZIP != "" {
				zipAbs = filepath.Join(p.Root, filepath.FromSlash(me.ZIP))
			}

			// Keep old single-line format, but add a tiny hint when available.
			// Example: [1] foo installed=false zip=/... source=nexus
			if me.Source != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s\tinstalled=%v\tzip=%s\tsource=%s\n", i+1, id, me.Installed, zipAbs, me.Source)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s\tinstalled=%v\tzip=%s\n", i+1, id, me.Installed, zipAbs)
			}
		}

		return nil
	},
}

func init() {
	downloadsCmd.Flags().BoolVar(&downloadsJSON, "json", false, "Output in JSON format")
}
