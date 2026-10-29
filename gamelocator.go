package gamelocator

import (
	"path/filepath"

	"github.com/Rethunk-Tech/go-safefs"
)

// GameDir is dir when it holds marker (a slash path below dir, so "Game/Bin/x.exe" works), else its "game" subfolder
// when that does (GOG's offline layout), else "".
func GameDir(dir, marker string) string {
	if dir == "" {
		return ""
	}
	dir = filepath.Clean(dir)
	for _, d := range []string{dir, filepath.Join(dir, "game")} {
		if safefs.IsFile(filepath.Join(d, marker)) {
			return d
		}
	}
	return ""
}
