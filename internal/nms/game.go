package nms

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Game struct {
	Path    string
	DataDir string
	ModsDir string
}

// ValidateGamePath checks whether the given path looks like a real, installed No Man's Sky directory.
// This is intentionally stricter than "folder exists" because Steam can leave empty directories after uninstall.
//
// Requirements:
// - <path>/GAMEDATA exists
// - <path>/Binaries exists
// - <path>/GAMEDATA/PCBANKS exists
// - <path>/GAMEDATA/PCBANKS contains at least one .pak file
//
// If the user reinstalls the game to the same location, these checks will pass again automatically.
func ValidateGamePath(gamePath string) (*Game, error) {
	if gamePath == "" {
		return nil, fmt.Errorf("empty game path")
	}

	root := filepath.Clean(gamePath)

	st, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("game path not found: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("game path is not a directory: %s", root)
	}

	dataDir := filepath.Join(root, "GAMEDATA")
	binDir := filepath.Join(root, "Binaries")
	pcbanks := filepath.Join(dataDir, "PCBANKS")

	if !isDir(dataDir) {
		return nil, fmt.Errorf("missing GAMEDATA directory (game not installed here?): %s", dataDir)
	}
	if !isDir(binDir) {
		return nil, fmt.Errorf("missing Binaries directory (game not installed here?): %s", binDir)
	}
	if !isDir(pcbanks) {
		return nil, fmt.Errorf("missing GAMEDATA/PCBANKS directory (game not installed here?): %s", pcbanks)
	}

	// Strong signal the game content exists: at least one .pak in PCBANKS
	hasPak, err := dirHasFileWithExt(pcbanks, ".pak")
	if err != nil {
		return nil, fmt.Errorf("failed to inspect PCBANKS: %w", err)
	}
	if !hasPak {
		return nil, fmt.Errorf("no .pak files found in PCBANKS (game likely not installed): %s", pcbanks)
	}

	return &Game{
		Path:    root,
		DataDir: dataDir,
		ModsDir: filepath.Join(dataDir, "MODS"),
	}, nil
}

// EnsureModsDir ensures the mods directory exists.
// This assumes ValidateGamePath has already succeeded; i.e., the game is installed.
func EnsureModsDir(g *Game) error {
	if g == nil {
		return fmt.Errorf("nil game")
	}
	return os.MkdirAll(g.ModsDir, 0o755)
}

// ListInstalledModFolders returns the list of subdirectories in GAMEDATA/MODS.
func ListInstalledModFolders(g *Game) ([]string, error) {
	if g == nil {
		return nil, fmt.Errorf("nil game")
	}
	if !isDir(g.ModsDir) {
		// not an error; means no mods dir yet
		return []string{}, nil
	}

	ents, err := os.ReadDir(g.ModsDir)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func dirHasFileWithExt(dir, ext string) (bool, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	ext = strings.ToLower(ext)
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if strings.HasSuffix(name, ext) {
			return true, nil
		}
	}
	return false, nil
}
