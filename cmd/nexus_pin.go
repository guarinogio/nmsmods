package cmd

import (
    "fmt"

    "nmsmods/internal/app"

    "github.com/spf13/cobra"
)

// Phase 4: pin/unpin Nexus-tracked mods to avoid update prompts.

var nexusPinCmd = &cobra.Command{
    Use:   "pin <id-or-index>",
    Short: "Pin a Nexus-tracked mod (prevent check-updates from flagging it)",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        p := mustPaths()

        return withStateLock(p, func() error {
            st, err := app.LoadState(p.State)
            if err != nil {
                return err
            }
            id, err := resolveModArg(args[0], st)
            if err != nil {
                return err
            }
            me := st.Mods[id]
            if me.Nexus == nil {
                return fmt.Errorf("mod %s is not Nexus-tracked", id)
            }

            on, _ := cmd.Flags().GetBool("on")
            off, _ := cmd.Flags().GetBool("off")
            if on && off {
                return fmt.Errorf("use only one of --on or --off")
            }

            // Default: turn on.
            desired := true
            if off {
                desired = false
            }
            if on {
                desired = true
            }

            me.Nexus.Pinned = desired
            st.Mods[id] = me
            if err := app.SaveState(p.State, st); err != nil {
                return err
            }

            if desired {
                fmt.Fprintf(cmd.OutOrStdout(), "Pinned: %s\n", id)
            } else {
                fmt.Fprintf(cmd.OutOrStdout(), "Unpinned: %s\n", id)
            }
            return nil
        })
    },
}

func init() {
    nexusPinCmd.Flags().Bool("on", false, "Pin (explicit)")
    nexusPinCmd.Flags().Bool("off", false, "Unpin")
    nexusCmd.AddCommand(nexusPinCmd)
}
