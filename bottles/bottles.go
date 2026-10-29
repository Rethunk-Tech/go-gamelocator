// Package bottles finds the Wine bottles Bottles keeps on Linux and the game folders inside them.
package bottles

import (
	"os"
	"path/filepath"

	"github.com/Rethunk-Tech/go-desktopenv/userdirs"
	"github.com/Rethunk-Tech/go-safefs"
)

// Dirs are the folders Bottles keeps its bottles in, the ones in extra first: native, then Flatpak. home must be a
// resolved home folder, not "".
func Dirs(home string, extra ...string) []string {
	return append(append([]string(nil), extra...),
		filepath.Join(userdirs.HostDataHome(home), "bottles", "bottles"),
		filepath.Join(userdirs.FlatpakApp(home, "com.usebottles.bottles"), "data", "bottles", "bottles"),
	)
}

// List is every bottle in dir: its folders that hold bottle.yml.
func List(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, ent := range entries {
		bottle := filepath.Join(dir, ent.Name())
		if safefs.IsDir(bottle) && safefs.IsFile(filepath.Join(bottle, "bottle.yml")) {
			out = append(out, bottle)
		}
	}
	return out
}

// GameDirs are where a Steam, EA or GOG install of folder sits inside a bottle's C: drive.
func GameDirs(bottle, folder string) []string {
	c := filepath.Join(bottle, "drive_c")
	return []string{
		filepath.Join(c, "Program Files (x86)", "Steam", "steamapps", "common", folder),
		filepath.Join(c, "Program Files", "Steam", "steamapps", "common", folder),
		filepath.Join(c, "Program Files", "EA Games", folder),
		filepath.Join(c, "Program Files (x86)", "EA Games", folder),
		filepath.Join(c, "GOG Games", folder),
		filepath.Join(c, "Program Files (x86)", "GOG Galaxy", "Games", folder),
	}
}

// Of is the bottle that holds dir, found by walking up to a folder with bottle.yml; "" when dir is in none or is
// relative.
func Of(dir string) string {
	if !filepath.IsAbs(dir) {
		return ""
	}
	for d := filepath.Clean(dir); ; d = filepath.Dir(d) {
		if safefs.IsFile(filepath.Join(d, "bottle.yml")) {
			return d
		}
		if filepath.Dir(d) == d {
			return ""
		}
	}
}
