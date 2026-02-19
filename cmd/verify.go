package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var verifyJSON bool

var verifyCmd = &cobra.Command{
	Use:   "verify <id-or-index>",
	Short: "Verify installed mod integrity",
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
		if !me.Installed {
			return fmt.Errorf("mod not installed")
		}

		countEXML := 0
		countMBIN := 0
		countPAK := 0

		err = filepath.Walk(me.InstalledPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			name := strings.ToLower(info.Name())
			if strings.HasSuffix(name, ".exml") {
				countEXML++
			}
			if strings.HasSuffix(name, ".mbin") {
				countMBIN++
			}
			if strings.HasSuffix(name, ".pak") {
				countPAK++
			}
			return nil
		})
		if err != nil {
			return err
		}

		result := map[string]any{
			"id":           id,
			"installed":    me.Installed,
			"exml_count":   countEXML,
			"mbin_count":   countMBIN,
			"pak_count":    countPAK,
			"health":       me.Health,
			"installed_at": me.InstalledAt,
		}

		if verifyJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		}

		fmt.Printf("EXML: %d, MBIN: %d, PAK: %d\n", countEXML, countMBIN, countPAK)
		if countPAK > 0 {
			fmt.Println("Warning: PAK files detected (likely incompatible with NMS 5.50+)")
		}
		return nil
	},
}

func init() {
	verifyCmd.Flags().BoolVar(&verifyJSON, "json", false, "Output in JSON format")
}
