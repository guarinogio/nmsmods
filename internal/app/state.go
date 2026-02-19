package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const CurrentStateVersion = 1

type NexusInfo struct {
	GameID   int    `json:"game_id,omitempty"`
	ModID    int    `json:"mod_id,omitempty"`
	FileID   int    `json:"file_id,omitempty"`
	FileName string `json:"file_name,omitempty"`
	Version  string `json:"version,omitempty"`
}

type ModEntry struct {
	// Existing fields (keep for backward compat)
	URL          string `json:"url,omitempty"`
	ZIP          string `json:"zip,omitempty"`     // relative to Root (e.g. downloads/foo.zip)
	Folder       string `json:"folder,omitempty"`  // install folder name under GAMEDATA/MODS
	Installed    bool   `json:"installed,omitempty"`
	DownloadedAt string `json:"downloaded_at,omitempty"`

	// New fields (v1)
	Source        string     `json:"source,omitempty"` // "local" | "url" | "nexus"
	DisplayName   string     `json:"display_name,omitempty"`
	Nexus         *NexusInfo `json:"nexus,omitempty"`
	InstalledAt   string     `json:"installed_at,omitempty"`
	InstalledPath string     `json:"installed_path,omitempty"`
	Health        string     `json:"health,omitempty"` // "ok" | "warning"
	SHA256        string     `json:"sha256,omitempty"`
}

type State struct {
	StateVersion int                `json:"state_version,omitempty"`
	Mods         map[string]ModEntry `json:"mods,omitempty"`
}

func NowRFC3339() string { return time.Now().Format(time.RFC3339) }

func LoadState(path string) (State, error) {
	// Default empty state if file doesn't exist
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return State{StateVersion: CurrentStateVersion, Mods: map[string]ModEntry{}}, nil
		}
		return State{}, err
	}

	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return State{}, err
	}

	// Backward compat: if missing map
	if st.Mods == nil {
		st.Mods = map[string]ModEntry{}
	}

	// Migration: missing StateVersion implies v0
	if st.StateVersion == 0 {
		st = migrateV0toV1(st)
		// Best-effort write-back so future loads are fast/stable
		_ = SaveState(path, st)
	}

	// Future: if more versions, chain migrations here
	if st.StateVersion < CurrentStateVersion {
		// no-op for now (CurrentStateVersion=1)
		st.StateVersion = CurrentStateVersion
		_ = SaveState(path, st)
	}

	return st, nil
}

func migrateV0toV1(st State) State {
	st.StateVersion = 1
	for id, me := range st.Mods {
		// Infer Source if missing
		if me.Source == "" {
			// If URL starts with file:// or dir://
			if len(me.URL) >= 7 && me.URL[:7] == "file://" {
				me.Source = "local"
			} else if len(me.URL) >= 6 && me.URL[:6] == "dir://" {
				me.Source = "local"
			} else if me.URL != "" {
				me.Source = "url"
			}
		}
		// DisplayName default: id (stable)
		if me.DisplayName == "" {
			me.DisplayName = id
		}
		// Health default
		if me.Health == "" && me.Installed {
			me.Health = "ok"
		}
		st.Mods[id] = me
	}
	return st
}

// SaveState writes atomically (temp file + rename) to avoid partial/corrupted JSON.
func SaveState(path string, st State) error {
	if st.Mods == nil {
		st.Mods = map[string]ModEntry{}
	}
	if st.StateVersion == 0 {
		st.StateVersion = CurrentStateVersion
	}

	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
