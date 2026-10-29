package heroic

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/Rethunk-Tech/go-safefs"
)

// Prefix is the Wine prefix folder (the one holding drive_c) Heroic runs the game appName in, from the game's
// GamesConfig file in the config folder cfg; "" when it has none yet, or when appName is not a single plain file name
// (so it can never leave GamesConfig). Heroic keeps a Proton prefix as a compatdata folder with the prefix in its pfx
// folder (Heroic-Games-Launcher/HeroicGamesLauncher, src/backend/utils.ts, winePrefix = join(winePrefix, 'pfx')),
// which is why a "proton" wine type means prefix/pfx only and any other type means the prefix only.
func Prefix(cfg, appName string) string {
	if !filepath.IsLocal(appName) || strings.ContainsAny(appName, `/\`) {
		return ""
	}
	b, err := safefs.ReadFile(filepath.Join(cfg, "GamesConfig", appName+".json"))
	if err != nil {
		return ""
	}
	var file map[string]json.RawMessage
	var game struct {
		Prefix string `json:"winePrefix"`
		Wine   struct {
			Type string `json:"type"`
		} `json:"wineVersion"`
	}
	// The file also holds scalar "version" and "explicit" keys beside the game's object.
	if json.Unmarshal(b, &file) != nil || json.Unmarshal(file[appName], &game) != nil {
		return ""
	}
	if game.Prefix == "" {
		return ""
	}
	if game.Wine.Type == "proton" {
		game.Prefix = filepath.Join(game.Prefix, "pfx")
	}
	if !safefs.IsDir(filepath.Join(game.Prefix, "drive_c")) {
		return ""
	}
	return game.Prefix
}
