package xbox

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageOKRefusesEmptyAndPathSeparators(t *testing.T) {
	if packageOK("") || packageOK(`Vendor/Game`) || packageOK(`Vendor\Game`) {
		t.Fatal("empty or a separator")
	}
	if !packageOK("Vendor.Game") {
		t.Fatal("a package family name")
	}
	b, err := os.ReadFile(filepath.Join("testdata", "relative-root.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(strings.TrimSpace(string(b))) {
		t.Fatal("the relative-root fixture must not be absolute")
	}
}

func TestDefaultLibraryFollowsTheSystemDrive(t *testing.T) {
	t.Setenv("SystemDrive", "D:")
	if got := DefaultLibrary(); got != `D:\XboxGames` {
		t.Fatalf("got %q", got)
	}
	t.Setenv("SystemDrive", "")
	if got := DefaultLibrary(); got != `C:\XboxGames` {
		t.Fatalf("got %q", got)
	}
}
