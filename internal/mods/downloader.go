package mods

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadURLToFile downloads a remote URL into a local file path.
// It writes to dest + ".part" first, then renames atomically.
func DownloadURLToFile(url, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	tmp := dest + ".part"

	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{
		Timeout: 0, // allow large downloads (no fixed timeout)
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	// Stream copy
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	// Rename atomically
	if err := os.Rename(tmp, dest); err != nil {
		return err
	}

	// Update mod time to now
	now := time.Now()
	_ = os.Chtimes(dest, now, now)

	return nil
}
