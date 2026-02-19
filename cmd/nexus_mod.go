package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var nexusModCmd = &cobra.Command{
	Use:   "mod <mod_id>",
	Short: "Get mod details from Nexus",
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

		mod, err := client.GetMod(ctx, nexusGameDomain, modID)
		if err != nil {
			return err
		}

		if f := cmd.Flags().Lookup("json"); f != nil && f.Changed {
			b, _ := json.MarshalIndent(mod, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "mod_id:    %d\n", mod.ModID)
		if mod.Name != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "name:      %s\n", mod.Name)
		}
		if mod.Author != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "author:    %s\n", mod.Author)
		}
		if mod.Version != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "version:   %s\n", mod.Version)
		}
		if mod.UpdatedTime != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "updated:   %s\n", mod.UpdatedTime)
		}
		if mod.Summary != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "summary:   %s\n", oneLine(mod.Summary))
		}
		return nil
	},
}

func init() {
	nexusModCmd.Flags().Bool("json", false, "Output in JSON format")
	nexusCmd.AddCommand(nexusModCmd)
}
