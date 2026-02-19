
package mods

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// SlugFromURL derives a stable-ish id from the URL filename.
func SlugFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "mod"
	}
	base := path.Base(u.Path)
	base = strings.TrimSuffix(base, path.Ext(base))
	base = strings.ToLower(base)
	base = nonAlnum.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		return "mod"
	}
	return base
}
