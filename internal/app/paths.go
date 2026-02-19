
package app

import (
	"errors"
	"os"
	"path/filepath"
)

type Paths struct {
	Root      string // ~/.nmsmods
	Config    string // ~/.nmsmods/config.json
	State     string // ~/.nmsmods/state.json
	Downloads string // ~/.nmsmods/downloads
	Staging   string // ~/.nmsmods/staging
}

func DefaultPaths() (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	root := filepath.Join(home, ".nmsmods")
	return &Paths{
		Root:      root,
		Config:    filepath.Join(root, "config.json"),
		State:     filepath.Join(root, "state.json"),
		Downloads: filepath.Join(root, "downloads"),
		Staging:   filepath.Join(root, "staging"),
	}, nil
}

func (p *Paths) Ensure() error {
	if p == nil || p.Root == "" {
		return errors.New("paths not initialized")
	}
	for _, dir := range []string{p.Root, p.Downloads, p.Staging} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
