package mods

import (
	"fmt"
	"strings"

	"nmsmods/internal/app"
)

// ResolveFolderCollision returns a safe destination folder.
// If another entry is already using the same folder *within the given profile*, we avoid clobbering it by default.
func ResolveFolderCollision(id string, desiredFolder string, profile string, st app.State) (string, bool) {
	for otherID, me := range st.Mods {
		if otherID == id {
			continue
		}
		pi, ok := me.Installations[profile]
		if ok && pi.Installed && strings.EqualFold(pi.Folder, desiredFolder) && pi.Folder != "" {
			return fmt.Sprintf("%s__%s", desiredFolder, id), true
		}
		// Back-compat: if state was not migrated for some reason.
		if me.Installed && strings.EqualFold(me.Folder, desiredFolder) && me.Folder != "" {
			return fmt.Sprintf("%s__%s", desiredFolder, id), true
		}
	}
	return desiredFolder, false
}
