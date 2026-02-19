package app

import (
	"os"
	"path/filepath"
)

type Paths struct {
	Root      string
	Downloads string
	Staging   string
	Config    string
	State     string
}

func PathsFromRoot(root string) *Paths {
	return &Paths{
		Root:      root,
		Downloads: filepath.Join(root, "downloads"),
		Staging:   filepath.Join(root, "staging"),
		Config:    filepath.Join(root, "config.json"),
		State:     filepath.Join(root, "state.json"),
	}
}

// DefaultPaths uses (in order):
// 1) explicit root override passed by caller (if non-empty)
// 2) env var NMSMODS_HOME
// 3) ~/.nmsmods
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
	return PathsFromRoot(filepath.Join(home, ".nmsmods")), nil
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
	return nil
}
