package wineprefix

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRootPrefersProtonsPfxFolder(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{filepath.Join(dir, "proton", "pfx", "drive_c"), filepath.Join(dir, "proton", "drive_c"), filepath.Join(dir, "wine", "drive_c")} {
		if err := os.MkdirAll(p, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	for in, want := range map[string]string{
		filepath.Join(dir, "proton"):  filepath.Join(dir, "proton", "pfx"),
		filepath.Join(dir, "wine"):    filepath.Join(dir, "wine"),
		filepath.Join(dir, "missing"): "",
		"":                            "",
	} {
		if got := Root(in); got != want {
			t.Errorf("Root(%q) = %q, want %q", in, got, want)
		}
	}
}
