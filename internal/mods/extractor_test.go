package mods

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writeZip(t *testing.T, zipPath string, files map[string]string) {
	t.Helper()
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range files {
		zf, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = zf.Write([]byte(content))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractZip_BlocksTraversal(t *testing.T) {
	tmp := t.TempDir()
	z := filepath.Join(tmp, "bad.zip")
	writeZip(t, z, map[string]string{
		"../../evil.txt": "nope",
	})

	dest := filepath.Join(tmp, "out")
	err := ExtractZip(z, dest)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestExtractZip_BlocksAbsolute(t *testing.T) {
	tmp := t.TempDir()
	z := filepath.Join(tmp, "abs.zip")

	absName := "/etc/passwd"
	if runtime.GOOS == "windows" {
		absName = "C:\\Windows\\system32\\drivers\\etc\\hosts"
	}

	writeZip(t, z, map[string]string{
		absName: "nope",
	})

	dest := filepath.Join(tmp, "out")
	err := ExtractZip(z, dest)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestExtractZip_OK(t *testing.T) {
	tmp := t.TempDir()
	z := filepath.Join(tmp, "ok.zip")
	writeZip(t, z, map[string]string{
		"ModA/FOO.EXML": "<xml/>",
	})

	dest := filepath.Join(tmp, "out")
	if err := ExtractZip(z, dest); err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	got := filepath.Join(dest, "ModA", "FOO.EXML")
	if _, err := os.Stat(got); err != nil {
		t.Fatalf("expected extracted file: %v", err)
	}
}
