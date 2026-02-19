
package mods

import (
	"os"
	"path/filepath"
)

// ChooseInstallFolder returns the folder name to install and the source path to copy from staging.
// Heuristic:
// - if stagingDir has a single top-level directory, install that directory (folder name = that dir name)
// - otherwise, install the whole stagingDir as a folder named modID
func ChooseInstallFolder(stagingDir, modID string) (folderName string, sourcePath string, err error) {
	entries, err := os.ReadDir(stagingDir)
	if err != nil {
		return "", "", err
	}
	var dirs []os.DirEntry
	var nonDirs []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		} else {
			nonDirs = append(nonDirs, e)
		}
	}
	if len(dirs) == 1 && len(nonDirs) == 0 {
		// single dir
		return dirs[0].Name(), filepath.Join(stagingDir, dirs[0].Name()), nil
	}
	// multiple entries -> use staging root as folder named modID
	return modID, stagingDir, nil
}
