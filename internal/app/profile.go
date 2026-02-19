package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var profileNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`)

// ActiveProfile returns the current profile, defaulting to "default".
func ActiveProfile(c Config) string {
	if c.ActiveProfile == "" {
		return "default"
	}
	return c.ActiveProfile
}

// ValidateProfileName ensures a profile name is safe to use as a directory name.
func ValidateProfileName(name string) error {
	if !profileNameRe.MatchString(name) {
		return fmt.Errorf("invalid profile name %q (allowed: letters, digits, ., _, -, max 64 chars)", name)
	}
	return nil
}

// ProfileRoot returns the root directory for a profile under the app root.
func ProfileRoot(p *Paths, profile string) string {
	return filepath.Join(p.Profiles, profile)
}

// ProfileModsDir returns the per-profile mods storage (authoritative store).
func ProfileModsDir(p *Paths, profile string) string {
	return filepath.Join(ProfileRoot(p, profile), "mods")
}

func EnsureProfileDirs(p *Paths, profile string) error {
	if err := ValidateProfileName(profile); err != nil {
		return err
	}
	return os.MkdirAll(ProfileModsDir(p, profile), 0o755)
}
