package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var infoJSON bool

type infoOut struct {
	ID           string `json:"id"`
	URL          string `json:"url,omitempty"`
	ZipPath      string `json:"zip_path,omitempty"`
	Folder       string `json:"folder,omitempty"`
	Installed    bool   `json:"installed"`
	DownloadedAt string `json:"downloaded_at,omitempty"`
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
		}

		if infoJSON {
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		fmt.Println("id:        ", out.ID)
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
		if out.DownloadedAt != "" {
			fmt.Println("downloaded:", out.DownloadedAt)
		}
		return nil
	},
}

func init() {
	infoCmd.Flags().BoolVar(&infoJSON, "json", false, "Output in JSON format")
}
