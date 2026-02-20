package mods

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const managedMarkerFile = ".nmsmods.managed.json"

// ManagedMarker is written into deployed folders inside the game's MODS directory.
// It allows nmsmods to refuse clobbering/deleting folders it does not own.
type ManagedMarker struct {
	Tag      string `json:"tag"`
	ModID    string `json:"mod_id"`
	Profile  string `json:"profile"`
	Deployed string `json:"deployed_at"`
	Tool     string `json:"tool"`
}

func managedTag(modID, profile string) string {
	return "nmsmods:" + modID + ":" + profile
}

func ReadManagedMarker(dest string) (ManagedMarker, error) {
	b, err := os.ReadFile(filepath.Join(dest, managedMarkerFile))
	if err != nil {
		return ManagedMarker{}, err
	}
	var m ManagedMarker
	if err := json.Unmarshal(b, &m); err != nil {
		return ManagedMarker{}, err
	}
	m.Tag = strings.TrimSpace(m.Tag)
	return m, nil
}

func readManagedTag(dest string) (string, error) {
	m, err := ReadManagedMarker(dest)
	if err != nil {
		return "", err
	}
	return m.Tag, nil
}

func writeManagedMarker(dest, modID, profile string) error {
	m := ManagedMarker{
		Tag:      managedTag(modID, profile),
		ModID:    modID,
		Profile:  profile,
		Deployed: time.Now().Format(time.RFC3339),
		Tool:     "nmsmods",
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dest, managedMarkerFile), b, 0o644)
}

func randSuffix() string {
	rand.Seed(time.Now().UnixNano())
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Deploy copies a stored mod folder into the game's MODS directory.
//
// Safety/robustness:
// - Folder name is validated to be a safe single path segment.
// - Uses an atomic "stage then rename" approach to avoid partial deployments.
// - Refuses to overwrite an existing folder that is not managed by nmsmods.
//
// Returns the deployed path (<gameModsDir>/<folder>).
func Deploy(storePath string, gameModsDir string, folder string, modID string, profile string) (string, error) {
	folder, err := SanitizeFolderName(folder, modID)
	if err != nil {
		return "", err
	}
	dest, err := SafeJoinUnder(gameModsDir, folder)
	if err != nil {
		return "", err
	}

	// Ensure MODS dir exists.
	if err := os.MkdirAll(gameModsDir, 0o755); err != nil {
		return "", err
	}

	// If destination exists, ensure it is managed (and matches this mod/profile).
	if st, err := os.Stat(dest); err == nil && st.IsDir() {
		tag, terr := readManagedTag(dest)
		if terr != nil {
			return "", fmt.Errorf("destination exists but is not managed by nmsmods: %s", dest)
		}
		if tag != managedTag(modID, profile) {
			return "", fmt.Errorf("destination exists but is managed by a different mod/profile: %s", dest)
		}
	}

	// Stage into a temp dir inside MODS for atomic rename.
	stage := filepath.Join(gameModsDir, "."+folder+".nmsmods.tmp-"+randSuffix())
	_ = os.RemoveAll(stage)
	if err := CopyDir(storePath, stage); err != nil {
		_ = os.RemoveAll(stage)
		return "", err
	}
	if err := writeManagedMarker(stage, modID, profile); err != nil {
		_ = os.RemoveAll(stage)
		return "", err
	}

	// Swap stage into place.
	backup := ""
	if _, err := os.Stat(dest); err == nil {
		backup = filepath.Join(gameModsDir, "."+folder+".nmsmods.bak-"+randSuffix())
		_ = os.RemoveAll(backup)
		if err := os.Rename(dest, backup); err != nil {
			_ = os.RemoveAll(stage)
			return "", err
		}
	}
	if err := os.Rename(stage, dest); err != nil {
		_ = os.RemoveAll(stage)
		if backup != "" {
			_ = os.Rename(backup, dest)
		}
		return "", err
	}
	if backup != "" {
		_ = os.RemoveAll(backup)
	}

	return dest, nil
}

// Undeploy removes a deployed mod folder from the game MODS directory, if present.
// It refuses to delete paths outside gameModsDir.
func Undeploy(gameModsDir string, folder string, modID string, profile string) error {
	folder, err := SanitizeFolderName(folder, modID)
	if err != nil {
		return err
	}
	dest, err := SafeJoinUnder(gameModsDir, folder)
	if err != nil {
		return err
	}

	st, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("refusing to remove non-directory: %s", dest)
	}

	tag, terr := readManagedTag(dest)
	if terr != nil {
		return fmt.Errorf("refusing to remove unmanaged folder: %s", dest)
	}
	if tag != managedTag(modID, profile) {
		return fmt.Errorf("refusing to remove folder managed by different mod/profile: %s", dest)
	}

	return os.RemoveAll(dest)
}
