package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"nmsmods/internal/app"
	"nmsmods/internal/mods"

	"github.com/spf13/cobra"
)

var verifyJSON bool

type verifyOut struct {
	ID              string           `json:"id"`
	ZipPath         string           `json:"zip_path,omitempty"`
	InstallFolder   string           `json:"install_folder,omitempty"`
	InstallPath     string           `json:"install_path,omitempty"`
	Result          mods.VerifyResult `json:"result"`
}

var verifyCmd = &cobra.Command{
	Use:   "verify <id-or-index>",
	Short: "Verify that a tracked mod is installed correctly and contains expected file types",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		_, game, err := requireGame(p)
		if err != nil {
			return err
		}

		st, err := app.LoadState(p.State)
		if err != nil {
			return err
		}

		id, err := resolveModArg(args[0], st)
		if err != nil {
			return err
		}

		me := st.Mods[id]
		out := verifyOut{ID: id}

		zipAbs := ""
		if me.ZIP != "" {
			zipAbs = joinPathFromState(p.Root, me.ZIP)
		}
		out.ZipPath = zipAbs

		res := mods.VerifyResult{}

		if zipAbs != "" {
			if _, err := os.Stat(zipAbs); err == nil {
				res.ZipExists = true
			}
		}

		folder := me.Folder
		if folder != "" {
			out.InstallFolder = folder
			installPath := filepath.Join(game.ModsDir, folder)
			out.InstallPath = installPath

			if st, err := os.Stat(installPath); err == nil && st.IsDir() {
				res.InstalledExists = true
				ok, verr := mods.HasRelevantFiles(installPath)
				if verr != nil {
					res.Reason = verr.Error()
				} else {
					res.HasModFiles = ok
					if !ok {
						res.Reason = "no .MBIN or .EXML files found under installed folder"
					}
				}
			} else {
				res.Reason = "installed folder not found"
			}
		} else {
			res.Reason = "no install folder recorded in state (mod may not be installed via nmsmods yet)"
		}

		out.Result = res

		if verifyJSON {
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		fmt.Println("id:           ", out.ID)
		if out.ZipPath != "" {
			fmt.Println("zip:          ", out.ZipPath, "exists=", res.ZipExists)
		} else {
			fmt.Println("zip:          ", "(none)")
		}
		if out.InstallPath != "" {
			fmt.Println("install:      ", out.InstallPath, "exists=", res.InstalledExists)
			fmt.Println("has_mod_files:", res.HasModFiles)
		} else {
			fmt.Println("install:      ", "(none)")
		}
		if res.Reason != "" {
			fmt.Println("note:         ", res.Reason)
		}
		return nil
	},
}

func init() {
	verifyCmd.Flags().BoolVar(&verifyJSON, "json", false, "Output in JSON format")
}
