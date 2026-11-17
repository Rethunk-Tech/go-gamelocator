package lutris

import (
	"path/filepath"
	"strings"
	"testing"
)

func FuzzWinePath(f *testing.F) {
	f.Add("/tmp/pfx", `C:\GOG Games\Stardew Valley`)
	f.Add("/tmp/pfx", `C:\..\..\outside`)
	f.Add("/tmp/pfx", `..\..\outside`)
	f.Add("/tmp/pfx", `Z:\home\player`)
	f.Add("/tmp/pfx", `/abs/unix`)
	f.Add("/tmp/pfx", `drive_c/GOG Games`)
	f.Add("/tmp/pfx", `C:\GOG Games\Stardew Valley\StardewValley.exe`)
	f.Add("", `C:\x`)
	f.Add("/tmp/pfx", `Z:C:\Users\me`)
	f.Add("/", "0")
	f.Fuzz(func(t *testing.T, prefix, p string) {
		if len(prefix) > 4<<10 || len(p) > 4<<10 {
			return
		}
		got := winePath(prefix, p)
		trimmed := strings.TrimSpace(p)
		unix, isDrive := fromWindowsDrive(prefix, trimmed)
		if isDrive && got != unix {
			t.Fatalf("winePath=%q fromWindowsDrive=%q", got, unix)
		}
		if got == "" {
			return
		}
		if isCDrive(trimmed) && !inRoot(filepath.Join(wineRoot(prefix), "drive_c"), got) {
			t.Fatalf("C: escaped prefix: path=%q got=%q", p, got)
		}
		if !isDrive && !filepath.IsAbs(trimmed) && !inRoot(wineRoot(prefix), got) {
			t.Fatalf("relative escaped: path=%q got=%q", p, got)
		}
	})
}

func isCDrive(p string) bool {
	if len(p) < 2 || p[1] != ':' {
		return false
	}
	drive := p[0] | 0x20
	if drive != 'c' {
		return false
	}
	return len(p) == 2 || p[2] == '\\' || p[2] == '/'
}

func inRoot(root, path string) bool {
	if root == "" {
		return false
	}
	r := filepath.Clean(root)
	p := filepath.Clean(path)
	if p == r {
		return true
	}
	sep := string(filepath.Separator)
	if r == sep {
		return strings.HasPrefix(p, sep)
	}
	return strings.HasPrefix(p, r+sep)
}

func FuzzInstallFromYAML(f *testing.F) {
	f.Add(string(fixtureGogLinux))
	f.Add(string(fixtureSteamRunner))
	f.Add(string(fixtureWineCDrive))
	f.Add(string(fixtureWineRelative))
	f.Add(string(fixtureWineZDrive))
	f.Add(string(fixtureWineCEscape))
	f.Add(string(fixtureSlugOnly))
	m := newMatcher(testGame)
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 64<<10 {
			return
		}
		got := m.installFromYAML(text)
		if reRunnerSteam.MatchString(text) && got != "" {
			t.Fatalf("steam runner yielded %q", got)
		}
	})
}

func FuzzRootOfMarker(f *testing.F) {
	f.Add("/games/Game/Bin/TS4_x64.exe", "Game/Bin/TS4_x64.exe")
	f.Add(`/pfx/drive_c/GOG Games/Stardew Valley/Stardew Valley.dll`, "Stardew Valley.dll")
	f.Add("/games/x", "../x")
	f.Fuzz(func(t *testing.T, exe, marker string) {
		if len(exe)+len(marker) > 8<<10 {
			return
		}
		got := rootOfMarker(exe, marker)
		if got == "" {
			return
		}
		slash := filepath.ToSlash(exe)
		prefix := filepath.ToSlash(got)
		if len(slash) <= len(prefix) || slash[len(prefix)] != '/' || !strings.EqualFold(slash[len(prefix)+1:], marker) {
			t.Fatalf("root %q does not place marker %q under exe %q", got, marker, exe)
		}
	})
}
