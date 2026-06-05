// Package fsname contains helpers for converting user defined names, like playlist
// titles, into names that are safe to use as file or directory names.
package fsname

import "strings"

// fallback is the name used when sanitising a name leaves nothing usable.
const fallback = "playlist"

// Sanitize replaces the path separator characters in a name so it cannot create
// nested directories or escape the target directory, and trims leading and trailing
// dots and spaces so traversal names like '..' are neutralised.
func Sanitize(name string) string {
	sanitized := strings.NewReplacer("/", "-", `\`, "-").Replace(name)

	sanitized = strings.Trim(sanitized, " .")
	if sanitized == "" {
		return fallback
	}

	return sanitized
}
