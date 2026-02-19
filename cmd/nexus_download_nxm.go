package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"nmsmods/internal/app"
	"nmsmods/internal/nexus"

	"github.com/spf13/cobra"
)

var nexusDownloadNXMCmd = &cobra.Command{
	Use:   "download-nxm <nxm_url> --id <id>",
	Short: "Resolve an nxm:// URL and download the ZIP into nmsmods downloads, storing Nexus metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, cfg, err := nexusPathsConfig()
		if err != nil {
			return err
		}
		client, err := newNexusClientFromConfig(cfg)
		if err != nil {
			return err
		}

		id, _ := cmd.Flags().GetString("id")
		id = strings.TrimSpace(id)
		if id == "" {
			return fmt.Errorf("missing --id")
		}

		nxm, err := nexus.ParseNXM(args[0])
		if err != nil {
			return err
		}

		ctx, cancel := nexusCtx()
		defer cancel()

		links, err := client.GetDownloadLinks(ctx, nxm.GameDomain, nxm.ModID, nxm.FileID, nxm.Key, nxm.Expires, nxm.UserID)
		if err != nil {
			return err
		}
		if len(links) == 0 || links[0].URI == "" {
			return fmt.Errorf("no download links returned")
		}
		bestURL := links[0].URI

		// Run existing downloader (same binary) so we reuse its behavior/state wiring.
		self, err := os.Executable()
		if err != nil {
			self = os.Args[0]
		}

		// Preserve NMSMODS_HOME if set (tests rely on it).
		cmdExec := exec.Command(self, "download", bestURL, "--id", id)
		cmdExec.Stdout = cmd.OutOrStdout()
		cmdExec.Stderr = cmd.ErrOrStderr()
		cmdExec.Env = os.Environ()

		if err := cmdExec.Run(); err != nil {
			return err
		}

		// Enrich state with Nexus metadata (best-effort).
		st, err := app.LoadState(p.State)
		if err != nil {
			return nil // best-effort
		}
		me, ok := st.Mods[id]
		if !ok {
			return nil
		}

		mod, _ := client.GetMod(ctx, nxm.GameDomain, nxm.ModID)
		files, _ := client.ListFiles(ctx, nxm.GameDomain, nxm.ModID)

		var fi *nexus.FileInfo
		for i := range files {
			if files[i].FileID == nxm.FileID {
				fi = &files[i]
				break
			}
		}

		me.Source = "nexus"

		// Prefer the Nexus mod name over the stable id for display purposes.
		// Note: download command sets DisplayName=id by default, so we override when it looks like a default.
		if mod != nil && mod.Name != "" {
			if me.DisplayName == "" || me.DisplayName == id {
				me.DisplayName = mod.Name
			}
		} else if me.DisplayName == "" {
			me.DisplayName = id
		}

		ni := &app.NexusInfo{
			GameDomain: nxm.GameDomain,
			ModID:      nxm.ModID,
			FileID:     nxm.FileID,
		}
		if mod != nil {
			ni.ModName = mod.Name
			ni.ModUpdatedTime = mod.UpdatedTime
			if ni.Version == "" {
				ni.Version = mod.Version
			}
		}
		if fi != nil {
			ni.FileName = fi.FileName
			if fi.Version != "" {
				ni.Version = fi.Version
			}
			ni.CategoryName = fi.CategoryName
			ni.UploadedTimestamp = fi.UploadedTimestamp
			ni.UploadedTime = fi.UploadedTime
		}
		me.Nexus = ni

		st.Mods[id] = me
		_ = app.SaveState(p.State, st)

		return nil
	},
}

func init() {
	nexusDownloadNXMCmd.Flags().String("id", "", "Stable id for the downloaded mod")
	nexusDownloadNXMCmd.MarkFlagRequired("id")
	nexusCmd.AddCommand(nexusDownloadNXMCmd)
}
