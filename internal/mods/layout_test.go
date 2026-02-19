package mods

import (
	"os"
	"path/filepath"
	"testing"
)

func touch(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestChooseInstallFolder_FlattensDoubleFolder(t *testing.T) {
	tmp := t.TempDir()
	// root/A/B/file
	touch(t, filepath.Join(tmp, "A", "B", "X.EXML"))

	folder, src, err := ChooseInstallFolder(tmp, "fallback")
	if err != nil {
		t.Fatal(err)
	}
	if folder != "B" {
		t.Fatalf("expected folder B, got %s", folder)
	}
	if src != filepath.Join(tmp, "A", "B") {
		t.Fatalf("unexpected src: %s", src)
	}
}

func TestChooseInstallFolder_FallbackWhenMixed(t *testing.T) {
	tmp := t.TempDir()
	touch(t, filepath.Join(tmp, "A", "X.EXML"))
	touch(t, filepath.Join(tmp, "Y.EXML"))

	folder, src, err := ChooseInstallFolder(tmp, "fallback")
	if err != nil {
		t.Fatal(err)
	}
	if folder != "fallback" {
		t.Fatalf("expected fallback, got %s", folder)
	}
	if src != tmp {
		t.Fatalf("expected src root, got %s", src)
	}
}
