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
	"github.com/Rethunk-Tech/go-gamelocator/internal/readcap"
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
	AppName, InstallPath string
}

// GOGInstalls are the install folders the GOG library in the config folder cfg lists for productID. The list comes in
// three shapes: a list, an object with "installed" or "games" lists, and an object keyed by product id. A bad entry is
// skipped so one wrong type cannot hide the rest.
func GOGInstalls(cfg, productID string) []string {
	path := filepath.Join(cfg, "gog_store", "installed.json")
	b, err := readcap.File(path, readcap.Config)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, g := range gogGames(b) {
		if g.AppName == productID && g.InstallPath != "" {
			dirs = append(dirs, g.InstallPath)
		}
	}
	return dirs
}

func gogGames(b []byte) []gogGame {
	var asList struct {
		Installed []json.RawMessage `json:"installed"`
		Games     []json.RawMessage `json:"games"`
	}
	if json.Unmarshal(b, &asList) == nil && len(asList.Installed)+len(asList.Games) > 0 {
		return decodeGOGGames(slices.Concat(asList.Installed, asList.Games), nil)
	}
	var list []json.RawMessage
	if json.Unmarshal(b, &list) == nil {
		return decodeGOGGames(list, nil)
	}
	var keyed map[string]json.RawMessage
	if json.Unmarshal(b, &keyed) != nil {
		return nil
	}
	return decodeGOGGames(nil, keyed)
}

func decodeGOGGames(list []json.RawMessage, keyed map[string]json.RawMessage) []gogGame {
	var games []gogGame
	add := func(raw json.RawMessage, id string) {
		if g, ok := parseGOGGame(raw, id); ok {
			games = append(games, g)
		}
	}
	for _, raw := range list {
		add(raw, "")
	}
	for id, raw := range keyed {
		add(raw, id)
	}
	return games
}

func parseGOGGame(raw json.RawMessage, id string) (gogGame, bool) {
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) != nil {
		return gogGame{}, false
	}
	path, ok := jsonString(m["install_path"])
	if !ok || path == "" {
		return gogGame{}, false
	}
	return gogGame{AppName: cmp.Or(jsonID(m["appName"]), id), InstallPath: path}, true
}

func jsonString(b json.RawMessage) (string, bool) {
	var s string
	if json.Unmarshal(b, &s) != nil {
		return "", false
	}
	return s, true
}

func jsonID(b json.RawMessage) string {
	if s, ok := jsonString(b); ok {
		return s
	}
	var n json.Number
	if json.Unmarshal(b, &n) == nil {
		return n.String()
	}
	return ""
}
