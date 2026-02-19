package app

import (
	"os"
	"path/filepath"
)

type Paths struct {
	Root      string
	Downloads string
	Staging   string
	Profiles  string

	Config string
	State  string
}

func PathsFromRoot(root string) *Paths {
	return &Paths{
		Root:      root,
		Downloads: filepath.Join(root, "downloads"),
		Staging:   filepath.Join(root, "staging"),
		Profiles:  filepath.Join(root, "profiles"),
		Config:    filepath.Join(root, "config.json"),
		State:     filepath.Join(root, "state.json"),
	}
}

func pathsFromXDG(home string) *Paths {
	// Config
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}
	configDir := filepath.Join(configHome, "nmsmods")

	// State (best place for caches/staging/download metadata)
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		stateHome = filepath.Join(home, ".local", "state")
	}
	stateDir := filepath.Join(stateHome, "nmsmods")

	return &Paths{
		Root:      stateDir,
		Downloads: filepath.Join(stateDir, "downloads"),
		Staging:   filepath.Join(stateDir, "staging"),
		Profiles:  filepath.Join(stateDir, "profiles"),
		Config:    filepath.Join(configDir, "config.json"),
		State:     filepath.Join(stateDir, "state.json"),
	}
}

// DefaultPathsWithOverride uses (in order):
// 1) explicit root override passed by caller (if non-empty)
// 2) env var NMSMODS_HOME
// 3) legacy ~/.nmsmods (if it already exists)
// 4) XDG dirs (XDG_STATE_HOME/XDG_CONFIG_HOME) with sensible defaults
func DefaultPathsWithOverride(rootOverride string) (*Paths, error) {
	if rootOverride != "" {
		return PathsFromRoot(rootOverride), nil
	}
	if v := os.Getenv("NMSMODS_HOME"); v != "" {
		return PathsFromRoot(v), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	legacy := filepath.Join(home, ".nmsmods")
	if _, err := os.Stat(legacy); err == nil {
		return PathsFromRoot(legacy), nil
	}

	return pathsFromXDG(home), nil
}

func DefaultPaths() (*Paths, error) {
	return DefaultPathsWithOverride("")
}

func (p *Paths) Ensure() error {
	if err := os.MkdirAll(p.Root, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(p.Downloads, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(p.Staging, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(p.Profiles, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p.Config), 0o755); err != nil {
		return err
	}
	return nil
}
