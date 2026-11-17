package bottles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzName(f *testing.F) {
	f.Add([]byte("Path: x\nName: \"My Games\"\n"))
	f.Add([]byte("Name: foo\n"))
	f.Add([]byte("Name:\n"))
	f.Add([]byte("name: ignored\n"))
	f.Add([]byte(""))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 64<<10 {
			return
		}
		bottle := t.TempDir()
		if err := os.WriteFile(filepath.Join(bottle, "bottle.yml"), b, 0o600); err != nil {
			t.Fatal(err)
		}
		got := Name(bottle)
		if got == "" {
			t.Fatal("empty name")
		}
		if strings.ContainsAny(got, "\n\x00") {
			t.Fatalf("name holds a newline or NUL: %q", got)
		}
	})
}
