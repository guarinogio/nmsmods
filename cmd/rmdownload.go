package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var rmDownloadCmd = &cobra.Command{
	Use:   "rm-download <id-or-index>",
	Short: "Remove the downloaded ZIP (and staging) for a tracked mod id/index; does not uninstall",
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

		// remove zip
		if me.ZIP != "" {
			zipAbs := filepath.Join(p.Root, filepath.FromSlash(me.ZIP))
			_ = os.Remove(zipAbs)
		}
		// remove staging
		stageDir := filepath.Join(p.Staging, id)
		_ = os.RemoveAll(stageDir)

		// keep state entry but clear zip reference
		me.ZIP = ""
		st.Mods[id] = me
		if err := app.SaveState(p.State, st); err != nil {
			return err
		}

		fmt.Println("Removed downloaded files for:", id)
		return nil
	},
}
