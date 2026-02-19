package mods

import (
	"fmt"
	"os"
	"path/filepath"
)

// ChooseInstallFolder decides which folder should be installed and which path should be copied.
//
// Deterministic rules:
// 1) If root contains exactly one directory and no files, descend into it (flatten 1 level).
// 2) Repeat step 1 up to 2 times (covers common "double folder" ZIP extractions).
// 3) If we descended at least once, we install into the LAST descended directory name,
//    copying from the final descended path (even if it contains files directly).
// 4) Otherwise:
//    - If root contains exactly one directory and no files: install that directory.
//    - Else: install whole root into fallbackFolderName.
//
// Returns: (folderName, sourcePathToCopy, error)
func ChooseInstallFolder(root string, fallbackFolderName string) (string, string, error) {
	root = filepath.Clean(root)

	cur := root
	lastDirName := ""
	descended := false

	// Flatten up to 2 levels when each level is a single directory (and nothing else)
	for depth := 0; depth < 2; depth++ {
		ents, err := os.ReadDir(cur)
		if err != nil {
			return "", "", err
		}
		if len(ents) != 1 || !ents[0].IsDir() {
			break
		}
		descended = true
		lastDirName = ents[0].Name()
		cur = filepath.Join(cur, lastDirName)
	}

	if descended {
		// Common case: ZIP extracts to Outer/Inner/files...
		// We want to install folder "Inner" and copy from ".../Outer/Inner".
		return lastDirName, cur, nil
	}

	ents, err := os.ReadDir(cur)
	if err != nil {
		return "", "", err
	}

	// If there is exactly one directory and no files, install that directory
	dirCount := 0
	fileCount := 0
	var onlyDirName string
	for _, e := range ents {
		if e.IsDir() {
			dirCount++
			onlyDirName = e.Name()
		} else {
			fileCount++
		}
	}
	if dirCount == 1 && fileCount == 0 {
		return onlyDirName, filepath.Join(cur, onlyDirName), nil
	}

	// Otherwise install root content into fallback folder
	if fallbackFolderName == "" {
		return "", "", fmt.Errorf("fallback folder name required")
	}
	return fallbackFolderName, cur, nil
}
