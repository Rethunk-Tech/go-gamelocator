package heroic

import (
	"cmp"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Rethunk-Tech/go-safefs"
)

func FuzzGOGGames(f *testing.F) {
	f.Add(fixtureGOGOneBad)
	f.Add([]byte(`{"installed":[{"appName":"1453375253","install_path":"/a"},{"appName":"1","install_path":"/nope"}]}`))
	f.Add([]byte(`{"games":[{"appName":"1453375253","install_path":"/a"}]}`))
	f.Add([]byte(`[{"appName":"1453375253","install_path":"/a"}]`))
	f.Add([]byte(`{"1453375253":{"install_path":"/a"}}`))
	f.Add([]byte(`junk`))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 64<<10 {
			return
		}
		for _, g := range gogGames(b) {
			if g.InstallPath == "" {
				t.Fatalf("empty install path: %+v", g)
			}
		}
	})
}

func FuzzEpicInstalls(f *testing.F) {
	f.Add(fixtureEpicOneBad)
	f.Add([]byte(`{"Garlic":{"app_name":"Garlic","install_path":"/games/Garlic","platform":"Windows"}}`))
	f.Add([]byte(`junk`))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 64<<10 {
			return
		}
		cfg := t.TempDir()
		path := EpicList(cfg)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, b, 0o600); err != nil {
			t.Fatal(err)
		}
		var prevName, prevDir string
		for i, in := range EpicInstalls(cfg) {
			if in.AppName == "" || in.Dir == "" {
				t.Fatalf("empty field: %+v", in)
			}
			if i > 0 && cmp.Or(strings.Compare(prevName, in.AppName), strings.Compare(prevDir, in.Dir)) > 0 {
				t.Fatalf("unsorted: %+v after %s %s", in, prevName, prevDir)
			}
			prevName, prevDir = in.AppName, in.Dir
		}
	})
}

func FuzzPrefix(f *testing.F) {
	f.Add(fixturePrefixRelative, "game")
	f.Add([]byte(`{"proton1":{"wineVersion":{"type":"proton"},"winePrefix":"/x"},"version":"v0"}`), "proton1")
	f.Add([]byte(`{"wine1":{"wineVersion":{"type":"wine"},"winePrefix":"/Wine"}}`), "wine1")
	f.Add([]byte(`{}`), "../x")
	f.Add([]byte(`{}`), `a\b`)
	f.Add([]byte(`{}`), "a/b")
	f.Add([]byte("0"), "\x00")
	f.Fuzz(func(t *testing.T, b []byte, appName string) {
		if len(b) > 64<<10 || len(appName) > 64 {
			return
		}
		cfg := t.TempDir()
		if !filepath.IsLocal(appName) || strings.ContainsAny(appName, "/\\\x00") {
			if Prefix(cfg, appName) != "" {
				t.Fatalf("appName %q left GamesConfig", appName)
			}
			return
		}
		if err := os.MkdirAll(filepath.Join(cfg, "GamesConfig"), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfg, "GamesConfig", appName+".json"), b, 0o600); err != nil {
			if Prefix(cfg, appName) != "" {
				t.Fatalf("unwritable appName %q still yielded a prefix", appName)
			}
			return
		}
		got := Prefix(cfg, appName)
		if got == "" {
			return
		}
		if !filepath.IsAbs(got) {
			t.Fatalf("relative prefix %q", got)
		}
		if !safefs.IsDir(filepath.Join(got, "drive_c")) {
			t.Fatalf("prefix %q has no drive_c", got)
		}
	})
}
