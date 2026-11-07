package steam

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Rethunk-Tech/go-gamelocator/internal/readcap"
	"github.com/Rethunk-Tech/go-safefs"
)

func TestShortcutWithNULIsRefused(t *testing.T) {
	for _, sc := range []Shortcut{
		{Name: "Bad\x00name", Exe: "/x", StartDir: "/"},
		{Name: "n", Exe: "/x\x00y", StartDir: "/"},
		{Name: "n", Exe: "/x", StartDir: "/", LaunchOptions: "a\x00b"},
	} {
		if _, _, _, err := addShortcut(nil, sc); !errors.Is(err, ErrNUL) {
			t.Fatalf("addShortcut(%+v) = %v, want ErrNUL", sc, err)
		}
	}
}

func TestReadCStringIsBounded(t *testing.T) {
	r := bufio.NewReader(bytes.NewReader(bytes.Repeat([]byte{'a'}, maxCString+10)))
	if _, err := readCString(r); err == nil || errors.Is(err, os.ErrNotExist) {
		t.Fatalf("readCString = %v, want a length error", err)
	}
}

func TestReadCapRefusesTooLargeAndPipes(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "big")
	writeAt(t, big, make([]byte, 100))
	if _, err := readcap.File(big, 50); !errors.Is(err, safefs.ErrTooLarge) {
		t.Fatalf("over the cap = %v", err)
	}
	if _, err := readcap.File(dir, 50); err == nil {
		t.Fatal("a directory was read")
	}
}

func TestLargestPNGOffsetPlusSizeCannotWrap(t *testing.T) {
	ico := make([]byte, 6+16)
	binary.LittleEndian.PutUint16(ico[2:], 1)
	binary.LittleEndian.PutUint16(ico[4:], 1)
	binary.LittleEndian.PutUint32(ico[6+8:], 0x7FFFFFF0)
	binary.LittleEndian.PutUint32(ico[6+12:], 0x7FFFFFF0)
	if largestPNG(ico) != nil {
		t.Fatal("entry outside the file was read")
	}
}

func TestInstallDirStaysInTheLibrary(t *testing.T) {
	root := fixture(t)
	outside := filepath.Join(root, "outside")
	if err := os.MkdirAll(outside, 0o750); err != nil {
		t.Fatal(err)
	}
	libs, err := Steam{Root: root}.Libraries()
	if err != nil || len(libs) == 0 {
		t.Fatalf("libraries = %v, %v", libs, err)
	}
	writeAt(t, filepath.Join(libs[0], "steamapps", "appmanifest_9999.acf"),
		[]byte("\"AppState\"\n{\n\t\"installdir\"\t\t\"../../../outside\"\n}\n"))
	if got, err := (Steam{Root: root}).InstallDir("9999"); err != nil || got != "" {
		t.Fatalf("InstallDir = %q, %v", got, err)
	}
}

func TestLibrariesRefuseRelativePaths(t *testing.T) {
	root := t.TempDir()
	writeAt(t, filepath.Join(root, "steamapps", "libraryfolders.vdf"),
		[]byte("\"libraryfolders\"\n{\n\t\"0\"\n\t{\n\t\t\"path\"\t\t\"\"\n\t}\n\t\"1\"\n\t{\n\t\t\"path\"\t\t\"rel/dir\"\n\t}\n}\n"))
	libs, err := Steam{Root: root}.Libraries()
	if err != nil || len(libs) != 0 {
		t.Fatalf("libraries = %v, %v", libs, err)
	}
}

func TestShellQuoteKeepsOrdinaryPathsAndQuotesTheRest(t *testing.T) {
	if got := OverrideCommand("/data/launcher"); !strings.Contains(got, " --filesystem=/data/launcher:ro ") {
		t.Fatalf("plain path was altered: %q", got)
	}
	got := OverrideCommand("/mnt/My Games/it's/$(touch x)")
	want := `flatpak override --user '--filesystem=/mnt/My Games/it'\''s/$(touch x):ro' ` + FlatpakID
	if got != want {
		t.Fatalf("OverrideCommand = %q, want %q", got, want)
	}
}
