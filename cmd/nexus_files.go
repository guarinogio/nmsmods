package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
)

var nexusFilesCmd = &cobra.Command{
	Use:   "files <mod_id>",
	Short: "List files for a mod (Main/Optional) from Nexus",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := nexusPathsConfig()
		if err != nil {
			return err
		}
		client, err := newNexusClientFromConfig(cfg)
		if err != nil {
			return err
		}

		modID, err := strconv.Atoi(args[0])
		if err != nil || modID <= 0 {
			return fmt.Errorf("invalid mod_id: %q", args[0])
		}

		ctx, cancel := nexusCtx()
		defer cancel()

		files, err := client.ListFiles(ctx, nexusGameDomain, modID)
		if err != nil {
			return err
		}

		sort.Slice(files, func(i, j int) bool {
			ai := files[i].UpdatedTime
			aj := files[j].UpdatedTime
			if ai != "" && aj != "" && ai != aj {
				return ai > aj
			}
			return files[i].FileID > files[j].FileID
		})

		if f := cmd.Flags().Lookup("json"); f != nil && f.Changed {
			b, _ := json.MarshalIndent(files, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		}

		if len(files) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No files.")
			return nil
		}

		for _, f := range files {
			line := fmt.Sprintf("[%d]", f.FileID)
			if f.CategoryName != "" {
				line += " " + f.CategoryName
			}
			if f.Version != "" {
				line += " — v" + f.Version
			}
			if f.Name != "" {
				line += " — " + f.Name
			}
			fmt.Fprintln(cmd.OutOrStdout(), line)
		}
		return nil
	},
}

func init() {
	nexusFilesCmd.Flags().Bool("json", false, "Output in JSON format")
	nexusCmd.AddCommand(nexusFilesCmd)
}
