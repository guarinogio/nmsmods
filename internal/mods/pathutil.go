package mods

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SanitizeFolderName ensures the folder name is a safe single path segment.
//
// Policy:
// - Must not be empty.
// - Must not contain path separators.
// - Must not contain traversal segments.
// - If it contains unusual characters, it will be slugified.
//
// Returns the sanitized folder name.
func SanitizeFolderName(name string, fallback string) (string, error) {
	n := strings.TrimSpace(name)
	if n == "" {
		n = strings.TrimSpace(fallback)
	}
	if n == "" {
		return "", fmt.Errorf("empty folder name")
	}
	// Reject obvious badness up front.
	if strings.Contains(n, "/") || strings.Contains(n, "\\") {
		n = filepath.Base(strings.ReplaceAll(n, "\\", "/"))
	}
	if n == "." || n == ".." || strings.Contains(n, "..") {
		n = ""
	}
	if n == "" {
		n = strings.TrimSpace(fallback)
	}
	if n == "" {
		return "", fmt.Errorf("invalid folder name")
	}
	// Make it filesystem-friendly.
	slug := SlugFromURL(n)
	if slug == "" {
		slug = SlugFromURL(fallback)
	}
	if slug == "" {
		return "", fmt.Errorf("failed to sanitize folder name")
	}
	return slug, nil
}

// SafeJoinUnder joins base + segment and guarantees the result stays within base.
// segment must be a single path element (no separators/traversal).
func SafeJoinUnder(base string, segment string) (string, error) {
	baseClean := filepath.Clean(base)
	seg := strings.TrimSpace(segment)
	if seg == "" {
		return "", fmt.Errorf("empty path segment")
	}
	if strings.Contains(seg, "/") || strings.Contains(seg, "\\") {
		return "", fmt.Errorf("invalid path segment: %q", segment)
	}
	if seg == "." || seg == ".." || strings.Contains(seg, "..") {
		return "", fmt.Errorf("invalid path segment: %q", segment)
	}
	out := filepath.Join(baseClean, seg)
	rel, err := filepath.Rel(baseClean, out)
	if err != nil {
		return "", err
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return "", fmt.Errorf("refusing path outside base")
	}
	return out, nil
}
