package mods

import (
	"archive/zip"
	"path"
	"strings"
)

// ProposedInstallFolderFromZip inspects the ZIP central directory and guesses the folder
// that would be installed using the same rule as ChooseInstallFolder:
//
// - If the zip effectively contains a single top-level directory, return that directory name.
// - Otherwise return fallbackID.
//
// This lets `install --dry-run` predict the destination folder without extracting.
func ProposedInstallFolderFromZip(zipPath string, fallbackID string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	top := map[string]struct{}{}

	for _, f := range r.File {
		name := f.Name
		name = strings.ReplaceAll(name, "\\", "/")
		name = strings.TrimLeft(name, "/")
		if name == "" || name == "." {
			continue
		}

		// normalize and grab first segment
		clean := path.Clean(name)
		if clean == "." {
			continue
		}

		seg := clean
		if i := strings.Index(seg, "/"); i >= 0 {
			seg = seg[:i]
		}
		seg = strings.TrimSpace(seg)
		if seg == "" || seg == "." {
			continue
		}
		top[seg] = struct{}{}
		// quick exit if more than 1 distinct top-level item
		if len(top) > 1 {
			return fallbackID, nil
		}
	}

	// If we found exactly one top-level folder name, use it.
	if len(top) == 1 {
		for k := range top {
			return k, nil
		}
	}

	// Empty zip or weird structure: fallback
	return fallbackID, nil
}
