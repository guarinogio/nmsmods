package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var nexusWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Validate current Nexus API key and show user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := nexusPathsConfig()
		if err != nil {
			return err
		}
		client, err := newNexusClientFromConfig(cfg)
		if err != nil {
			return err
		}

		ctx, cancel := nexusCtx()
		defer cancel()

		me, err := client.ValidateUser(ctx)
		if err != nil {
			return err
		}

		if f := cmd.Flags().Lookup("json"); f != nil && f.Changed {
			b, _ := json.MarshalIndent(me, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "user_id: %d\n", me.UserID)
		if me.Name != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "name:    %s\n", me.Name)
		}
		if me.Email != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "email:   %s\n", me.Email)
		}
		if me.IsPremium {
			fmt.Fprintln(cmd.OutOrStdout(), "premium: true")
		}
		if me.IsSupporter {
			fmt.Fprintln(cmd.OutOrStdout(), "supporter: true")
		}
		return nil
	},
}

func init() {
	nexusWhoamiCmd.Flags().Bool("json", false, "Output in JSON format")
	nexusCmd.AddCommand(nexusWhoamiCmd)
}
