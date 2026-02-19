package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHomeOverrideBeatsEnv(t *testing.T) {
	envHome := t.TempDir()
	overrideHome := t.TempDir()

	os.Setenv("NMSMODS_HOME", envHome)
	defer os.Unsetenv("NMSMODS_HOME")

	homeOverride = overrideHome
	defer func() { homeOverride = "" }()

	p := mustPaths()
	if filepath.Clean(p.Root) != filepath.Clean(overrideHome) {
		t.Fatalf("expected override to win")
	}
}
