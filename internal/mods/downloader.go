
package mods

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func DownloadToFile(rawURL, destPath string) (int64, error) {
	if rawURL == "" {
		return 0, errors.New("url is empty")
	}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return 0, err
	}
	// A reasonable UA can help some CDNs; keep it simple.
	req.Header.Set("User-Agent", "nmsmods/0.1 (+personal)")

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("download failed: %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return 0, err
	}
	tmp := destPath + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return 0, err
	}
	defer func() {
		f.Close()
	}()

	// Simple progress: print bytes every ~1s
	var written int64
	buf := make([]byte, 32*1024)
	last := time.Now()
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			wn, werr := f.Write(buf[:n])
			if werr != nil {
				return written, werr
			}
			written += int64(wn)
			if time.Since(last) > time.Second {
				fmt.Printf("\rDownloaded %.2f MB", float64(written)/(1024*1024))
				last = time.Now()
			}
		}
		if rerr != nil {
			if errors.Is(rerr, io.EOF) {
				break
			}
			return written, rerr
		}
	}
	fmt.Printf("\rDownloaded %.2f MB\n", float64(written)/(1024*1024))

	if err := f.Close(); err != nil {
		return written, err
	}
	return written, os.Rename(tmp, destPath)
}
