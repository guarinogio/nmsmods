package cmd

import (
	"fmt"
	"path/filepath"
	"sort"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var downloadsCmd = &cobra.Command{
	Use:   "downloads",
	Short: "List downloaded mods tracked in state.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		st, err := app.LoadState(p.State)
		if err != nil {
			return err
		}

		if len(st.Mods) == 0 {
			fmt.Println("(none)")
			return nil
		}

		// Sort keys for stable numbering
		ids := make([]string, 0, len(st.Mods))
		for id := range st.Mods {
			ids = append(ids, id)
		}
		sort.Strings(ids)

		for i, id := range ids {
			me := st.Mods[id]
			zipAbs := ""
			if me.ZIP != "" {
				zipAbs = filepath.Join(p.Root, filepath.FromSlash(me.ZIP))
			}
			fmt.Printf("[%d] %s\tinstalled=%v\tzip=%s\n",
				i+1, id, me.Installed, zipAbs)
		}

		return nil
	},
}
