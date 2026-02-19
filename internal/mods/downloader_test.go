package mods

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownload_InvalidZipRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not a zip"))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	dest := filepath.Join(tmp, "bad.zip")

	err := DownloadURLToFile(srv.URL, dest)
	if err == nil {
		t.Fatalf("expected error for invalid zip")
	}
}

func TestDownload_ValidZip(t *testing.T) {
	validZip := buildValidZip(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(validZip)
	}))
	defer srv.Close()

	tmp := t.TempDir()
	dest := filepath.Join(tmp, "good.zip")

	err := DownloadURLToFile(srv.URL, dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected file to exist after download")
	}
}

func buildValidZip(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	w, err := zw.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}
