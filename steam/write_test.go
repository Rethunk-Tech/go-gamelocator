package steam

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Rethunk-Tech/go-safefs"
)

const testSuffix = ".test.bak"

var testEdit = EditOptions{BackupSuffix: testSuffix}

func TestBackupBeforeEditOnlyCopiesOnce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steam.vdf")
	if err := safefs.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := backupBeforeEdit(path, testSuffix, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := safefs.WriteFileAtomic(path, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := backupBeforeEdit(path, testSuffix, 0o600); err != nil {
		t.Fatal(err)
	}
	body, err := safefs.ReadFile(filepath.Join(dir, "steam.vdf"+testSuffix))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "old" {
		t.Fatalf("backup = %q", body)
	}
}

func TestAddShortcutCreatesTheAccountsFirstShortcutsFile(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o700); err != nil {
		t.Fatal(err)
	}
	users := `"users" { "76561198000000002" { "AccountName" "b" "MostRecent" "1" } }`
	if err := os.WriteFile(filepath.Join(root, "config", "loginusers.vdf"), []byte(users), 0o600); err != nil {
		t.Fatal(err)
	}
	res, err := Steam{Root: root}.AddShortcut(Shortcut{Name: "Main (Stardew Valley)", Exe: "/opt/launcher", StartDir: "/opt"}, testEdit)
	if err != nil || res != Added {
		t.Fatalf("add = %v, %v", res, err)
	}
	dir := filepath.Join(root, "userdata", "39734274", "config")
	if _, err := os.Stat(filepath.Join(dir, "shortcuts.vdf")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "shortcuts.vdf"+testSuffix)); !os.IsNotExist(err) {
		t.Fatalf("a file that did not exist has no backup: %v", err)
	}
}

func TestEditsRefuseAnEmptyBackupSuffixBeforeTouchingAFile(t *testing.T) {
	root := t.TempDir()
	st := Steam{Root: root}
	if _, err := st.AddShortcut(Shortcut{Name: "x", Exe: "/x"}, EditOptions{}); err == nil {
		t.Fatal("AddShortcut")
	}
	if _, err := st.RemoveShortcuts("/x", "", EditOptions{}); err == nil {
		t.Fatal("RemoveShortcuts")
	}
	if _, err := st.SetLaunchOptions("1", func(string) string { return "" }, EditOptions{}); err == nil {
		t.Fatal("SetLaunchOptions")
	}
	if entries, _ := os.ReadDir(root); len(entries) != 0 {
		t.Fatalf("a refused edit left %v", entries)
	}
}
