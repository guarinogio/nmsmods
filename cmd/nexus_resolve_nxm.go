package cmd

import (
	"encoding/json"
	"fmt"

	"nmsmods/internal/nexus"

	"github.com/spf13/cobra"
)

var nexusResolveNXMCmd = &cobra.Command{
	Use:   "resolve-nxm <nxm_url>",
	Short: "Resolve an nxm:// URL to one or more direct download URLs",
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

		nxmURL := args[0]
		nxm, err := nexus.ParseNXM(nxmURL)
		if err != nil {
			return err
		}

		ctx, cancel := nexusCtx()
		defer cancel()

		links, err := client.GetDownloadLinks(ctx, nxm.GameDomain, nxm.ModID, nxm.FileID, nxm.Key, nxm.Expires, nxm.UserID)
		if err != nil {
			return err
		}

		if f := cmd.Flags().Lookup("json"); f != nil && f.Changed {
			b, _ := json.MarshalIndent(links, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		}

		if len(links) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No download links returned.")
			return nil
		}

		for i, l := range links {
			label := l.ShortName
			if label == "" {
				label = l.Name
			}
			if label == "" {
				label = fmt.Sprintf("mirror-%d", i+1)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s\n%s\n", i+1, label, l.URI)
		}
		return nil
	},
}

func init() {
	nexusResolveNXMCmd.Flags().Bool("json", false, "Output in JSON format")
	nexusCmd.AddCommand(nexusResolveNXMCmd)
}
