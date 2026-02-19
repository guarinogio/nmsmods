package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestState_MigratesV0ToV1(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "state.json")

	// v0: no state_version, old fields only
	v0 := []byte(`{
  "mods": {
    "m1": { "url": "https://x/y.zip", "zip": "downloads/y.zip", "installed": true, "folder": "ModA" }
  }
}`)
	if err := os.WriteFile(p, v0, 0o644); err != nil {
		t.Fatal(err)
	}

	st, err := LoadState(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.StateVersion != 1 {
		t.Fatalf("expected state_version=1, got %d", st.StateVersion)
	}
	me := st.Mods["m1"]
	if me.Source != "url" {
		t.Fatalf("expected source=url, got %q", me.Source)
	}
	if me.DisplayName == "" {
		t.Fatalf("expected display_name set")
	}
	if me.Health == "" {
		t.Fatalf("expected health set when installed")
	}

	// Ensure saved as v1
	st2, err := LoadState(p)
	if err != nil {
		t.Fatal(err)
	}
	if st2.StateVersion != 1 {
		t.Fatalf("expected persisted state_version=1, got %d", st2.StateVersion)
	}
}
