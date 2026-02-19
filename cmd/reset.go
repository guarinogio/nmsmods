package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var resetAll bool
var resetKeepDownloads bool
var resetDryRun bool

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset nmsmods local state/caches (useful for troubleshooting and tests)",
	Long: `Reset removes local nmsmods files under the configured home directory.

Default behavior:
- remove state.json
- clean staging/

Options:
- --all: remove the entire nmsmods home directory and recreate it
- --keep-downloads: keep downloads/ (default true unless --all)
- --dry-run: show what would be removed without deleting`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()

		actions := []string{}

		if resetAll {
			actions = append(actions, fmt.Sprintf("remove: %s", p.Root))
			if resetDryRun {
				fmt.Println("[dry-run] Would reset:")
				for _, a := range actions {
					fmt.Println(" -", a)
				}
				return nil
			}
			if err := os.RemoveAll(p.Root); err != nil {
				return err
			}
			return p.Ensure()
		}

		// default: remove state.json and staging/
		actions = append(actions, fmt.Sprintf("remove: %s", p.State))
		actions = append(actions, fmt.Sprintf("clean:  %s", p.Staging))

		if !resetKeepDownloads {
			actions = append(actions, fmt.Sprintf("remove: %s", p.Downloads))
		}

		if resetDryRun {
			fmt.Println("[dry-run] Would reset:")
			for _, a := range actions {
				fmt.Println(" -", a)
			}
			return nil
		}

		_ = os.Remove(p.State)
		_ = os.RemoveAll(p.Staging)
		_ = os.MkdirAll(p.Staging, 0o755)

		if !resetKeepDownloads {
			_ = os.RemoveAll(p.Downloads)
			_ = os.MkdirAll(p.Downloads, 0o755)
		}

		// keep config.json by default (do not delete)
		_ = os.MkdirAll(filepath.Dir(p.Config), 0o755)

		fmt.Println("Reset completed.")
		return nil
	},
}

func init() {
	resetCmd.Flags().BoolVar(&resetAll, "all", false, "Remove the entire nmsmods home directory and recreate it")
	resetCmd.Flags().BoolVar(&resetKeepDownloads, "keep-downloads", true, "Keep downloads/ directory (ignored with --all)")
	resetCmd.Flags().BoolVar(&resetDryRun, "dry-run", false, "Print what would happen without making changes")
}
