package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"nmsmods/internal/app"
	"nmsmods/internal/nexus"
	"nmsmods/internal/nms"
	"nmsmods/internal/steam"

	"github.com/spf13/cobra"
)

// nxmCmd provides a dedicated handler for Nexus "Mod Manager Download" links (nxm://...)
// so that a browser click can perform download + install/update + enable + deploy.
var nxmCmd = &cobra.Command{
	Use:   "nxm",
	Short: "Handle Nexus Mod Manager (nxm://) links",
}

var nxmHandleCmd = &cobra.Command{
	Use:   "handle <nxm_url>",
	Short: "One-click: download and auto-install/update a Nexus mod via an nxm:// link",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := mustPaths()
		quiet, _ := cmd.Flags().GetBool("quiet")
		profileFlag, _ := cmd.Flags().GetString("profile")
		rawURL := strings.TrimSpace(args[0])

		// Log + (best-effort) desktop notification because this is typically launched from a browser.
		return withStateLock(p, func() error {
			// Prefer the state root (XDG state dir) for logs.
			logPath := filepath.Join(p.Root, "nxm-handler.log")
			logf := func(format string, a ...any) {
				_ = os.MkdirAll(filepath.Dir(logPath), 0o755)
				f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
				if err != nil {
					return
				}
				defer f.Close()
				ts := time.Now().Format(time.RFC3339)
				_, _ = fmt.Fprintf(f, "%s "+format+"\n", append([]any{ts}, a...)...)
			}
			finish := func(title, msg string, err error) error {
				if err != nil {
					logf("ERROR: %s: %v", msg, err)
					if !quiet {
						_ = notify(title, msg+": "+err.Error())
					}
					return err
				}
				logf("OK: %s", msg)
				if !quiet {
					_ = notify(title, msg)
				}
				return nil
			}

			cfg, err := loadConfig(p)
			if err != nil {
				return finish("nmsmods", "failed to load config", err)
			}

			// Optional: override profile by setting it active (persistent).
			if pf := strings.TrimSpace(profileFlag); pf != "" {
				cfg.ActiveProfile = pf
				if err := app.SaveConfig(p.Config, cfg); err != nil {
					return finish("nmsmods", "failed to save config", err)
				}
			}

			// Ensure game path is configured; try auto-detect if missing.
			if strings.TrimSpace(cfg.GamePath) == "" {
				cand, derr := autoDetectSingleGamePath()
				if derr != nil {
					return finish("nmsmods", "game path not set", derr)
				}
				cfg.GamePath = cand
				if err := app.SaveConfig(p.Config, cfg); err != nil {
					return finish("nmsmods", "failed to save config", err)
				}
			}

			game, err := nms.ValidateGamePath(cfg.GamePath)
			if err != nil {
				return finish("nmsmods", "invalid game path", fmt.Errorf("run: nmsmods set-path <path> (or set-path --auto): %w", err))
			}
			if err := nms.EnsureModsDir(game); err != nil {
				return finish("nmsmods", "failed to ensure MODS dir", err)
			}

			// Ensure active profile dirs exist.
			profile, err := ensureActiveProfileDirs(p, &cfg)
			if err != nil {
				return finish("nmsmods", "failed to ensure profile dirs", err)
			}

			client, err := newNexusClientFromConfig(cfg)
			if err != nil {
				return finish("nmsmods", "nexus login required", err)
			}

			nxm, err := nexus.ParseNXM(rawURL)
			if err != nil {
				return finish("nmsmods", "invalid nxm url", err)
			}

			id := fmt.Sprintf("nx-%d", nxm.ModID)

			// Load state to decide install/update/no-op.
			st, err := app.LoadState(p.State)
			if err != nil {
				return finish("nmsmods", "failed to load state", err)
			}
			me := st.Mods[id]
			prevFileID := 0
			if me.Nexus != nil {
				prevFileID = me.Nexus.FileID
			}
			var pi app.ProfileInstall
			if me.Installations != nil {
				pi = me.Installations[profile]
			}

			// Resolve download links.
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			links, err := client.GetDownloadLinks(ctx, nxm.GameDomain, nxm.ModID, nxm.FileID, nxm.Key, nxm.Expires, nxm.UserID)
			if err != nil {
				return finish("nmsmods", "failed to resolve download links", err)
			}
			if len(links) == 0 || links[0].URI == "" {
				return finish("nmsmods", "no download links returned", fmt.Errorf("no download links returned"))
			}
			bestURL := links[0].URI

			// Reuse existing download command wiring.
			self, _ := os.Executable()
			if strings.TrimSpace(self) == "" {
				self = os.Args[0]
			}
			dl := exec.Command(self, "download", bestURL, "--id", id)
			dl.Env = os.Environ()
			_ = dl.Run() // output is not visible in handler contexts

			// Enrich state with Nexus metadata (best-effort).
			st, err = app.LoadState(p.State)
			if err == nil {
				me = st.Mods[id]
				me.Source = "nexus"

				modInfo, _ := client.GetMod(ctx, nxm.GameDomain, nxm.ModID)
				files, _ := client.ListFiles(ctx, nxm.GameDomain, nxm.ModID)
				var fi *nexus.FileInfo
				for i := range files {
					if files[i].FileID == nxm.FileID {
						fi = &files[i]
						break
					}
				}

				if modInfo != nil && modInfo.Name != "" {
					if me.DisplayName == "" || me.DisplayName == id {
						me.DisplayName = modInfo.Name
					}
				} else if me.DisplayName == "" {
					me.DisplayName = id
				}

				ni := &app.NexusInfo{GameDomain: nxm.GameDomain, ModID: nxm.ModID, FileID: nxm.FileID}
				if modInfo != nil {
					ni.ModName = modInfo.Name
					ni.ModUpdatedTime = modInfo.UpdatedTime
					if ni.Version == "" {
						ni.Version = modInfo.Version
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
			}

			// Decide action.
			action := "install"
			if pi.Installed {
				if prevFileID == nxm.FileID && prevFileID != 0 {
					action = "enable"
				} else {
					action = "reinstall"
				}
			}

			var runErr error
			switch action {
			case "enable":
				runErr = exec.Command(self, "enable", id).Run()
			case "reinstall":
				runErr = exec.Command(self, "reinstall", id).Run()
			default:
				runErr = exec.Command(self, "install", id).Run()
			}
			if runErr != nil {
				return finish("nmsmods", "auto-install failed", runErr)
			}
			_ = exec.Command(self, "profile", "deploy").Run()

			switch action {
			case "enable":
				return finish("nmsmods", fmt.Sprintf("Already installed; ensured enabled: %s", id), nil)
			case "reinstall":
				return finish("nmsmods", fmt.Sprintf("Updated & redeployed: %s", id), nil)
			default:
				return finish("nmsmods", fmt.Sprintf("Installed & deployed: %s", id), nil)
			}
		})
	},
}

func init() {
	nxmHandleCmd.Flags().String("profile", "", "Override active profile (persists in config)")
	nxmHandleCmd.Flags().Bool("quiet", false, "Do not show desktop notifications (still logs to nxm-handler.log)")
	nxmCmd.AddCommand(nxmHandleCmd)
}

func autoDetectSingleGamePath() (string, error) {
	paths, err := steam.GuessNMSPaths()
	if err != nil {
		return "", err
	}
	valid := []string{}
	for _, cand := range paths {
		g, err := nms.ValidateGamePath(cand)
		if err == nil {
			valid = append(valid, g.Path)
		}
	}
	seen := map[string]struct{}{}
	uniq := []string{}
	for _, v := range valid {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		uniq = append(uniq, v)
	}
	if len(uniq) == 1 {
		return uniq[0], nil
	}
	if len(uniq) == 0 {
		return "", fmt.Errorf("no valid No Man's Sky installations found. Run: nmsmods set-path <path>")
	}
	return "", fmt.Errorf("multiple No Man's Sky installations detected; set one explicitly with: nmsmods set-path <path>\n- %s", strings.Join(uniq, "\n- "))
}

func notify(title, body string) error {
	if _, err := exec.LookPath("notify-send"); err != nil {
		return err
	}
	if len(body) > 500 {
		body = body[:500] + "…"
	}
	return exec.Command("notify-send", title, body).Run()
}
