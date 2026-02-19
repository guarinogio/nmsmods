
package app

import (
	"encoding/json"
	"os"
	"time"
)

type ModEntry struct {
	URL         string `json:"url"`
	ZIP         string `json:"zip"`   // relative to Paths.Root
	Folder      string `json:"folder"`// folder name installed under GAMEDATA/MODS
	Installed   bool   `json:"installed"`
	DownloadedAt string `json:"downloaded_at,omitempty"`
}

type State struct {
	Mods map[string]ModEntry `json:"mods"`
}

func LoadState(path string) (State, error) {
	var s State
	s.Mods = map[string]ModEntry{}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, err
	}
	if len(b) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return s, err
	}
	if s.Mods == nil {
		s.Mods = map[string]ModEntry{}
	}
	return s, nil
}

func SaveState(path string, s State) error {
	if s.Mods == nil {
		s.Mods = map[string]ModEntry{}
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func NowRFC3339() string {
	return time.Now().Format(time.RFC3339)
}
