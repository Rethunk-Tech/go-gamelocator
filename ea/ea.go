// Package ea finds games installed by the EA App on Windows.
package ea

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
)

// Dirs are the folders the EA App installs games into, the ones in custom first. Only Windows has the EA App; on Linux
// an EA game is found through the Bottles or Lutris prefix that holds it.
func Dirs(custom ...string) []string { return dirs(runtime.GOOS, os.Getenv, custom...) }

func dirs(goos string, env func(string) string, custom ...string) []string {
	out := slices.Clone(custom)
	if goos != "windows" {
		return out
	}
	for _, v := range []string{"ProgramFiles", "ProgramFiles(x86)"} {
		if pf := env(v); pf != "" {
			out = append(out, filepath.Join(pf, "EA Games"), filepath.Join(pf, "Electronic Arts"))
		}
	}
	return out
}
