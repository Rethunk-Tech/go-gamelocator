package steam

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Depot is one depot an installed app has on disk, from its appmanifest's InstalledDepots block.
type Depot struct {
	ID       string
	Manifest string
	Size     int64
}

// InstalledDepots lists the depots appID has installed in library (a library folder, as Libraries returns it),
// ordered by depot id. An app with no InstalledDepots block has none; an app with no manifest in library wraps
// os.ErrNotExist.
func InstalledDepots(library, appID string) ([]Depot, error) {
	if !filepath.IsLocal(appID) {
		return nil, fmt.Errorf("invalid Steam app id %q", appID)
	}
	m, err := parseVDF(filepath.Join(library, "steamapps", "appmanifest_"+appID+".acf"))
	if err != nil {
		return nil, err
	}
	state, _ := m["AppState"].(map[string]any)
	installed, _ := state["InstalledDepots"].(map[string]any)
	var out []Depot
	for id, v := range installed {
		e, ok := v.(map[string]any)
		if !ok {
			continue
		}
		d := Depot{ID: id}
		d.Manifest, _ = e["manifest"].(string)
		size, _ := e["size"].(string)
		d.Size, _ = strconv.ParseInt(size, 10, 64)
		out = append(out, d)
	}
	slices.SortFunc(out, func(a, b Depot) int {
		if len(a.ID) != len(b.ID) {
			return len(a.ID) - len(b.ID)
		}
		return strings.Compare(a.ID, b.ID)
	})
	return out, nil
}

// DepotDownloadDir is the folder Steam's console command download_depot writes depotID of appID to, in library.
func DepotDownloadDir(library, appID, depotID string) string {
	return filepath.Join(library, "steamapps", "content", "app_"+appID, "depot_"+depotID)
}
