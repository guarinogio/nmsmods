package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var infoJSON bool

// infoOut is the stable JSON shape for `nmsmods info --json`.
// Keep the original keys for backward compatibility and add new fields as needed.
type infoOut struct {
	ID string `json:"id"`

	// Backward-compatible fields
	URL          string `json:"url,omitempty"`
	ZipPath      string `json:"zip_path,omitempty"` // absolute path to zip under downloads/
	Folder       string `json:"folder,omitempty"`
	Installed    bool   `json:"installed"`
	DownloadedAt string `json:"downloaded_at,omitempty"`

	// Newer state fields
	Source        string         `json:"source,omitempty"` // local|url|nexus
	DisplayName   string         `json:"display_name,omitempty"`
	Nexus         *app.NexusInfo `json:"nexus,omitempty"`
	InstalledAt   string         `json:"installed_at,omitempty"`
	InstalledPath string         `json:"installed_path,omitempty"`
	Health        string         `json:"health,omitempty"` // ok|warning
	SHA256        string         `json:"sha256,omitempty"`

	// Convenience: relative zip path as stored in state (debugging)
	ZipRel string `json:"zip_rel,omitempty"`
}

var infoCmd = &cobra.Command{
	Use:   "info <id-or-index>",
	Short: "Show detailed info about a tracked mod",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		st, err := app.LoadState(p.State)
		if err != nil {
			return err
		}
		id, err := resolveModArg(args[0], st)
		if err != nil {
			return err
		}
		me := st.Mods[id]

		zipAbs := ""
		if me.ZIP != "" {
			zipAbs = filepath.Join(p.Root, filepath.FromSlash(me.ZIP))
		}

		out := infoOut{
			ID:           id,
			URL:          me.URL,
			ZipPath:      zipAbs,
			Folder:       me.Folder,
			Installed:    me.Installed,
			DownloadedAt: me.DownloadedAt,

			Source:        me.Source,
			DisplayName:   me.DisplayName,
			Nexus:         me.Nexus,
			InstalledAt:   me.InstalledAt,
			InstalledPath: me.InstalledPath,
			Health:        me.Health,
			SHA256:        me.SHA256,

			ZipRel: me.ZIP,
		}

		if infoJSON {
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		}

		fmt.Println("id:        ", out.ID)
		if out.DisplayName != "" && out.DisplayName != out.ID {
			fmt.Println("name:      ", out.DisplayName)
		}
		if out.Source != "" {
			fmt.Println("source:    ", out.Source)
		}
		if out.URL != "" {
			fmt.Println("url:       ", out.URL)
		}
		if out.ZipPath != "" {
			fmt.Println("zip:       ", out.ZipPath)
		}
		if out.Folder != "" {
			fmt.Println("folder:    ", out.Folder)
		}
		fmt.Println("installed: ", out.Installed)
		if out.InstalledAt != "" {
			fmt.Println("installed:", out.InstalledAt)
		}
		if out.InstalledPath != "" {
			fmt.Println("path:     ", out.InstalledPath)
		}
		if out.Health != "" {
			fmt.Println("health:   ", out.Health)
		}
		if out.SHA256 != "" {
			fmt.Println("sha256:   ", out.SHA256)
		}
		if out.DownloadedAt != "" {
			fmt.Println("downloaded:", out.DownloadedAt)
		}
		// Nexus details (best-effort, only if present)
		if out.Nexus != nil {
			if out.Nexus.GameDomain != "" {
				fmt.Println("nexus.game:", out.Nexus.GameDomain)
			}
			if out.Nexus.ModID != 0 {
				fmt.Println("nexus.mod:", out.Nexus.ModID)
			}
			if out.Nexus.FileID != 0 {
				fmt.Println("nexus.file:", out.Nexus.FileID)
			}
			if out.Nexus.Version != "" {
				fmt.Println("nexus.ver:", out.Nexus.Version)
			}
		}

		return nil
	},
}

func init() {
	infoCmd.Flags().BoolVar(&infoJSON, "json", false, "Output in JSON format")
}
