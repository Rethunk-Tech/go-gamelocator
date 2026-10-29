package heroic

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"testing"

	"github.com/Rethunk-Tech/go-gamelocator/epic"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestConfigDirsFollowCustomXDGConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Heroic reads %APPDATA% on Windows")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // userdirs honours XDG only for os.UserHomeDir, which reads USERPROFILE on Windows
	cfg := filepath.Join(t.TempDir(), "cfg")
	want := filepath.Join(cfg, "heroic")
	t.Run("native", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", cfg)
		if got := ConfigDirs(home); !slices.Contains(got, want) {
			t.Fatalf("%q missing from %v", want, got)
		}
	})
	t.Run("flatpak", func(t *testing.T) {
		t.Setenv("FLATPAK_ID", "test.app")
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "own", "config"))
		t.Setenv("HOST_XDG_CONFIG_HOME", cfg)
		if got := ConfigDirs(home); !slices.Contains(got, want) {
			t.Fatalf("%q missing from %v", want, got)
		}
	})
}

func TestConfigDirsListTheUsersFoldersFirst(t *testing.T) {
	home := t.TempDir()
	got := ConfigDirs(home, "/mine")
	if got[0] != "/mine" || len(got) < 2 {
		t.Fatalf("got %v", got)
	}
}

func TestConfigDirsOnWindowsReadAppDataUnderTheHomeWhenTheVariableIsUnset(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("%APPDATA% is Windows'")
	}
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	t.Setenv("APPDATA", "")
	if got, want := ConfigDirs(home), filepath.Join(home, "AppData", "Roaming", "heroic"); !slices.Contains(got, want) {
		t.Fatalf("%q missing from %v", want, got)
	}
}

func TestGOGInstallsReadsAllThreeShapes(t *testing.T) {
	const id = "1453375253"
	for name, body := range map[string]string{
		"object with installed": `{"installed":[{"appName":"` + id + `","install_path":"/a"},{"appName":"1","install_path":"/nope"}]}`,
		"object with games":     `{"games":[{"appName":"` + id + `","install_path":"/a"}]}`,
		"list":                  `[{"appName":"` + id + `","install_path":"/a"}]`,
		"keyed":                 `{"` + id + `":{"install_path":"/a"}}`,
	} {
		cfg := t.TempDir()
		writeFile(t, filepath.Join(cfg, "gog_store", "installed.json"), body)
		if got := GOGInstalls(cfg, id); !slices.Equal(got, []string{"/a"}) {
			t.Errorf("%s: %v", name, got)
		}
	}
	cfg := t.TempDir()
	if got := GOGInstalls(cfg, id); got != nil {
		t.Errorf("no file: %v", got)
	}
	writeFile(t, filepath.Join(cfg, "gog_store", "installed.json"), "junk")
	if got := GOGInstalls(cfg, id); got != nil {
		t.Errorf("junk: %v", got)
	}
}

func TestEpicInstallsReadsLegendarysList(t *testing.T) {
	cfg := t.TempDir()
	if got := EpicInstalls(cfg); len(got) != 0 {
		t.Fatalf("no list: %v", got)
	}
	writeFile(t, EpicList(cfg), fmt.Sprintf(`{"Garlic":{"app_name":"Garlic","install_path":%q,"platform":"Windows"}}`, "/games/Garlic"))
	if got, want := EpicInstalls(cfg), []epic.Install{{AppName: "Garlic", Dir: "/games/Garlic"}}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// The GamesConfig shape is a real Heroic 2.x file's.
func TestPrefixIsThePfxFolderOfAProtonPrefix(t *testing.T) {
	dir := t.TempDir()
	prefix := filepath.Join(dir, "Prefixes", "Game")
	for _, p := range []string{filepath.Join(prefix, "pfx", "drive_c"), filepath.Join(dir, "Wine", "drive_c"), filepath.Join(dir, "GamesConfig")} {
		if err := os.MkdirAll(p, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	write := func(app, body string) { writeFile(t, filepath.Join(dir, "GamesConfig", app+".json"), body) }
	write("proton1", `{"proton1":{"wineVersion":{"type":"proton"},"winePrefix":`+strconv.Quote(prefix)+`},"version":"v0"}`)
	write("wine1", `{"wine1":{"wineVersion":{"type":"wine"},"winePrefix":`+strconv.Quote(filepath.Join(dir, "Wine"))+`}}`)
	write("gone", `{"gone":{"winePrefix":`+strconv.Quote(filepath.Join(dir, "missing"))+`}}`)
	for app, want := range map[string]string{"proton1": filepath.Join(prefix, "pfx"), "wine1": filepath.Join(dir, "Wine"), "gone": "", "none": "", "../x": "", "a/b": "", `a\b`: "", "": "", "..": ""} {
		if got := Prefix(dir, app); got != want {
			t.Errorf("%s: %q, want %q", app, got, want)
		}
	}
}
