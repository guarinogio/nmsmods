package mods

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ChooseInstallFolder decides which folder should be installed and which path should be copied.
//
// Deterministic rules:
//  1. If root contains exactly one directory and no files, descend into it (flatten 1 level).
//  2. Repeat step 1 up to 2 times (covers common "double folder" ZIP extractions).
//  3. If we descended at least once, we install into the LAST descended directory name,
//     copying from the final descended path (even if it contains files directly).
//  4. Otherwise:
//     - If root contains exactly one directory and no files: install that directory.
//     - Else: install whole root into fallbackFolderName.
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
		folder, err := SanitizeFolderName(lastDirName, fallbackFolderName)
		if err != nil {
			return "", "", err
		}
		return folder, cur, nil
	}

	// If the extracted tree contains .pak files, prefer the directory that looks like
	// GAMEDATA/MODS/<Folder>/... or MODS/<Folder>/... to avoid installing the wrong level.
	if folder, src, ok, err := chooseFolderByPak(root, fallbackFolderName); err != nil {
		return "", "", err
	} else if ok {
		return folder, src, nil
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
		folder, err := SanitizeFolderName(onlyDirName, fallbackFolderName)
		if err != nil {
			return "", "", err
		}
		return folder, filepath.Join(cur, onlyDirName), nil
	}

	// Otherwise install root content into fallback folder
	if fallbackFolderName == "" {
		return "", "", fmt.Errorf("fallback folder name required")
	}
	folder, err := SanitizeFolderName(fallbackFolderName, fallbackFolderName)
	if err != nil {
		return "", "", err
	}
	return folder, cur, nil
}

func chooseFolderByPak(root string, fallback string) (folder string, srcPath string, ok bool, err error) {
	// Walk the extracted root and find .pak files.
	type cand struct {
		folder string
		src    string
		score  int
		count  int
	}
	cands := map[string]*cand{}

	err = filepath.WalkDir(root, func(p string, d os.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			// Skip common junk.
			bn := strings.ToLower(d.Name())
			if bn == "__macosx" {
				return filepath.SkipDir
			}
			return nil
		}
		name := strings.ToLower(d.Name())
		if !strings.HasSuffix(name, ".pak") {
			return nil
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		// Find MODS segment; prefer folder immediately after it.
		for i := 0; i < len(parts)-1; i++ {
			if strings.EqualFold(parts[i], "MODS") && i+1 < len(parts) {
				f := parts[i+1]
				f2, serr := SanitizeFolderName(f, fallback)
				if serr != nil {
					continue
				}
				src := filepath.Join(root, filepath.FromSlash(strings.Join(parts[:i+2], "/")))
				key := f2 + "|" + src
				c := cands[key]
				if c == nil {
					c = &cand{folder: f2, src: src, score: 100}
					cands[key] = c
				}
				c.count++
				return nil
			}
		}
		// Fallback: top-level folder, if any.
		if len(parts) >= 2 {
			f2, serr := SanitizeFolderName(parts[0], fallback)
			if serr == nil {
				src := filepath.Join(root, parts[0])
				key := f2 + "|" + src
				c := cands[key]
				if c == nil {
					c = &cand{folder: f2, src: src, score: 10}
					cands[key] = c
				}
				c.count++
			}
		}
		return nil
	})
	if err != nil {
		return "", "", false, err
	}
	if len(cands) == 0 {
		return "", "", false, nil
	}
	// Pick best score, then highest pak count.
	best := (*cand)(nil)
	for _, c := range cands {
		if best == nil || c.score > best.score || (c.score == best.score && c.count > best.count) {
			best = c
		}
	}
	if best == nil {
		return "", "", false, nil
	}
	return best.folder, best.src, true, nil
}
