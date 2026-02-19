
package mods

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractZip extracts a zip file into destDir.
// Personal v0.1: minimal safety checks; v1.0 should harden against traversal.
func ExtractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	for _, f := range r.File {
		name := f.Name
		// avoid absolute paths
		name = strings.TrimPrefix(name, "/")
		outPath := filepath.Join(destDir, name)
		// ensure within destDir (basic)
		if !strings.HasPrefix(filepath.Clean(outPath), filepath.Clean(destDir)) {
			return errors.New("zip entry would write outside destination")
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		of, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(of, rc); err != nil {
			of.Close()
			rc.Close()
			return err
		}
		of.Close()
		rc.Close()
	}
	return nil
}
