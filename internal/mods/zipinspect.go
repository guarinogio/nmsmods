package mods

import (
	"archive/zip"
	"path"
	"strings"
)

// ProposedInstallFolderFromZip guesses the destination folder name without extracting.
//
// Heuristic:
// - If all entries share a single top-level directory A, return A.
// - If all entries share a single top-level A AND also share a single second-level directory B
//   AND entries are at least 3 segments deep (A/B/file...), return B (common "double folder").
// - Else return fallbackID.
func ProposedInstallFolderFromZip(zipPath string, fallbackID string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	top := map[string]struct{}{}
	second := map[string]struct{}{}

	hasAny := false
	allEligibleForSecond := true // must be true only if ALL entries have len(parts) >= 3
	seenEligible := false        // at least one entry with len(parts) >= 3

	for _, f := range r.File {
		name := strings.ReplaceAll(f.Name, "\\", "/")
		name = strings.TrimLeft(name, "/")
		if name == "" || name == "." {
			continue
		}
		clean := path.Clean(name)
		if clean == "." {
			continue
		}
		hasAny = true

		parts := strings.Split(clean, "/")
		if len(parts) >= 1 && parts[0] != "" && parts[0] != "." {
			top[parts[0]] = struct{}{}
		}
		if len(top) > 1 {
			return fallbackID, nil
		}

		// Only consider "double folder" if entries look like A/B/file... (>=3 segments)
		if len(parts) >= 3 {
			seenEligible = true
			if parts[1] != "" && parts[1] != "." {
				second[parts[1]] = struct{}{}
			}
		} else {
			allEligibleForSecond = false
		}
	}

	if !hasAny {
		return fallbackID, nil
	}

	if len(top) == 1 {
		// double-folder case
		if seenEligible && allEligibleForSecond && len(second) == 1 {
			for b := range second {
				return b, nil
			}
		}
		// single-folder case
		for a := range top {
			return a, nil
		}
	}

	return fallbackID, nil
}
