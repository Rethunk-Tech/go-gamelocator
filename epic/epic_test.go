package epic

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

func TestListDirsAreWindowsOnlyAndUserFoldersComeFirst(t *testing.T) {
	env := func(k string) string {
		if k == "ProgramData" {
			return `C:\PD`
		}
		return ""
	}
	if got, want := listDirs("windows", env, "D:/mine"), []string{"D:/mine", filepath.Join(`C:\PD`, "Epic", "UnrealEngineLauncher")}; !slices.Equal(got, want) {
		t.Fatalf("windows = %v, want %v", got, want)
	}
	if got := listDirs("linux", env, "/mine"); !slices.Equal(got, []string{"/mine"}) {
		t.Fatalf("linux has no launcher folder of its own: %v", got)
	}
}

func writeList(t *testing.T, dir string, installs ...Install) {
	t.Helper()
	body := `{"InstallationList":[`
	for i, in := range installs {
		if i > 0 {
			body += ","
		}
		body += fmt.Sprintf(`{"InstallLocation":%q,"AppName":%q}`, in.Dir, in.AppName)
	}
	body += `]}`
	if err := os.WriteFile(filepath.Join(dir, ListFile), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestInstallsReadsTheLaunchersList(t *testing.T) {
	dir := t.TempDir()
	if got := Installs(dir); len(got) != 0 {
		t.Fatalf("no list: %v", got)
	}
	want := []Install{{AppName: "Other", Dir: "/lib/Other"}, {AppName: "Garlic", Dir: "/lib/Garlic"}}
	writeList(t, dir, want...)
	if got := Installs(dir); !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if err := os.WriteFile(filepath.Join(dir, ListFile), []byte("junk"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := Installs(dir); len(got) != 0 {
		t.Fatalf("junk: %v", got)
	}
}

func TestAppNameAtPrefersTheListedProductAndFallsBackToTheFirst(t *testing.T) {
	list := t.TempDir()
	game := filepath.Join(t.TempDir(), "Garlic")
	writeList(t, list, Install{AppName: "Second", Dir: game})
	names := []string{"First", "Second"}
	if got := appNameAt([]string{list}, game, names); got != "Second" {
		t.Fatalf("listed: %q", got)
	}
	if got := appNameAt([]string{list}, filepath.Join(game, "game"), names); got != "Second" {
		t.Fatalf("a game folder one below the listed one: %q", got)
	}
	if got := appNameAt([]string{list}, "/elsewhere", names); got != "First" {
		t.Fatalf("unlisted: %q", got)
	}
	if got := appNameAt(nil, "/elsewhere", nil); got != "" {
		t.Fatalf("no names: %q", got)
	}
}

func TestPublicWrappersUseTheHostOS(t *testing.T) {
	if runtime.GOOS != "windows" {
		if got := ListDirs("/mine"); !slices.Equal(got, []string{"/mine"}) {
			t.Fatalf("ListDirs = %v", got)
		}
	}
	// No launcher lists the folder, so the first name stands and no names gives "".
	if got := AppNameAt(t.TempDir(), []string{"A", "B"}); got != "A" {
		t.Fatalf("AppNameAt = %q", got)
	}
	if got := AppNameAt(t.TempDir(), nil); got != "" {
		t.Fatalf("AppNameAt with no names = %q", got)
	}
}
