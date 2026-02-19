package cmd

import (
	"fmt"

	"nmsmods/internal/app"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nmsmods", app.Version)
	},
}
