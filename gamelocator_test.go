package gamelocator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGameDirFindsNestedMarkerAndGameSubfolder(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "Game", "Bin")
	if err := os.MkdirAll(bin, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "TS4_x64.exe"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := GameDir(root, "Game/Bin/TS4_x64.exe"); got != root {
		t.Fatalf("GameDir = %q, want %q", got, root)
	}
	if got := GameDir(root, "TS4_x64.exe"); got != "" {
		t.Fatalf("bare marker found %q", got)
	}
	offline := t.TempDir()
	if err := os.MkdirAll(filepath.Join(offline, "game"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(offline, "game", "x.dll"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := GameDir(offline, "x.dll"); got != filepath.Join(offline, "game") {
		t.Fatalf("the game subfolder: %q", got)
	}
	if got := GameDir("", "x"); got != "" {
		t.Fatalf("empty dir: %q", got)
	}
}
