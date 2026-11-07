package steam

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestInstalledDepots(t *testing.T) {
	lib := t.TempDir()
	body, err := os.ReadFile("testdata/appmanifest_depots.acf")
	if err != nil {
		t.Fatal(err)
	}
	writeAt(t, filepath.Join(lib, "steamapps", "appmanifest_1000.acf"), body)
	got, err := InstalledDepots(lib, "1000")
	if err != nil {
		t.Fatal(err)
	}
	want := []Depot{{"999", "1111111111111111111", 42}, {"1001", "7446623914455268361", 1048576}}
	if !slices.Equal(got, want) {
		t.Fatalf("depots = %#v", got)
	}
	if _, err := InstalledDepots(lib, "2000"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing manifest: %v", err)
	}
	if _, err := InstalledDepots(lib, "../x"); err == nil {
		t.Fatal("an app id leaving its folder was accepted")
	}
}

func TestDepotDownloadDir(t *testing.T) {
	want := filepath.Join("lib", "steamapps", "content", "app_1000", "depot_1001")
	if got := DepotDownloadDir("lib", "1000", "1001"); got != want {
		t.Fatalf("dir = %s", got)
	}
}

func TestDepotJSONNames(t *testing.T) {
	t.Parallel()

	b, err := json.Marshal(Depot{ID: "1", Manifest: "2", Size: 3})
	if err != nil || string(b) != `{"id":"1","manifest":"2","size":3}` {
		t.Fatalf("json = %s, %v", b, err)
	}
}

func TestSteamIDsThatAreNotNumbersAreRefused(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"", "..", "../x", "12a", "1/2", `1\2`, "-1", " 1", "123456789012345678901"} {
		if got := DepotDownloadDir("lib", id, "1001"); got != "" {
			t.Errorf("DepotDownloadDir(app %q) = %q", id, got)
		}
		if got := DepotDownloadDir("lib", "1000", id); got != "" {
			t.Errorf("DepotDownloadDir(depot %q) = %q", id, got)
		}
		if got, err := (Steam{Root: t.TempDir()}).InstallDir(id); got != "" || err != nil {
			t.Errorf("InstallDir(%q) = %q, %v", id, got, err)
		}
		if got := BuildID(t.TempDir(), id); got != "" {
			t.Errorf("BuildID(%q) = %q", id, got)
		}
	}
}
