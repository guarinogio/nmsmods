package steam

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GuessNMSPaths tries to locate No Man's Sky on Linux Steam installations.
//
// It checks common Steam roots and also parses steamapps/libraryfolders.vdf
// to discover additional Steam library locations.
func GuessNMSPaths() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	steamRoots := []string{
		filepath.Join(home, ".local", "share", "Steam"),
		filepath.Join(home, ".steam", "steam"),
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".local", "share", "Steam"),
	}

	libs := map[string]struct{}{}
	for _, root := range steamRoots {
		if st, err := os.Stat(root); err == nil && st.IsDir() {
			libs[root] = struct{}{}
			for _, lp := range readLibraryFolders(root) {
				libs[lp] = struct{}{}
			}
		}
	}

	// Probe each library for the game.
	out := []string{}
	seen := map[string]struct{}{}
	for lib := range libs {
		cand := filepath.Join(lib, "steamapps", "common", "No Man's Sky")
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			cand = filepath.Clean(cand)
			if _, ok := seen[cand]; !ok {
				seen[cand] = struct{}{}
				out = append(out, cand)
			}
		}
	}

	return out, nil
}

func readLibraryFolders(steamRoot string) []string {
	vdf := filepath.Join(steamRoot, "steamapps", "libraryfolders.vdf")
	f, err := os.Open(vdf)
	if err != nil {
		return nil
	}
	defer f.Close()

	// Minimal parser that extracts any quoted value for keys "path" inside nested blocks,
	// plus legacy format where numeric keys map directly to a path string.
	paths := []string{}
	seen := map[string]struct{}{}

	s := bufio.NewScanner(f)
	// allow long lines
	s.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		k, v, ok := parseVDFKV(line)
		if !ok {
			continue
		}
		lk := strings.ToLower(k)
		if lk == "path" || isNumericKey(lk) {
			p := strings.ReplaceAll(v, "\\\\", "\\")
			p = filepath.Clean(p)
			if p == "." || p == "" {
				continue
			}
			if _, ok := seen[p]; !ok {
				seen[p] = struct{}{}
				paths = append(paths, p)
			}
		}
	}
	return paths
}

func parseVDFKV(line string) (key, val string, ok bool) {
	// Supports patterns:
	//   "key"  "value"
	//   "0"    "/path"
	// Ignores braces and unquoted tokens.
	if !strings.Contains(line, "\"") {
		return "", "", false
	}
	parts := extractQuoted(line)
	if len(parts) < 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func extractQuoted(s string) []string {
	out := []string{}
	in := false
	cur := strings.Builder{}
	esc := false
	for _, r := range s {
		if !in {
			if r == '"' {
				in = true
				cur.Reset()
			}
			continue
		}
		if esc {
			cur.WriteRune(r)
			esc = false
			continue
		}
		if r == '\\' {
			esc = true
			continue
		}
		if r == '"' {
			out = append(out, cur.String())
			in = false
			continue
		}
		cur.WriteRune(r)
	}
	return out
}

func isNumericKey(k string) bool {
	if k == "" {
		return false
	}
	for _, r := range k {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func debugf(_ string, _ ...any) {
	// Reserved for future debug logging without importing fmt everywhere.
	_ = fmt.Sprintf
}
