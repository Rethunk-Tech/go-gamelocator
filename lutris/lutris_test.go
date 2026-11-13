package lutris

import (
	_ "embed"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/Rethunk-Tech/go-safefs"
)

//go:embed testdata/gog-linux.yml
var fixtureGogLinux []byte

//go:embed testdata/steam-runner.yml
var fixtureSteamRunner []byte

//go:embed testdata/wine-c-drive.yml
var fixtureWineCDrive []byte

//go:embed testdata/wine-relative.yml
var fixtureWineRelative []byte

//go:embed testdata/wine-z-drive.yml
var fixtureWineZDrive []byte

//go:embed testdata/wine-c-escape.yml
var fixtureWineCEscape []byte

//go:embed testdata/slug-only.yml
var fixtureSlugOnly []byte

// testGame is a game the way a Game implementation describes it to the Lutris locator.
var testGame = Game{Slug: "stardew-valley", Keyword: "stardew", Marker: "Stardew Valley.dll"}

func TestInstallFromFixture(t *testing.T) {
	dir := t.TempDir()
	if err := safefs.WriteFile(filepath.Join(dir, "Stardew Valley.dll"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	replaced := strings.ReplaceAll(string(fixtureGogLinux), "/home/player/Games/Stardew Valley", dir)
	got := newMatcher(testGame).installFromYAML(replaced)
	if got != dir {
		t.Fatalf("install = %q", got)
	}
	steam := newMatcher(testGame).installFromYAML(string(fixtureSteamRunner))
	if steam != "" {
		t.Fatalf("steam runner = %q", steam)
	}
}

func TestInstallFromWineWindowsPaths(t *testing.T) {
	dir := t.TempDir()
	prefix := filepath.Join(dir, "prefix")
	game := filepath.Join(prefix, "drive_c", "GOG Games", "Stardew Valley")
	outside := filepath.Join(dir, "outside")
	unix := filepath.Join(dir, "unix")
	for _, p := range []string{game, outside, unix} {
		if err := os.MkdirAll(p, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := safefs.WriteFile(filepath.Join(p, "Stardew Valley.dll"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	m := newMatcher(testGame)
	fill := func(body []byte, prefixVal, unixWin string) string {
		s := strings.ReplaceAll(string(body), "PREFIX", prefixVal)
		return strings.ReplaceAll(s, "UNIXWIN", unixWin)
	}
	if got := m.installFromYAML(fill(fixtureWineCDrive, prefix, "")); got != game {
		t.Fatalf("C: = %q, want %q", got, game)
	}
	if got := m.installFromYAML(fill(fixtureWineRelative, prefix, "")); got != game {
		t.Fatalf("relative = %q, want %q", got, game)
	}
	z := strings.ReplaceAll(unix, `/`, `\`)
	if got := m.installFromYAML(fill(fixtureWineZDrive, prefix, z)); got != unix {
		t.Fatalf("Z: = %q, want %q", got, unix)
	}
	if got := m.installFromYAML(fill(fixtureWineCEscape, prefix, "")); got != "" {
		t.Fatalf("escaped C: = %q", got)
	}
}

func TestLocateUsesTheGameSlugWhenTheConfigOmitsGameSlug(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("lutris discovery is Linux-only")
	}
	home := t.TempDir()
	games := filepath.Join(home, ".local", "share", "lutris", "games")
	if err := os.MkdirAll(games, 0o700); err != nil {
		t.Fatal(err)
	}
	install := filepath.Join(home, "sdv")
	if err := os.MkdirAll(install, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := safefs.WriteFile(filepath.Join(install, "Stardew Valley.dll"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	yml := strings.ReplaceAll(string(fixtureSlugOnly), "DIR", install)
	if err := safefs.WriteFile(filepath.Join(games, "stardew.yml"), []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}
	found := Locate(home, testGame)
	if len(found) != 1 || found[0].Dir != install || found[0].Slug != testGame.Slug {
		t.Fatalf("Locate = %+v", found)
	}
}

func TestLocateLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("lutris discovery is Linux-only")
	}
	home := t.TempDir()
	games := filepath.Join(home, ".local", "share", "lutris", "games")
	if err := os.MkdirAll(games, 0o700); err != nil {
		t.Fatal(err)
	}
	install := filepath.Join(home, "sdv")
	if err := os.MkdirAll(install, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := safefs.WriteFile(filepath.Join(install, "Stardew Valley.dll"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	yml := "game_slug: stardew-valley\nscript:\n  runner: linux\ngame:\n  working_dir: " + install + "\n"
	if err := safefs.WriteFile(filepath.Join(games, "stardew.yml"), []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}
	found := Locate(home, testGame)
	if len(found) != 1 || found[0].Dir != install {
		t.Fatalf("Locate = %+v", found)
	}
}

func TestConfigDirsFollowCustomXDGHomes(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Lutris is Linux-only")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // userdirs honours XDG and APPDATA only for os.UserHomeDir, which reads USERPROFILE on Windows
	cfg, data := filepath.Join(t.TempDir(), "cfg"), filepath.Join(t.TempDir(), "data")
	want := []string{filepath.Join(data, "lutris", "games"), filepath.Join(cfg, "lutris", "games")}
	check := func(t *testing.T) {
		t.Helper()
		got := ConfigDirs(home)
		for _, w := range want {
			if !slices.Contains(got, w) {
				t.Fatalf("%q missing from %v", w, got)
			}
		}
	}
	t.Run("native", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", cfg)
		t.Setenv("XDG_DATA_HOME", data)
		check(t)
	})
	t.Run("flatpak", func(t *testing.T) {
		t.Setenv("FLATPAK_ID", "test.app")
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "own", "config"))
		t.Setenv("XDG_DATA_HOME", filepath.Join(home, "own", "data"))
		t.Setenv("HOST_XDG_CONFIG_HOME", cfg)
		t.Setenv("HOST_XDG_DATA_HOME", data)
		check(t)
	})
}

// Without the slug, a config matches only when its keyword also names an exe; the install is then found from the exe
// or, failing the working dir, from the folder the exe sits in.
func TestInstallFromYAMLMatchingAndResolving(t *testing.T) {
	dir := t.TempDir()
	if err := safefs.WriteFile(filepath.Join(dir, "Stardew Valley.dll"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	m := newMatcher(testGame)
	for name, c := range map[string]struct{ yml, want string }{
		"exe names the keyword":        {"game:\n  exe: " + dir + "/stardew.exe\n", dir},
		"working dir is gone":          {"game:\n  working_dir: " + dir + "/gone\n  exe: " + dir + "/stardew.exe\n", dir},
		"keyword only in a comment":    {"# stardew\ngame:\n  exe: " + dir + "/other.exe\n", ""},
		"no keyword anywhere":          {"game:\n  exe: " + dir + "/other.exe\n", ""},
		"keyword exe in a missing dir": {"game:\n  exe: " + dir + "/gone/stardew.exe\n", ""},
	} {
		if got := m.installFromYAML(c.yml); got != c.want {
			t.Errorf("%s: %q, want %q", name, got, c.want)
		}
	}
	if _, _, err := m.installFromConfig(dir, "absent.yml"); err == nil {
		t.Error("an absent config read as empty")
	}
}
