package mods

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Defaults tuned for "real mods can be big" while still preventing obvious abuse.
// You can override these with env vars:
//   NMSMODS_MAX_FILE_BYTES   (default 8 GiB)
//   NMSMODS_MAX_TOTAL_BYTES  (default 50 GiB)
//   NMSMODS_MAX_FILES        (default 20000)
const (
	defaultMaxFileBytes  = int64(8) * 1024 * 1024 * 1024  // 8 GiB
	defaultMaxTotalBytes = int64(50) * 1024 * 1024 * 1024 // 50 GiB
	defaultMaxFiles      = 20000
)

type extractLimits struct {
	maxFileBytes  int64
	maxTotalBytes int64
	maxFiles      int
}

func readLimitsFromEnv() extractLimits {
	lim := extractLimits{
		maxFileBytes:  defaultMaxFileBytes,
		maxTotalBytes: defaultMaxTotalBytes,
		maxFiles:      defaultMaxFiles,
	}

	if v := strings.TrimSpace(os.Getenv("NMSMODS_MAX_FILE_BYTES")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			lim.maxFileBytes = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("NMSMODS_MAX_TOTAL_BYTES")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			lim.maxTotalBytes = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("NMSMODS_MAX_FILES")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			lim.maxFiles = n
		}
	}

	return lim
}

func isTraversal(cleanRel string) bool {
	if cleanRel == ".." {
		return true
	}
	prefix := ".." + string(os.PathSeparator)
	return strings.HasPrefix(cleanRel, prefix)
}

func looksLikeWindowsAbsPath(s string) bool {
	// e.g. "C:\foo" or "C:/foo" (zip often uses forward slashes)
	if len(s) >= 2 && s[1] == ':' {
		return true
	}
	return false
}

func sanitizeZipName(name string) (string, error) {
	// IMPORTANT: reject absolute-like entries BEFORE normalizing away the leading slash.
	// ZIP entries can start with "/" or "\" to try to escape.
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "\\") {
		return "", fmt.Errorf("absolute path not allowed: %s", name)
	}
	if looksLikeWindowsAbsPath(name) {
		return "", fmt.Errorf("absolute path not allowed: %s", name)
	}

	// ZIP names are spec'd with forward slashes; attackers may include backslashes.
	name = strings.ReplaceAll(name, "\\", "/")

	if name == "" || name == "." {
		return "", errors.New("empty zip entry name")
	}

	// Convert to OS path separators then clean.
	rel := filepath.FromSlash(name)
	rel = filepath.Clean(rel)

	// Reject absolute after cleaning.
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute path not allowed: %s", name)
	}
	// Reject traversal after cleaning.
	if isTraversal(rel) {
		return "", fmt.Errorf("path traversal not allowed: %s", name)
	}

		for _, r := range rel {
		if r < 32 {
			return "", fmt.Errorf("invalid control character in path")
		}
	}
	
	if strings.Contains(rel, ":") {
		return "", fmt.Errorf("invalid character ':' in path")
	}

	return rel, nil
}

// ExtractZip extracts a ZIP file into destDir securely.
// Security properties:
// - rejects absolute paths and traversal
// - rejects symlinks
// - enforces max files, max per-file size, and max total extracted bytes (configurable via env vars)
func ExtractZip(zipPath, destDir string) error {
	lim := readLimitsFromEnv()

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if len(r.File) > lim.maxFiles {
		return fmt.Errorf("zip contains too many files (%d > %d)", len(r.File), lim.maxFiles)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	destRootClean := filepath.Clean(destDir)
	if !strings.HasSuffix(destRootClean, string(os.PathSeparator)) {
		destRootClean += string(os.PathSeparator)
	}

	var total int64

	for _, f := range r.File {
		// Reject symlinks (based on mode bits in the ZIP header)
		if f.FileInfo().Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks not allowed: %s", f.Name)
		}

		rel, err := sanitizeZipName(f.Name)
		if err != nil {
			return err
		}

		outPath := filepath.Join(destDir, rel)
		outClean := filepath.Clean(outPath)

		// Ensure extraction stays inside destDir
		if !strings.HasPrefix(outClean+string(os.PathSeparator), destRootClean) && outClean != filepath.Clean(destDir) {
			return fmt.Errorf("illegal path outside destination: %s", outPath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, 0o755); err != nil {
				return err
			}
			continue
		}

		// Enforce declared uncompressed size when present
		if f.UncompressedSize64 > 0 && int64(f.UncompressedSize64) > lim.maxFileBytes {
			return fmt.Errorf("file too large: %s (%d bytes > %d)", f.Name, f.UncompressedSize64, lim.maxFileBytes)
		}

		// Track totals (use declared size; additionally enforce via LimitReader during copy)
		if f.UncompressedSize64 > 0 {
			total += int64(f.UncompressedSize64)
			if total > lim.maxTotalBytes {
				return fmt.Errorf("zip exceeds maximum total size (%d > %d bytes)", total, lim.maxTotalBytes)
			}
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		dst, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			rc.Close()
			return err
		}

		// Hard cap actual bytes written regardless of header claims
		written, err := io.Copy(dst, io.LimitReader(rc, lim.maxFileBytes+1))
		dst.Close()
		rc.Close()

		if err != nil {
			return err
		}
		if written > lim.maxFileBytes {
			return fmt.Errorf("file too large while extracting (>%d bytes): %s", lim.maxFileBytes, f.Name)
		}

		// If header didn't provide size, count actual bytes
		if f.UncompressedSize64 == 0 {
			total += written
			if total > lim.maxTotalBytes {
				return fmt.Errorf("zip exceeds maximum total size (%d > %d bytes)", total, lim.maxTotalBytes)
			}
		}
	}

	return nil
}
