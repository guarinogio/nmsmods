package mods

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func mkZip(t *testing.T, zipPath string, names []string) {
	t.Helper()
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for _, n := range names {
		zf, err := w.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = zf.Write([]byte("x"))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestProposedInstallFolderFromZip_SingleFolder(t *testing.T) {
	tmp := t.TempDir()
	z := filepath.Join(tmp, "a.zip")
	mkZip(t, z, []string{"ModA/file.EXML"})
	got, err := ProposedInstallFolderFromZip(z, "fallback")
	if err != nil {
		t.Fatal(err)
	}
	if got != "ModA" {
		t.Fatalf("expected ModA, got %s", got)
	}
}

func TestProposedInstallFolderFromZip_DoubleFolder(t *testing.T) {
	tmp := t.TempDir()
	z := filepath.Join(tmp, "b.zip")
	mkZip(t, z, []string{"Outer/Inner/file.EXML"})
	got, err := ProposedInstallFolderFromZip(z, "fallback")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Inner" {
		t.Fatalf("expected Inner, got %s", got)
	}
}

func TestProposedInstallFolderFromZip_MixedFallback(t *testing.T) {
	tmp := t.TempDir()
	z := filepath.Join(tmp, "c.zip")
	mkZip(t, z, []string{"A/file.EXML", "B/file2.EXML"})
	got, err := ProposedInstallFolderFromZip(z, "fallback")
	if err != nil {
		t.Fatal(err)
	}
	if got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
}
