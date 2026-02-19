// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nmsmods",
	Short: "Manage No Man's Sky mods",
	Long:  "Download, install, list and uninstall NMS mods by managing folders under GAMEDATA/MODS.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	registerCommands(rootCmd)
}
