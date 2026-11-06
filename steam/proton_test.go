package steam

import (
	"path/filepath"
	"slices"
	"testing"
)

// Builds outside Valve's numbered releases rank by the numbers in their names, so GE-Proton10 is newer than GE-Proton9.
func TestProtonRankOrdersOtherBuildsByTheirNumbers(t *testing.T) {
	got := []string{"GE-Proton9-27", "GE-Proton10-9", "GE-Proton10-10"}
	slices.SortFunc(got, protonRank)
	if want := []string{"GE-Proton10-10", "GE-Proton10-9", "GE-Proton9-27"}; !slices.Equal(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
	got = []string{"GE-Proton10-34", "Proton - Experimental", "Proton 8.0", "Proton 9.0 (Beta)"}
	slices.SortFunc(got, protonRank)
	if want := []string{"Proton 9.0 (Beta)", "Proton 8.0", "Proton - Experimental", "GE-Proton10-34"}; !slices.Equal(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestPickProtonMatchesSteamsToolBeforeTheBest(t *testing.T) {
	found := []ProtonBuild{{Dir: "/c/GE-Proton10-34"}, {Dir: "/c/Proton 9.0 (Beta)"}, {Dir: "/c/Proton - Experimental"}}
	for tool, want := range map[string]string{"proton_9": "Proton 9.0 (Beta)", "proton_experimental": "Proton - Experimental", "GE-Proton10-34": "GE-Proton10-34", "": "GE-Proton10-34", "proton_8": "GE-Proton10-34"} {
		if b, ok := PickProton(found, "", tool); !ok || filepath.Base(b.Dir) != want {
			t.Errorf("tool %q picked %v, want %s", tool, b.Dir, want)
		}
	}
	if b, _ := PickProton(found, "Proton - Experimental", "proton_9"); filepath.Base(b.Dir) != "Proton - Experimental" {
		t.Errorf("a named build lost to the tool: %v", b.Dir)
	}
	if _, ok := PickProton(nil, "x", "y"); ok {
		t.Error("picked from nothing")
	}
}

func TestLibraryOfWalksUpToSteamapps(t *testing.T) {
	root := filepath.Join(t.TempDir(), "lib")
	dir := filepath.Join(root, "steamapps", "common", "Game", "Bin")
	if got, ok := LibraryOf(dir); !ok || got != filepath.Join(root, "steamapps") {
		t.Fatalf("got %q %v", got, ok)
	}
	if _, ok := LibraryOf(root); ok {
		t.Fatal("a folder outside any library")
	}
}

func TestPrefixVersionIsConfigInfosFirstLine(t *testing.T) {
	prefix := t.TempDir()
	if got := PrefixVersion(prefix); got != "" {
		t.Fatalf("no file: %q", got)
	}
	writeAt(t, filepath.Join(prefix, "config_info"), []byte("9.0-4\nother\n"))
	if got := PrefixVersion(prefix); got != "9.0-4" {
		t.Fatalf("got %q", got)
	}
	writeAt(t, filepath.Join(prefix, "config_info"), []byte("\nother\n"))
	if got := PrefixVersion(prefix); got != "" {
		t.Fatalf("blank first line: %q", got)
	}
}

// The mapping is a flat list of {name, config, priority} blocks under Software/Valve/Steam, whose keys Steam writes
// in either case.
func TestCompatToolIsTheExactEntryWithNoDefaultFallback(t *testing.T) {
	root := t.TempDir()
	s := Steam{Root: root}
	if got := s.CompatTool("1966720"); got != "" {
		t.Fatalf("no config.vdf: %q", got)
	}
	writeAt(t, filepath.Join(root, "config", "config.vdf"), []byte(`"InstallConfigStore"
{
	"software"
	{
		"Valve"
		{
			"Steam"
			{
				"CompatToolMapping"
				{
					"0"
					{
						"name"		"proton_9"
						"config"		""
						"priority"		"75"
					}
					"1966720"
					{
						"name"		"GE-Proton9-5"
						"config"		""
						"priority"		"250"
					}
				}
			}
		}
	}
}
`))
	if got := s.CompatTool("1966720"); got != "GE-Proton9-5" {
		t.Fatalf("per-game entry: %q", got)
	}
	if got := s.CompatTool("0"); got != "proton_9" {
		t.Fatalf("all-games entry: %q", got)
	}
	if got := s.CompatTool("42"); got != "" {
		t.Fatalf("an app with no entry must not take the default: %q", got)
	}
}

// Steam keeps compatdata in the library of the app's default install: the game's own library wins, then any library
// of any root, and with none the game's own is where Steam will create it.
func TestCompatDataSearchesEveryLibrary(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "steam")
	second := filepath.Join(base, "second")
	own := filepath.Join(base, "own", "steamapps")
	writeAt(t, filepath.Join(root, "steamapps", "libraryfolders.vdf"),
		[]byte(`"libraryfolders" { "0" { "path" "`+root+`" } "1" { "path" "`+second+`" } }`))
	st := Steam{Root: root}

	if got, want := CompatData("42", own, st), filepath.Join(own, "compatdata", "42"); got != want {
		t.Fatalf("none yet = %q, want %q", got, want)
	}
	writeAt(t, filepath.Join(second, "steamapps", "compatdata", "42", "config_info"), []byte("x"))
	if got, want := CompatData("42", own, st), filepath.Join(second, "steamapps", "compatdata", "42"); got != want {
		t.Fatalf("other library = %q, want %q", got, want)
	}
	writeAt(t, filepath.Join(own, "compatdata", "42", "config_info"), []byte("x"))
	if got, want := CompatData("42", own, st), filepath.Join(own, "compatdata", "42"); got != want {
		t.Fatalf("own library first = %q, want %q", got, want)
	}
	if _, ok := st.CompatData("../x"); ok {
		t.Fatal("an id that escapes the folder")
	}
	if got, want := ProtonUserDir("/c"), filepath.Join("/c", "pfx", "drive_c", "users", "steamuser"); got != want {
		t.Fatalf("ProtonUserDir = %q", got)
	}
}
