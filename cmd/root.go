package cmd

import (
	"fmt"
	"io"
	"os"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var homeOverride string

var rootCmd = &cobra.Command{
	Use:     "nmsmods",
	Short:   "Manage No Man's Sky mods",
	Long:    "Download, install, list and uninstall NMS mods by managing folders under GAMEDATA/MODS.",
	Version: app.Version,
	// We print errors ourselves in Execute() so we can keep stdout clean for --json.
	SilenceErrors: true,
}

func Execute() {
	if err := ExecuteWithArgs(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		// Keep stderr stable and avoid Cobra printing twice.
		fmt.Fprintln(os.Stderr, err.Error())
		// For now exit with 1 for any command error.
		// (Later phases may introduce richer exit codes.)
		os.Exit(1)
	}
}

// ExecuteWithArgs runs the CLI without calling os.Exit (test-friendly).
func ExecuteWithArgs(args []string, out io.Writer, errOut io.Writer) error {
	rootCmd.SetOut(out)
	rootCmd.SetErr(errOut)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&homeOverride, "home", "", "Override nmsmods data directory (defaults to $NMSMODS_HOME, ~/.nmsmods, or XDG dirs)")
	rootCmd.SetVersionTemplate("nmsmods {{.Version}}\n")

	// If a command supports --json and the flag is set, suppress usage on errors.
	// This keeps stdout as valid JSON (no usage text appended).
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if f := cmd.Flags().Lookup("json"); f != nil && f.Changed {
			cmd.SilenceUsage = true
		}
	}

	registerCommands(rootCmd)
}
