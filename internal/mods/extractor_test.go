package mods

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZip_RejectsTraversal(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "bad.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("../evil.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte("nope"))
	_ = zw.Close()
	_ = f.Close()

	dest := filepath.Join(tmp, "out")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := ExtractZip(zipPath, dest); err == nil {
		t.Fatalf("expected error for traversal zip")
	}
}
