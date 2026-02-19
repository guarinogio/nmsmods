package app

// IsInstalledInAnyProfile returns true if the mod entry is installed in at least one profile.
// It also checks legacy v2 fields for backwards compatibility.
func IsInstalledInAnyProfile(me ModEntry) bool {
	if me.Installed {
		return true
	}
	for _, pi := range me.Installations {
		if pi.Installed {
			return true
		}
	}
	return false
}
