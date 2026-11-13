//go:build unix

package steam

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// Steam owns localconfig.vdf, so an edit keeps its mode even where the umask would narrow a new file's.
func TestSetLaunchOptionsKeepsTheFilesModeUnderAUmask(t *testing.T) {
	s, dir := signedIn(t)
	path := filepath.Join(dir, "localconfig.vdf")
	writeAt(t, path, []byte(localConfig))
	if err := os.Chmod(path, 0o666); err != nil { //nolint:gosec // G302: the mode Steam left is what the test keeps
		t.Fatal(err)
	}
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)
	if _, err := s.SetLaunchOptions("413150", func(string) string { return "-x" }, testEdit); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o666 {
		t.Fatalf("mode %o, want 666", got)
	}
}
