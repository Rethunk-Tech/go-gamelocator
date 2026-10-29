package steam

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mkSteam(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "steamapps"), 0o750); err != nil {
		t.Fatal(err)
	}
}

// Roots the player added are searched before the usual folders, and are the only way to find Steam on Windows when
// the registry has no entry.
func TestPlayerAddedRootsAreSearchedFirst(t *testing.T) {
	home := t.TempDir()
	added := filepath.Join(t.TempDir(), "Games", "Steam")
	notSteam := filepath.Join(t.TempDir(), "empty")
	mkSteam(t, added)
	if err := os.MkdirAll(notSteam, 0o750); err != nil {
		t.Fatal(err)
	}
	if roots := Roots(home, added); roots[0] != added {
		t.Fatalf("roots = %v", roots)
	}
	all := LocateAll(home, notSteam, added)
	if len(all) == 0 || all[0].Root != added || all[0].Kind != KindNative {
		t.Fatalf("all = %#v", all)
	}
	for _, s := range all {
		if s.Root == notSteam {
			t.Fatalf("a folder without steamapps is not a Steam: %#v", all)
		}
	}
	if s, st := Locate(home, added); st != Found || s.Root != added {
		t.Fatalf("locate = %#v %s", s, st)
	}
	if _, st := Locate(home, notSteam); st == Found {
		t.Fatalf("found a Steam in %s", notSteam)
	}
}

func TestExtraRootsFeedProtonBuildsAndTheirLibraries(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(t.TempDir(), "Steam")
	library := filepath.Join(t.TempDir(), "Library")
	mkSteam(t, library)
	vdfPath := `"libraryfolders" { "0" { "path" "` + strings.ReplaceAll(library, `\`, `\\`) + `" } }`
	writeAt(t, filepath.Join(root, "steamapps", "libraryfolders.vdf"), []byte(vdfPath))
	for _, p := range []string{
		filepath.Join(root, "compatibilitytools.d", "GE-Proton10-34", "proton"),
		filepath.Join(library, "steamapps", "common", "Proton 9.0", "proton"),
		filepath.Join(library, "steamapps", "common", "Not Proton", "readme"),
	} {
		writeAt(t, p, []byte("#!/bin/sh\n"))
	}
	got := ProtonBuilds(home, root)
	if len(got) != 2 || filepath.Base(got[0].Dir) != "Proton 9.0" || filepath.Base(got[1].Dir) != "GE-Proton10-34" || got[0].SteamRoot != root {
		t.Fatalf("builds = %#v", got)
	}
	if b, ok := FindProton(home, "GE-Proton10-34", root); !ok || filepath.Base(b.Dir) != "GE-Proton10-34" {
		t.Fatalf("named build = %#v %v", b, ok)
	}
	if _, ok := FindProton(home, "", filepath.Join(t.TempDir(), "none")); ok && len(ProtonBuilds(home)) == 0 {
		t.Fatal("found a build with no Steam")
	}
}
