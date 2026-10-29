package bottles

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

func TestOfWalksUpToBottleYml(t *testing.T) {
	bottle := filepath.Join(t.TempDir(), "Games")
	game := filepath.Join(bottle, "drive_c", "Stardew Valley")
	if err := os.MkdirAll(game, 0o750); err != nil {
		t.Fatal(err)
	}
	if Of(game) != "" {
		t.Fatal("no bottle.yml yet")
	}
	if err := os.WriteFile(filepath.Join(bottle, "bottle.yml"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := Of(game); got != bottle {
		t.Fatalf("Of = %q, want %q", got, bottle)
	}
	if got := Of("drive_c/Stardew Valley"); got != "" {
		t.Fatalf("a relative dir must not search the working directory: %q", got)
	}
}

func TestListAndGameDirsOfABottle(t *testing.T) {
	dir := t.TempDir()
	bottle := filepath.Join(dir, "Games")
	if err := os.MkdirAll(filepath.Join(dir, "NotABottle"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(bottle, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bottle, "bottle.yml"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "stray.yml"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := List(dir); !slices.Equal(got, []string{bottle}) {
		t.Fatalf("List = %v", got)
	}
	if got := List(filepath.Join(dir, "absent")); got != nil {
		t.Fatalf("absent dir: %v", got)
	}
	if got := GameDirs(bottle, "Stardew Valley"); len(got) != 6 || got[0] != filepath.Join(bottle, "drive_c", "Program Files (x86)", "Steam", "steamapps", "common", "Stardew Valley") {
		t.Fatalf("GameDirs = %v", got)
	}
}

func TestDirsFollowCustomXDGData(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Bottles is Linux-only")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // userdirs honours XDG only for os.UserHomeDir, which reads USERPROFILE on Windows
	data := filepath.Join(t.TempDir(), "data")
	want := filepath.Join(data, "bottles", "bottles")
	t.Run("native", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", data)
		if got := Dirs(home); !slices.Contains(got, want) {
			t.Fatalf("%q missing from %v", want, got)
		}
	})
	t.Run("flatpak", func(t *testing.T) {
		t.Setenv("FLATPAK_ID", "test.app")
		t.Setenv("XDG_DATA_HOME", filepath.Join(home, "own", "data"))
		t.Setenv("HOST_XDG_DATA_HOME", data)
		if got := Dirs(home); !slices.Contains(got, want) {
			t.Fatalf("%q missing from %v", want, got)
		}
	})
	if got := Dirs(home, "/mine"); got[0] != "/mine" {
		t.Fatalf("the user's folders come first: %v", got)
	}
}
