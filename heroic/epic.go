package heroic

import (
	"encoding/json"
	"path/filepath"

	"github.com/Rethunk-Tech/go-gamelocator/epic"
	"github.com/Rethunk-Tech/go-safefs"
)

// EpicList is the list of the Epic games Heroic installed through legendary, below a Heroic config folder.
func EpicList(cfg string) string {
	return filepath.Join(cfg, "legendaryConfig", "legendary", "installed.json")
}

// EpicInstalls reads the list in the config folder cfg; an unreadable list holds nothing.
func EpicInstalls(cfg string) []epic.Install {
	b, err := safefs.ReadFile(EpicList(cfg))
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
	return out
}
