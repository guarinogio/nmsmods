package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var cleanDryRun bool
var cleanStaging bool
var cleanParts bool
var cleanOrphanZips bool

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean staging and other cached files",
	Long: `Clean removes local caches under ~/.nmsmods.

Defaults:
- cleans staging/ (removes extracted temp folders)

Optional:
- --parts: remove *.part files in downloads/
- --orphan-zips: remove ZIP files in downloads/ that are not referenced by state.json
- --dry-run: show what would be deleted without deleting`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()

		return withStateLock(p, func() error {
			st, err := app.LoadState(p.State)
			if err != nil {
				return err
			}

			// Default behavior: staging cleanup enabled if no flags explicitly set
			if !cmd.Flags().Changed("staging") && !cmd.Flags().Changed("parts") && !cmd.Flags().Changed("orphan-zips") {
				cleanStaging = true
			}

			var actions []string

			if cleanStaging {
				actions = append(actions, fmt.Sprintf("remove: %s", p.Staging))
				if !cleanDryRun {
					_ = os.RemoveAll(p.Staging)
					_ = os.MkdirAll(p.Staging, 0o755)
				}
			}

			if cleanParts {
				matches, _ := filepath.Glob(filepath.Join(p.Downloads, "*.part"))
				for _, m := range matches {
					actions = append(actions, fmt.Sprintf("remove: %s", m))
					if !cleanDryRun {
						_ = os.Remove(m)
					}
				}
			}

			if cleanOrphanZips {
				// Build set of referenced zip absolute paths
				ref := map[string]struct{}{}
				for _, me := range st.Mods {
					if me.ZIP == "" {
						continue
					}
					abs := joinPathFromState(p.Root, me.ZIP)
					ref[abs] = struct{}{}
				}

				entries, err := os.ReadDir(p.Downloads)
				if err == nil {
					for _, e := range entries {
						if e.IsDir() {
							continue
						}
						if !isZipFile(e.Name()) {
							continue
						}
						abs := filepath.Join(p.Downloads, e.Name())
						if _, ok := ref[abs]; !ok {
							actions = append(actions, fmt.Sprintf("remove orphan: %s", abs))
							if !cleanDryRun {
								_ = os.Remove(abs)
							}
						}
					}
				}
			}

			if len(actions) == 0 {
				fmt.Println("(nothing to do)")
				return nil
			}

			if cleanDryRun {
				fmt.Println("[dry-run] Would remove:")
			} else {
				fmt.Println("Cleaned:")
			}
			for _, a := range actions {
				fmt.Println(" -", a)
			}
			return nil
		})
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Print what would happen without making changes")
	cleanCmd.Flags().BoolVar(&cleanStaging, "staging", false, "Clean staging directory (extracted temp folders)")
	cleanCmd.Flags().BoolVar(&cleanParts, "parts", false, "Remove partial downloads (*.part) in downloads/")
	cleanCmd.Flags().BoolVar(&cleanOrphanZips, "orphan-zips", false, "Remove ZIPs in downloads/ not referenced by state.json")
}
