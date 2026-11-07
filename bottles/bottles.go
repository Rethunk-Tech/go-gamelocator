// Package bottles finds the Wine bottles Bottles keeps on Linux and the game folders inside them.
package bottles

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Rethunk-Tech/go-desktopenv/userdirs"
	"github.com/Rethunk-Tech/go-gamelocator/internal/readcap"
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

// UserDir is the Windows profile folder inside a bottle: its one real user's, else Wine's default steamuser.
func UserDir(bottle string) string {
	users := filepath.Join(bottle, "drive_c", "users")
	if entries, err := os.ReadDir(users); err == nil {
		for _, e := range entries {
			if e.IsDir() && e.Name() != "Public" {
				return filepath.Join(users, e.Name())
			}
		}
	}
	return filepath.Join(users, "steamuser")
}

var reName = regexp.MustCompile(`(?m)^Name:\s*(.+?)\s*$`)

// Name is the name bottles-cli knows the bottle by: bottle.yml's Name, else its folder's name.
func Name(bottle string) string {
	if b, err := readcap.File(filepath.Join(bottle, "bottle.yml"), readcap.Config); err == nil {
		if m := reName.FindSubmatch(b); m != nil {
			if n := strings.Trim(string(m[1]), `"'`); n != "" {
				return n
			}
		}
	}
	return filepath.Base(bottle)
}

// IsFlatpak reports whether the bottle belongs to Flatpak Bottles, whose CLI is not on the host's PATH.
func IsFlatpak(bottle string) bool {
	return strings.Contains(filepath.ToSlash(bottle), "/.var/app/com.usebottles.bottles/")
}

// RunArgv is the command line that runs exe with args inside bottle through bottles-cli, or through flatpak run for
// a Flatpak bottle.
func RunArgv(bottle, exe string, args ...string) []string {
	argv := []string{"bottles-cli"}
	if IsFlatpak(bottle) {
		argv = []string{"flatpak", "run", "--command=bottles-cli", "com.usebottles.bottles"}
	}
	argv = append(argv, "run", "-b", Name(bottle), "-e", exe)
	if len(args) > 0 {
		argv = append(argv, "--")
		argv = append(argv, args...)
	}
	return argv
}
