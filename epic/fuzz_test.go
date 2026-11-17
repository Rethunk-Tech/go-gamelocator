package epic

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzInstalls(f *testing.F) {
	f.Add(fixtureOneBadEntry)
	f.Add([]byte(`{"InstallationList":[{"InstallLocation":"/lib/Good","AppName":"Good"},{"InstallLocation":"/lib/Also","AppName":"Also"}]}`))
	f.Add([]byte(`junk`))
	f.Add([]byte(`{"InstallationList":[]}`))
	f.Add([]byte(`{"InstallationList":["not-an-object"]}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 64<<10 {
			return
		}
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ListFile), b, 0o600); err != nil {
			t.Fatal(err)
		}
		for _, in := range Installs(dir) {
			if in.AppName == "" || in.Dir == "" {
				t.Fatalf("empty field: %+v", in)
			}
		}
	})
}
