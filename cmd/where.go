package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var whereJSON bool

type whereReport struct {
	ConfiguredGamePath string   `json:"game_path,omitempty"`
	DetectedGamePaths  []string `json:"detected_game_paths,omitempty"`
}

var whereCmd = &cobra.Command{
	Use:   "where",
	Short: "Show detected/configured No Man's Sky path",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()

		cfg, err := loadConfig(p)
		if err != nil {
			return err
		}

		guesses, _ := detectGamePaths()

		if whereJSON {
			rep := whereReport{
				ConfiguredGamePath: cfg.GamePath,
				DetectedGamePaths:  guesses,
			}
			b, _ := json.MarshalIndent(rep, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		}

		// Preserve the old UX: print configured if set; otherwise print first detected (but do NOT persist).
		if cfg.GamePath != "" {
			fmt.Fprintln(cmd.OutOrStdout(), cfg.GamePath)
			return nil
		}
		if len(guesses) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), guesses[0])
			return nil
		}
		return fmt.Errorf("no game path configured and none detected (run: nmsmods set-path <path>)")
	},
}

func init() {
	whereCmd.Flags().BoolVar(&whereJSON, "json", false, "Output in JSON format")
}
