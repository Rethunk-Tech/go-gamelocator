//go:build windows

package steam

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/Rethunk-Tech/go-gamelocator/internal/winregtest"
	"golang.org/x/sys/windows/registry"
)

func TestCandidatesReadTheSteamKeys(t *testing.T) {
	base := winregtest.Scratch(t, "steam-candidates")
	winregtest.SetString(t, base+`\user\Valve\Steam`, "SteamPath", `c:/Program Files (x86)/Steam`)
	winregtest.SetString(t, base+`\machine\Wow6432Node\Valve\Steam`, "InstallPath", `D:\Steam`)
	winregtest.SetString(t, base+`\machine\Valve\Steam`, "InstallPath", `E:\Steam`)
	got := candidatesAt(registry.CURRENT_USER, registry.CURRENT_USER, base+`\user`, base+`\machine`)
	want := []string{filepath.Clean(`c:/Program Files (x86)/Steam`), `D:\Steam`, `E:\Steam`}
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := candidatesAt(registry.CURRENT_USER, registry.CURRENT_USER, base+`\none`, base+`\none`); len(got) != 0 {
		t.Fatalf("absent keys gave %q", got)
	}
}
