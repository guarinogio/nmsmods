
package steam

import (
	"os"
	"path/filepath"
)

// GuessNMSPaths tries a few common Steam locations for No Man's Sky on Linux.
func GuessNMSPaths() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	candidates := []string{
		filepath.Join(home, ".local", "share", "Steam", "steamapps", "common", "No Man's Sky"),
		filepath.Join(home, ".steam", "steam", "steamapps", "common", "No Man's Sky"),
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".local", "share", "Steam", "steamapps", "common", "No Man's Sky"),
	}
	exists := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			exists = append(exists, c)
		}
	}
	return exists, nil
}
