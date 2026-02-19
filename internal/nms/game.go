
package nms

import (
	"errors"
	"os"
	"path/filepath"
)

type Game struct {
	Root   string
	ModsDir string // <Root>/GAMEDATA/MODS
}

func ValidateGamePath(root string) (*Game, error) {
	if root == "" {
		return nil, errors.New("game path is empty")
	}
	stat, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, errors.New("game path is not a directory")
	}

	// Basic heuristics: expect GAMEDATA and Binaries directories (common to most installs).
	// Don't require executables since Linux/Proton variants differ.
	gamedata := filepath.Join(root, "GAMEDATA")
	if _, err := os.Stat(gamedata); err != nil {
		return nil, errors.New("GAMEDATA directory not found under game path")
	}
	binaries := filepath.Join(root, "Binaries")
	if _, err := os.Stat(binaries); err != nil {
		// allow missing, but most installs have it; keep a soft check by not erroring hard
	}

	mods := filepath.Join(root, "GAMEDATA", "MODS")
	return &Game{Root: root, ModsDir: mods}, nil
}

func EnsureModsDir(g *Game) error {
	if g == nil || g.ModsDir == "" {
		return errors.New("game not initialized")
	}
	return os.MkdirAll(g.ModsDir, 0o755)
}

func ListInstalledModFolders(g *Game) ([]string, error) {
	if g == nil {
		return nil, errors.New("game not initialized")
	}
	entries, err := os.ReadDir(g.ModsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}
