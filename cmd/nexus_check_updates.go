package cmd

import (
    "encoding/json"
    "fmt"
    "sort"

    "nmsmods/internal/app"

    "github.com/spf13/cobra"
)

// Phase 3: check updates for Nexus-tracked mods.
// This does NOT download anything (Nexus requires a fresh nxm:// key/expires/user_id for download links).

type nexusUpdateRow struct {
    ID         string         `json:"id"`
    Pinned     bool           `json:"pinned,omitempty"`
    Current    *app.NexusInfo `json:"current,omitempty"`
    Latest     *app.NexusInfo `json:"latest,omitempty"`
    HasUpdate  bool           `json:"has_update"`
    Reason     string         `json:"reason,omitempty"`
    ModUpdated string         `json:"mod_updated_time,omitempty"`
}

var nexusCheckUpdatesCmd = &cobra.Command{
    Use:   "check-updates [id-or-index]",
    Short: "Check if Nexus-tracked mods have updates available (metadata only)",
    Args:  cobra.RangeArgs(0, 1),
    RunE: func(cmd *cobra.Command, args []string) error {
        p, cfg, err := nexusPathsConfig()
        if err != nil {
            return err
        }
        client, err := newNexusClientFromConfig(cfg)
        if err != nil {
            return err
        }

        st, err := app.LoadState(p.State)
        if err != nil {
            return err
        }

        // Optional: single target
        targetID := ""
        if len(args) == 1 {
            id, rerr := resolveModArg(args[0], st)
            if rerr != nil {
                return rerr
            }
            targetID = id
        }

        ids := sortedModIDs(st)
        if targetID != "" {
            ids = []string{targetID}
        }

        ctx, cancel := nexusCtx()
        defer cancel()

        out := make([]nexusUpdateRow, 0, len(ids))

        for _, id := range ids {
            me := st.Mods[id]
            if me.Nexus == nil {
                continue
            }

            row := nexusUpdateRow{
                ID:      id,
                Current: me.Nexus,
                Pinned:  me.Nexus.Pinned,
            }
            if me.Nexus.Pinned {
                row.HasUpdate = false
                row.Reason = "pinned"
                out = append(out, row)
                continue
            }

            // Fetch latest mod + files (best-effort).
            mod, err := client.GetMod(ctx, me.Nexus.GameDomain, me.Nexus.ModID)
            if err == nil && mod != nil {
                row.ModUpdated = mod.UpdatedTime
            }

            files, err := client.ListFiles(ctx, me.Nexus.GameDomain, me.Nexus.ModID)
            if err != nil {
                row.HasUpdate = false
                row.Reason = fmt.Sprintf("list_files_failed: %v", err)
                out = append(out, row)
                continue
            }
            if len(files) == 0 {
                row.HasUpdate = false
                row.Reason = "no_files"
                out = append(out, row)
                continue
            }

            // Choose "latest" file heuristic:
            // 1) Prefer IsPrimary true
            // 2) Else prefer category MAIN (common Nexus convention)
            // 3) Else newest uploaded_timestamp
            // 4) Tiebreak: highest file_id
            sort.Slice(files, func(i, j int) bool {
                ai := files[i]
                aj := files[j]
                if ai.IsPrimary != aj.IsPrimary {
                    return ai.IsPrimary
                }
                aMain := ai.CategoryName == "MAIN"
                bMain := aj.CategoryName == "MAIN"
                if aMain != bMain {
                    return aMain
                }
                if ai.UploadedTimestamp != aj.UploadedTimestamp {
                    return ai.UploadedTimestamp > aj.UploadedTimestamp
                }
                return ai.FileID > aj.FileID
            })
            latest := files[0]

            latestInfo := &app.NexusInfo{
                GameDomain:        me.Nexus.GameDomain,
                ModID:             me.Nexus.ModID,
                FileID:            latest.FileID,
                ModName:           me.Nexus.ModName,
                FileName:          latest.FileName,
                Version:           latest.Version,
                CategoryName:      latest.CategoryName,
                UploadedTimestamp: latest.UploadedTimestamp,
                UploadedTime:      latest.UploadedTime,
                ModUpdatedTime:    row.ModUpdated,
            }
            row.Latest = latestInfo

            // Determine update: file_id differs OR uploaded_timestamp newer.
            if me.Nexus.FileID != 0 && latest.FileID != 0 && latest.FileID != me.Nexus.FileID {
                row.HasUpdate = true
                row.Reason = "file_id_changed"
            }
            if latest.UploadedTimestamp != 0 && me.Nexus.UploadedTimestamp != 0 && latest.UploadedTimestamp > me.Nexus.UploadedTimestamp {
                row.HasUpdate = true
                if row.Reason == "" {
                    row.Reason = "uploaded_timestamp_newer"
                }
            }
            if latest.Version != "" && me.Nexus.Version != "" && latest.Version != me.Nexus.Version {
                // Keep as supplementary signal
                if !row.HasUpdate {
                    row.HasUpdate = true
                    row.Reason = "version_changed"
                }
            }

            out = append(out, row)
        }

        if f := cmd.Flags().Lookup("json"); f != nil && f.Changed {
            b, _ := json.MarshalIndent(out, "", "  ")
            fmt.Fprintln(cmd.OutOrStdout(), string(b))
            return nil
        }

        if len(out) == 0 {
            fmt.Fprintln(cmd.OutOrStdout(), "No Nexus-tracked mods in state.")
            return nil
        }

        anyUpdate := false
        for _, r := range out {
            if r.Pinned {
                fmt.Fprintf(cmd.OutOrStdout(), "- %s: pinned\n", r.ID)
                continue
            }
            if r.HasUpdate {
                anyUpdate = true
                cur := ""
                if r.Current != nil {
                    if r.Current.Version != "" {
                        cur = "v" + r.Current.Version
                    } else if r.Current.FileID != 0 {
                        cur = fmt.Sprintf("file %d", r.Current.FileID)
                    }
                }
                lat := ""
                if r.Latest != nil {
                    if r.Latest.Version != "" {
                        lat = "v" + r.Latest.Version
                    } else if r.Latest.FileID != 0 {
                        lat = fmt.Sprintf("file %d", r.Latest.FileID)
                    }
                }
                if cur != "" || lat != "" {
                    fmt.Fprintf(cmd.OutOrStdout(), "- %s: update available (%s -> %s) [%s]\n", r.ID, cur, lat, r.Reason)
                } else {
                    fmt.Fprintf(cmd.OutOrStdout(), "- %s: update available [%s]\n", r.ID, r.Reason)
                }
            } else {
                fmt.Fprintf(cmd.OutOrStdout(), "- %s: up-to-date\n", r.ID)
            }
        }
        if anyUpdate {
            fmt.Fprintln(cmd.OutOrStdout(), "\nNote: Nexus downloads require a fresh nxm:// URL (key/expires/user_id). Use: nmsmods nexus download-nxm <nxm_url> --id <id> to update.")
        }
        return nil
    },
}

func init() {
    nexusCheckUpdatesCmd.Flags().Bool("json", false, "Output in JSON format")
    nexusCmd.AddCommand(nexusCheckUpdatesCmd)
}
