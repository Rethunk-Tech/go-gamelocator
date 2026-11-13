package heroic

import (
	"cmp"
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Rethunk-Tech/go-gamelocator/epic"
	"github.com/Rethunk-Tech/go-gamelocator/internal/readcap"
)

// EpicList is the list of the Epic games Heroic installed through legendary, below a Heroic config folder.
func EpicList(cfg string) string {
	return filepath.Join(cfg, "legendaryConfig", "legendary", "installed.json")
}

// EpicInstalls reads the list in the config folder cfg; an unreadable list holds nothing.
func EpicInstalls(cfg string) []epic.Install {
	b, err := readcap.File(EpicList(cfg), readcap.Config)
	if err != nil {
		return nil
	}
	var list map[string]struct {
		AppName string `json:"app_name"`
		Path    string `json:"install_path"`
	}
	if json.Unmarshal(b, &list) != nil {
		return nil
	}
	out := make([]epic.Install, 0, len(list))
	for _, in := range list {
		out = append(out, epic.Install{Dir: in.Path, AppName: in.AppName})
	}
	// Map order is random; callers take the first match.
	slices.SortFunc(out, func(a, b epic.Install) int {
		return cmp.Or(strings.Compare(a.AppName, b.AppName), strings.Compare(a.Dir, b.Dir))
	})
	return out
}
