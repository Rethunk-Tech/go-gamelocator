//go:build !windows

package steam

import (
	"path/filepath"
	"runtime"
)

func candidates(home string) []string { return unixCandidates(runtime.GOOS, home) }

// unixCandidates takes goos so the macOS folder is tested on Linux.
func unixCandidates(goos, home string) []string {
	if goos == "darwin" {
		return []string{filepath.Join(home, "Library", "Application Support", "Steam")}
	}
	return []string{
		filepath.Join(home, ".local", "share", "Steam"),
		filepath.Join(home, ".steam", "steam"),
		filepath.Join(home, ".steam", "root"),
		filepath.Join(home, ".steam", "debian-installation"),
		filepath.Join(home, "snap", "steam", "common", ".local", "share", "Steam"),
	}
}
