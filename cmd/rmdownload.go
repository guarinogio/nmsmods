package cmd

import (
	"fmt"
	"os"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var rmDownloadCmd = &cobra.Command{
	Use:   "rm-download <id-or-index>",
	Short: "Delete a downloaded ZIP from the downloads folder (does not uninstall)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()

		return withStateLock(p, func() error {
			st, err := app.LoadState(p.State)
			if err != nil {
				return err
			}
			id, err := resolveModArg(args[0], st)
			if err != nil {
				return err
			}
			me := st.Mods[id]
			if me.ZIP == "" {
				return fmt.Errorf("no zip tracked for %s", id)
			}

			zipAbs := joinPathFromState(p.Root, me.ZIP)
			if _, err := os.Stat(zipAbs); err == nil {
				if err := os.Remove(zipAbs); err != nil {
					return err
				}
			}

			me.ZIP = ""
			st.Mods[id] = me
			if err := app.SaveState(p.State, st); err != nil {
				return err
			}

			fmt.Println("Removed download:", zipAbs)
			return nil
		})
	},
}
