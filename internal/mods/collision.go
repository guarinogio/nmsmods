package mods

import (
	"fmt"
	"strings"

	"nmsmods/internal/app"
)

// ResolveFolderCollision returns a safe destination folder.
// If another INSTALLED entry is already using the same folder, we avoid clobbering it by default.
func ResolveFolderCollision(id string, desiredFolder string, st app.State) (string, bool) {
	for otherID, me := range st.Mods {
		if otherID == id {
			continue
		}
		if me.Installed && strings.EqualFold(me.Folder, desiredFolder) && me.Folder != "" {
			return fmt.Sprintf("%s__%s", desiredFolder, id), true
		}
	}
	return desiredFolder, false
}
