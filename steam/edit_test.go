package steam

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Rethunk-Tech/go-safefs"
)

// signedIn gives a Steam root whose current account's userdata config folder is returned beside it.
func signedIn(t *testing.T) (Steam, string) {
	t.Helper()
	root := t.TempDir()
	writeAt(t, filepath.Join(root, "config", "loginusers.vdf"), []byte(`"users" { "76561198000000002" { "AccountName" "b" "MostRecent" "1" } }`))
	return Steam{Root: root}, filepath.Join(root, "userdata", "39734274", "config")
}

func TestSetLaunchOptionsBacksUpOnceAndReportsAMissingAppsSection(t *testing.T) {
	s, dir := signedIn(t)
	path := filepath.Join(dir, "localconfig.vdf")
	writeAt(t, path, []byte(localConfig))
	merge := func(old string) string { return "-x " + old }
	if got, err := s.SetLaunchOptions("413150", merge, testEdit); err != nil || got != "-x -old" {
		t.Fatalf("set = %q, %v", got, err)
	}
	if got, err := s.SetLaunchOptions("413150", merge, testEdit); err != nil || got != "-x -x -old" {
		t.Fatalf("set again = %q, %v", got, err)
	}
	if got, err := s.LaunchOptions("413150"); err != nil || got != "-x -x -old" {
		t.Fatalf("read back = %q, %v", got, err)
	}
	if body, err := safefs.ReadFile(path + testSuffix); err != nil || string(body) != localConfig {
		t.Fatalf("the backup is the original: %q, %v", body, err)
	}
	writeAt(t, path, []byte(`"UserLocalConfigStore" { }`))
	if _, err := s.SetLaunchOptions("1", merge, EditOptions{BackupSuffix: ".other"}); !strings.Contains(errString(err), ErrNoAppsSection.Error()) {
		t.Fatalf("err = %v", err)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestRemoveShortcutsFindsEveryAccountsFile(t *testing.T) {
	s, dir := signedIn(t)
	if _, err := s.AddShortcut(Shortcut{Name: "A", Exe: "/opt/launcher", StartDir: "/opt", LaunchOptions: "--play=a"}, testEdit); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddShortcut(Shortcut{Name: "B", Exe: "/opt/other", StartDir: "/opt"}, testEdit); err != nil {
		t.Fatal(err)
	}
	n, err := s.RemoveShortcuts("/opt/launcher", "--play=", testEdit)
	if err != nil || n != 1 {
		t.Fatalf("removed %d, %v", n, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "shortcuts.vdf"+testSuffix)); err != nil {
		t.Fatalf("the first edit takes a backup: %v", err)
	}
}

// A file in the shape a client writes keeps every byte but the one value, or the one block, an edit changes.
func TestSetLaunchOptionsKeepsAClientsFileByteForByte(t *testing.T) {
	orig, err := os.ReadFile("testdata/localconfig-client.vdf")
	if err != nil {
		t.Fatal(err)
	}
	// A value replaced, a key added to a known app, and a block added for an app the file lacks.
	for _, app := range []string{"1966720", "413150", "7"} {
		s, dir := signedIn(t)
		writeAt(t, filepath.Join(dir, "localconfig.vdf"), orig)
		got, err := s.SetLaunchOptions(app, func(old string) string { return "-x " + old }, testEdit)
		if err != nil {
			t.Fatal(err)
		}
		next, _ := safefs.ReadFile(filepath.Join(dir, "localconfig.vdf"))
		if app == "1966720" && (got != "-x -old %command%" || strings.Count(string(next), "LaunchOptions") != 1) {
			t.Fatalf("value = %q", got)
		}
		if again, err := s.LaunchOptions(app); err != nil || again != got {
			t.Fatalf("%s reads back %q, %v; wrote %q", app, again, err, got)
		}
		if !strings.Contains(string(next), "\"SteamDefaultDialog\"\t\t\"#app_games\"\n") || !strings.HasSuffix(string(next), "\t\"WebStorage\"\n\t{\n\t\t\"GameNotesTutorialCompleted\"\t\t\"1\"\n\t}\n}\n") {
			t.Fatalf("app %s: other content changed:\n%s", app, next)
		}
	}
}
