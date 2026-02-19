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
}

func Execute() {
	if err := ExecuteWithArgs(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
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
	rootCmd.PersistentFlags().StringVar(&homeOverride, "home", "", "Override nmsmods data directory (defaults to $NMSMODS_HOME or ~/.nmsmods)")
	rootCmd.SetVersionTemplate("nmsmods {{.Version}}\n")

	registerCommands(rootCmd)
}
