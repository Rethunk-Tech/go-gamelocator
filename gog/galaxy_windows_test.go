package gog

import (
	"path/filepath"
	"testing"

	"github.com/Rethunk-Tech/go-gamelocator/internal/winregtest"
	"golang.org/x/sys/windows/registry"
)

func TestGalaxyRegistrySeam(t *testing.T) {
	dir := t.TempDir()
	writeGame(t, dir)
	lookupGalaxy = func(string) string { return dir }
	t.Cleanup(func() { lookupGalaxy = readGalaxyRegistry })
	got := Locate(t.TempDir(), testGame, Roots{})
	if len(got) == 0 || got[0].Dir != dir || got[0].Source != SourceGOG {
		t.Fatalf("galaxy = %#v", got)
	}
}

func TestWindowsOfflineProgramFiles(t *testing.T) {
	pf := t.TempDir()
	dir := filepath.Join(pf, "GOG Galaxy", "Games", "Stardew Valley")
	writeGame(t, dir)
	t.Setenv("ProgramFiles(x86)", pf)
	lookupGalaxy = func(string) string { return "" }
	t.Cleanup(func() { lookupGalaxy = readGalaxyRegistry })
	got := Locate(t.TempDir(), testGame, Roots{})
	found := false
	for _, in := range got {
		if in.Dir == dir && in.Source == SourceGOG {
			found = true
		}
	}
	if !found {
		t.Fatalf("offline program files missing: %#v", got)
	}
}

func TestGalaxyRegistryReaders(t *testing.T) {
	base := winregtest.Scratch(t, "gog-galaxy")
	winregtest.SetString(t, base+`\GalaxyClient\paths`, "client", `D:/GOG Galaxy/`)
	winregtest.SetString(t, base+`\Games\1453375253`, "path", `D:\GOG Games\Stardew Valley`)
	if got := galaxyClient(registry.CURRENT_USER, base); got != filepath.Clean(`D:/GOG Galaxy/`) {
		t.Fatalf("client = %q", got)
	}
	if got := galaxyClient(registry.CURRENT_USER, base+`\absent`); got != "" {
		t.Fatalf("absent client = %q", got)
	}
	if got := readGalaxyGame(registry.CURRENT_USER, base, "1453375253"); got != `D:\GOG Games\Stardew Valley` {
		t.Fatalf("game = %q", got)
	}
	for _, id := range []string{"", "1", `..\x`, "a/b"} {
		if got := readGalaxyGame(registry.CURRENT_USER, base, id); got != "" {
			t.Fatalf("product id %q read %q", id, got)
		}
	}
}
