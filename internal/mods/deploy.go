package mods

import (
	"fmt"
	"os"
	"path/filepath"
)

// Deploy copies a stored mod folder into the game's MODS directory.
// It removes any existing destination folder first.
func Deploy(storePath string, gameModsDir string, folder string) (string, error) {
	dest := filepath.Join(gameModsDir, folder)
	if _, err := os.Stat(dest); err == nil {
		if err := os.RemoveAll(dest); err != nil {
			return "", err
		}
	}
	if err := CopyDir(storePath, dest); err != nil {
		return "", err
	}
	return dest, nil
}

// Undeploy removes a deployed mod folder from the game MODS directory, if present.
// It refuses to delete paths outside gameModsDir.
func Undeploy(gameModsDir string, folder string) error {
	dest := filepath.Join(gameModsDir, folder)
	cleanMods := filepath.Clean(gameModsDir)
	cleanDest := filepath.Clean(dest)
	if len(cleanDest) < len(cleanMods) || cleanDest[:len(cleanMods)] != cleanMods {
		return fmt.Errorf("refusing to remove outside mods dir: %s", dest)
	}
	if _, err := os.Stat(dest); err == nil {
		return os.RemoveAll(dest)
	}
	return nil
}
