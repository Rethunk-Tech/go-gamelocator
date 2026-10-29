package xbox

import "testing"

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
