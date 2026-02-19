package cmd

import "github.com/spf13/cobra"

// registerCommands centralizes wiring so root.go never drifts from actual cmd files.
func registerCommands(root *cobra.Command) {
	root.AddCommand(versionCmd)

	root.AddCommand(whereCmd)
	root.AddCommand(setPathCmd)
	root.AddCommand(doctorCmd)

	root.AddCommand(downloadCmd)
	root.AddCommand(downloadsCmd)
	root.AddCommand(infoCmd)
	root.AddCommand(verifyCmd)

	root.AddCommand(installCmd)
	root.AddCommand(installDirCmd)
	root.AddCommand(reinstallCmd)
	root.AddCommand(enableCmd)
	root.AddCommand(disableCmd)
	root.AddCommand(profileCmd)
	root.AddCommand(completionCmd)

	root.AddCommand(installedCmd)
	root.AddCommand(uninstallCmd)
	root.AddCommand(rmDownloadCmd)

	root.AddCommand(cleanCmd)
	root.AddCommand(resetCmd)

	// Phase 1: Nexus API group
	root.AddCommand(nexusCmd)
}
