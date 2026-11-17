package gog

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzMinigalaxyConfig(f *testing.F) {
	f.Add(fixtureMinigalaxyRelative)
	f.Add([]byte(`{"install_dir":"/abs/Games"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`junk`))
	f.Add([]byte(`{"install_dir":""}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 64<<10 {
			return
		}
		home := t.TempDir()
		cfg := filepath.Join(home, ".config", "minigalaxy")
		if err := os.MkdirAll(cfg, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfg, "config.json"), b, 0o600); err != nil {
			t.Fatal(err)
		}
		for _, d := range MinigalaxyInstallDirs(home) {
			if d == "." || !filepath.IsAbs(d) {
				t.Fatalf("relative install_dir was kept: %q", d)
			}
		}
	})
}
