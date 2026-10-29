package ea

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestDirsAreWindowsOnlyAndUserFoldersComeFirst(t *testing.T) {
	env := map[string]string{"ProgramFiles": `C:\PF`, "ProgramFiles(x86)": `C:\PF86`}
	get := func(k string) string { return env[k] }
	got := dirs("windows", get, "D:/mine")
	want := []string{
		"D:/mine",
		filepath.Join(`C:\PF`, "EA Games"), filepath.Join(`C:\PF`, "Electronic Arts"),
		filepath.Join(`C:\PF86`, "EA Games"), filepath.Join(`C:\PF86`, "Electronic Arts"),
	}
	if !slices.Equal(got, want) {
		t.Fatalf("windows dirs = %v, want %v", got, want)
	}
	if got := dirs("linux", get, "/mine"); !slices.Equal(got, []string{"/mine"}) {
		t.Fatalf("linux has no EA App folders of its own: %v", got)
	}
}
