// Package gog locates a game installed from GOG: by GOG Galaxy, the offline installer, Heroic or Minigalaxy.
package gog

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/Rethunk-Tech/go-desktopenv/userdirs"
	gamelocator "github.com/Rethunk-Tech/go-gamelocator"
	"github.com/Rethunk-Tech/go-gamelocator/heroic"
	"github.com/Rethunk-Tech/go-safefs"
)

// Source is the launcher a GOG copy came from.
type Source uint8

const (
	// SourceGOG is a GOG Galaxy or offline-installer copy.
	SourceGOG Source = iota
	// SourceHeroic is a copy Heroic installed from GOG.
	SourceHeroic
	// SourceMinigalaxy is a copy Minigalaxy installed from GOG.
	SourceMinigalaxy
)

// Game identifies one game to Locate.
type Game struct {
	// ProductID is the game's GOG product id, as Galaxy's registry and Heroic's installed.json name it.
	ProductID string
	// Folder is the folder name GOG installers and Minigalaxy give the game inside a games folder.
	Folder string
	// Marker is a file every install of the game holds, at its root or one "game" folder down; a slash path.
	Marker string
}

// Install is one GOG-sourced game folder.
type Install struct {
	Dir    string
	Source Source
	// Heroic is the Heroic config folder that lists a Heroic copy, whose GamesConfig holds its prefix; empty otherwise.
	Heroic string
	// AppName is the game's name in Heroic, its GOG product id.
	AppName string
}

// Roots are folders the user added: Heroic config folders and folders of GOG games.
type Roots struct {
	Heroic []string
	Games  []string
	// Minigalaxy are Minigalaxy install folders the user added.
	Minigalaxy []string
}

// Locate finds g's GOG folders under home (and Windows Galaxy / default paths), the user's own folders first. home
// must be a resolved home folder, not "".
func Locate(home string, g Game, r Roots) []Install {
	var out []Install
	seen := map[string]struct{}{}
	add := func(dir string, source Source, heroicCfg ...string) {
		dir = gamelocator.GameDir(dir, g.Marker)
		if dir == "" {
			return
		}
		if _, ok := seen[dir]; ok {
			return
		}
		seen[dir] = struct{}{}
		in := Install{Dir: dir, Source: source}
		if len(heroicCfg) > 0 {
			in.Heroic, in.AppName = heroicCfg[0], g.ProductID
		}
		out = append(out, in)
	}
	add(galaxyPath(g.ProductID), SourceGOG)
	for _, dir := range MinigalaxyInstallDirs(home, r.Minigalaxy...) {
		add(filepath.Join(dir, g.Folder), SourceMinigalaxy)
	}
	for _, dir := range OfflineDirs(home, r) {
		add(filepath.Join(dir, g.Folder), SourceGOG)
	}
	for _, cfg := range heroic.ConfigDirs(home, r.Heroic...) {
		for _, dir := range heroic.GOGInstalls(cfg, g.ProductID) {
			add(dir, SourceHeroic, cfg)
		}
	}
	return out
}

// GamesDirs are the folders GOG installers put games in, the ones in custom first. home must be a resolved home
// folder, not "".
func GamesDirs(home string, custom ...string) []string {
	out := slices.Clone(custom)
	if runtime.GOOS == "windows" {
		return append(out, windowsGamesDirs()...)
	}
	return append(out, filepath.Join(home, "GOG Games"))
}

// OfflineDirs are the GOG games folders not owned by Minigalaxy, which installs into a folder that is often GOG's
// default too; a folder is credited to one launcher only.
func OfflineDirs(home string, r Roots) []string {
	mini := MinigalaxyInstallDirs(home, r.Minigalaxy...)
	return slices.DeleteFunc(GamesDirs(home, r.Games...), func(d string) bool { return slices.Contains(mini, d) })
}

// MinigalaxyConfigDirs are Minigalaxy's config folders (native, then Flatpak). home must be a resolved home folder,
// not "".
func MinigalaxyConfigDirs(home string) []string {
	return []string{
		filepath.Join(userdirs.HostConfigHome(home), "minigalaxy"),
		filepath.Join(userdirs.FlatpakApp(home, "io.github.sharkwouter.Minigalaxy"), "config", "minigalaxy"),
	}
}

// MinigalaxyInstallDirs are the folders Minigalaxy installs games into: the ones in custom first, then each config's
// install_dir (its default, ~/GOG Games, when the config does not set one).
func MinigalaxyInstallDirs(home string, custom ...string) []string {
	out := slices.Clone(custom)
	for _, cfg := range MinigalaxyConfigDirs(home) {
		b, err := safefs.ReadFile(filepath.Join(cfg, "config.json"))
		if err != nil {
			continue
		}
		var c struct {
			InstallDir string `json:"install_dir"`
		}
		if json.Unmarshal(b, &c) != nil || c.InstallDir == "" {
			c.InstallDir = filepath.Join(home, "GOG Games")
		}
		if !slices.Contains(out, c.InstallDir) {
			out = append(out, c.InstallDir)
		}
	}
	return out
}
