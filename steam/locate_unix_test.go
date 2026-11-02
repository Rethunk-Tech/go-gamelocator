//go:build !windows

package steam

import (
	"path/filepath"
	"testing"
)

func TestMacSteamFolderIsListed(t *testing.T) {
	want := filepath.Join("h", "Library", "Application Support", "Steam")
	if got := unixCandidates("darwin", "h"); len(got) != 1 || got[0] != want {
		t.Fatalf("darwin = %v", got)
	}
	if got := unixCandidates("linux", "h"); filepath.Join("h", ".local", "share", "Steam") != got[0] {
		t.Fatalf("linux = %v", got)
	}
}
