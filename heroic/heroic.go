// Package heroic reads the Heroic Games Launcher's config folder: where it is, the GOG and Epic (legendary) games it
// installed, and the Wine prefix it runs a game in.
package heroic

import (
	"cmp"
	"encoding/json"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/Rethunk-Tech/go-desktopenv/userdirs"
	"github.com/Rethunk-Tech/go-safefs"
)

// ConfigDirs are Heroic's config folders, the ones in custom first: %APPDATA%\heroic on Windows, else native, Flatpak
// and Snap. home must be a resolved home folder, not "".
func ConfigDirs(home string, custom ...string) []string {
	out := slices.Clone(custom)
	if runtime.GOOS == "windows" {
		return append(out, filepath.Join(userdirs.AppData(home), "heroic"))
	}
	return append(out,
		filepath.Join(userdirs.HostConfigHome(home), "heroic"),
		filepath.Join(userdirs.FlatpakApp(home, "com.heroicgameslauncher.hgl"), "config", "heroic"),
		filepath.Join(home, "snap", "heroic", "current", ".config", "heroic"),
	)
}

type gogGame struct {
	AppName     string `json:"appName"`
	InstallPath string `json:"install_path"`
}

// GOGInstalls are the install folders the GOG library in the config folder cfg lists for productID. The list comes in
// three shapes: a list, an object with "installed" or "games" lists, and an object keyed by product id.
func GOGInstalls(cfg, productID string) []string {
	path := filepath.Join(cfg, "gog_store", "installed.json")
	b, err := safefs.ReadFile(path)
	if err != nil {
		return nil
	}
	var games []gogGame
	var asList struct {
		Installed []gogGame `json:"installed"`
		Games     []gogGame `json:"games"`
	}
	switch {
	case json.Unmarshal(b, &asList) == nil && len(asList.Installed)+len(asList.Games) > 0:
		games = slices.Concat(asList.Installed, asList.Games)
	case json.Unmarshal(b, &games) == nil:
	default:
		var keyed map[string]gogGame
		if json.Unmarshal(b, &keyed) != nil {
			return nil
		}
		for id, g := range keyed {
			g.AppName = cmp.Or(g.AppName, id)
			games = append(games, g)
		}
	}
	var dirs []string
	for _, g := range games {
		if g.AppName != productID {
			continue
		}
		dirs = append(dirs, g.InstallPath)
	}
	return dirs
}
