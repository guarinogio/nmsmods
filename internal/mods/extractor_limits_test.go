package mods

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func mkZipLimits(t *testing.T, zipPath string, entries map[string][]byte, headers []*zip.FileHeader) {
	t.Helper()
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)

	// regular entries
	for name, data := range entries {
		zf, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = zf.Write(data)
	}

	// custom headers (symlink etc)
	for _, h := range headers {
		zf, err := w.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = zf.Write([]byte("x"))
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func setEnvInt(t *testing.T, key string, v int) func() {
	t.Helper()
	prev, had := os.LookupEnv(key)
	if err := os.Setenv(key, strconv.Itoa(v)); err != nil {
		t.Fatal(err)
	}
	return func() {
		if had {
			_ = os.Setenv(key, prev)
		} else {
			_ = os.Unsetenv(key)
		}
	}
}

func TestExtractZip_RespectsMaxFiles(t *testing.T) {
	tmp := t.TempDir()
	z := filepath.Join(tmp, "many.zip")

	files := map[string][]byte{
		"a.txt": []byte("1"),
		"b.txt": []byte("2"),
		"c.txt": []byte("3"),
		"d.txt": []byte("4"),
	}

	restore := setEnvInt(t, "NMSMODS_MAX_FILES", 3)
	defer restore()

	mkZipLimits(t, z, files, nil)

	dest := filepath.Join(tmp, "out")
	err := ExtractZip(z, dest)
	if err == nil {
		t.Fatalf("expected error due to max files, got nil")
	}
}

func TestExtractZip_RespectsMaxFileBytes(t *testing.T) {
	tmp := t.TempDir()
	z := filepath.Join(tmp, "bigfile.zip")

	restore := setEnvInt(t, "NMSMODS_MAX_FILE_BYTES", 10)
	defer restore()

	mkZipLimits(t, z, map[string][]byte{
		"ModA/too_big.bin": []byte("01234567890"), // 11 bytes
	}, nil)

	dest := filepath.Join(tmp, "out")
	err := ExtractZip(z, dest)
	if err == nil {
		t.Fatalf("expected error due to max file bytes, got nil")
	}
}

func TestExtractZip_RespectsMaxTotalBytes(t *testing.T) {
	tmp := t.TempDir()
	z := filepath.Join(tmp, "total.zip")

	restore := setEnvInt(t, "NMSMODS_MAX_TOTAL_BYTES", 15)
	defer restore()

	mkZipLimits(t, z, map[string][]byte{
		"a.bin": []byte("0123456789"), // 10
		"b.bin": []byte("0123456789"), // 10 total 20 > 15
	}, nil)

	dest := filepath.Join(tmp, "out")
	err := ExtractZip(z, dest)
	if err == nil {
		t.Fatalf("expected error due to max total bytes, got nil")
	}
}

func TestExtractZip_BlocksSymlinkEntries(t *testing.T) {
	tmp := t.TempDir()
	z := filepath.Join(tmp, "symlink.zip")

	h := &zip.FileHeader{
		Name:   "link",
		Method: zip.Store,
	}
	// mark as symlink
	h.SetMode(os.ModeSymlink | 0o777)

	mkZipLimits(t, z, nil, []*zip.FileHeader{h})

	dest := filepath.Join(tmp, "out")
	err := ExtractZip(z, dest)
	if err == nil {
		t.Fatalf("expected error due to symlink entry, got nil")
	}
}
