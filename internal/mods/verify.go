package mods

import (
	"os"
	"path/filepath"
	"strings"
)

type VerifyResult struct {
	ZipExists       bool   `json:"zip_exists"`
	InstalledExists bool   `json:"installed_exists"`
	HasModFiles     bool   `json:"has_mod_files"`
	Reason          string `json:"reason,omitempty"`
}

// HasRelevantFiles returns true if the directory contains at least one .MBIN or .EXML (case-insensitive).
func HasRelevantFiles(root string) (bool, error) {
	found := false
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if strings.HasSuffix(name, ".mbin") || strings.HasSuffix(name, ".exml") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return false, err
	}
	return found, nil
}
