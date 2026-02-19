package mods

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasRelevantFiles(t *testing.T) {
	tmp := t.TempDir()
	if ok, err := HasRelevantFiles(tmp); err != nil || ok {
		t.Fatalf("expected false, got ok=%v err=%v", ok, err)
	}

	if err := os.WriteFile(filepath.Join(tmp, "x.MBIN"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	ok, err := HasRelevantFiles(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("expected true")
	}
}
