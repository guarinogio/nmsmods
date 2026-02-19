package mods

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const defaultMaxDownloadBytes = int64(20) * 1024 * 1024 * 1024 // 20GB

func maxDownloadBytes() int64 {
	return defaultMaxDownloadBytes
}

func DownloadURLToFile(url, dest string) error {
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		err := downloadOnce(url, dest)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	return fmt.Errorf("download failed after retries: %w", lastErr)
}

func downloadOnce(url, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	tmp := dest + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	if resp.ContentLength > 0 && resp.ContentLength > maxDownloadBytes() {
		return fmt.Errorf("download too large: %d bytes", resp.ContentLength)
	}

	written, err := io.Copy(out, io.LimitReader(resp.Body, maxDownloadBytes()+1))
	if err != nil {
		return err
	}
	if written > maxDownloadBytes() {
		return errors.New("download exceeded max allowed size")
	}

	if err := out.Close(); err != nil {
		return err
	}

	// Validate it's actually a zip
	if _, err := zip.OpenReader(tmp); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("downloaded file is not a valid zip")
	}

	return os.Rename(tmp, dest)
}
