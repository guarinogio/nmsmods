package mods

import (
	"archive/zip"
	"path"
	"strings"
)

// ProposedInstallFolderFromZip guesses the destination folder name without extracting.
//
// Heuristic:
//   - If all entries share a single top-level directory A, return A.
//   - If all entries share a single top-level A AND also share a single second-level directory B
//     AND entries are at least 3 segments deep (A/B/file...), return B (common "double folder").
//   - Else return fallbackID.
func ProposedInstallFolderFromZip(zipPath string, fallbackID string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	top := map[string]struct{}{}
	second := map[string]struct{}{}

	// Prefer paths that contain MODS/<folder>/... and .pak files.
	modsPakFolderCounts := map[string]int{}
	topPakCounts := map[string]int{}

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

		// Detect .pak files for better folder guess.
		isPak := !f.FileInfo().IsDir() && strings.HasSuffix(strings.ToLower(parts[len(parts)-1]), ".pak")
		if isPak {
			for i := 0; i < len(parts)-1; i++ {
				if strings.EqualFold(parts[i], "MODS") && i+1 < len(parts) {
					cand, _ := SanitizeFolderName(parts[i+1], fallbackID)
					modsPakFolderCounts[cand]++
					break
				}
			}
			if len(parts) >= 2 {
				cand, _ := SanitizeFolderName(parts[0], fallbackID)
				topPakCounts[cand]++
			}
		}
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
		return SanitizeFolderName(fallbackID, fallbackID)
	}

	// Best-effort: if we saw MODS/<folder> with .pak, prefer the most frequent.
	best := ""
	bestN := 0
	for k, n := range modsPakFolderCounts {
		if n > bestN {
			best = k
			bestN = n
		}
	}
	if best != "" {
		return best, nil
	}
	// Next: top-level folder with most .pak occurrences.
	best = ""
	bestN = 0
	for k, n := range topPakCounts {
		if n > bestN {
			best = k
			bestN = n
		}
	}
	if best != "" {
		return best, nil
	}

	if len(top) == 1 {
		// double-folder case
		if seenEligible && allEligibleForSecond && len(second) == 1 {
			for b := range second {
				return SanitizeFolderName(b, fallbackID)
			}
		}
		// single-folder case
		for a := range top {
			return SanitizeFolderName(a, fallbackID)
		}
	}

	return SanitizeFolderName(fallbackID, fallbackID)
}
