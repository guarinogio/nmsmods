package nexus

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// NXM represents a parsed nxm:// URL, as used by Nexus "Mod Manager Download".
//
// Example:
//   nxm://nomanssky/mods/3718/files/43996?key=...&expires=...&user_id=...
type NXM struct {
	GameDomain string
	ModID      int
	FileID     int

	Key     string
	Expires string
	UserID  string
}

func ParseNXM(raw string) (*NXM, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "nxm" {
		return nil, fmt.Errorf("invalid scheme %q (expected nxm)", u.Scheme)
	}

	game := u.Host
	if game == "" {
		return nil, fmt.Errorf("missing game domain host")
	}

	// path: /mods/{mod_id}/files/{file_id}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "mods" || parts[2] != "files" {
		return nil, fmt.Errorf("invalid nxm path %q (expected /mods/<mod_id>/files/<file_id>)", u.Path)
	}

	modID, err := strconv.Atoi(parts[1])
	if err != nil || modID <= 0 {
		return nil, fmt.Errorf("invalid mod_id %q", parts[1])
	}
	fileID, err := strconv.Atoi(parts[3])
	if err != nil || fileID <= 0 {
		return nil, fmt.Errorf("invalid file_id %q", parts[3])
	}

	q := u.Query()
	n := &NXM{
		GameDomain: game,
		ModID:      modID,
		FileID:     fileID,
		Key:        q.Get("key"),
		Expires:    q.Get("expires"),
		UserID:     q.Get("user_id"),
	}

	// These are required for download_link.json.
	if n.Key == "" || n.Expires == "" || n.UserID == "" {
		return nil, fmt.Errorf("missing required query params (need key, expires, user_id)")
	}

	return n, nil
}
