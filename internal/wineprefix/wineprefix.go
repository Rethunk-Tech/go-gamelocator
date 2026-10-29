// Package wineprefix finds the folder of a Wine prefix that holds drive_c.
package wineprefix

import (
	"path/filepath"

	"github.com/Rethunk-Tech/go-safefs"
)

// Root is the folder holding drive_c for the prefix a launcher config names: the prefix's pfx folder when the runner
// is Proton, else the prefix itself. It is "" when the prefix is empty or neither holds drive_c.
func Root(prefix string) string {
	if prefix == "" {
		return ""
	}
	for _, p := range []string{filepath.Join(prefix, "pfx"), prefix} {
		if safefs.IsDir(filepath.Join(p, "drive_c")) {
			return p
		}
	}
	return ""
}
