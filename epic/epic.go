// Package epic reads the install list the Epic Games Launcher keeps on Windows.
package epic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/Rethunk-Tech/go-safefs"
)

// ListFile is the Epic Games Launcher's list of the games it installed and where. r2modman reads it in
// src/providers/generic/game/directory_resolver/win/EGSDirectoryResolver.ts, Gale in
// src-tauri/src/profile/launch/platform.rs.
const ListFile = "LauncherInstalled.dat"

// Install is one game the launcher lists: its AppName and install folder.
type Install struct {
	AppName, Dir string
}

// ListDirs are the folders holding the launcher's install list, the ones in custom first. Only Windows has the
// launcher.
func ListDirs(custom ...string) []string { return listDirs(runtime.GOOS, os.Getenv, custom...) }

func listDirs(goos string, env func(string) string, custom ...string) []string {
	out := slices.Clone(custom)
	if pd := env("ProgramData"); goos == "windows" && pd != "" {
		out = append(out, filepath.Join(pd, "Epic", "UnrealEngineLauncher"))
	}
	return out
}

// Installs reads the install list in dir; an unreadable list holds nothing.
func Installs(dir string) []Install {
	b, err := safefs.ReadFile(filepath.Join(dir, ListFile))
	if err != nil {
		return nil
	}
	var list struct {
		Installs []struct {
			Location string `json:"InstallLocation"`
			AppName  string `json:"AppName"`
		} `json:"InstallationList"`
	}
	if json.Unmarshal(b, &list) != nil {
		return nil
	}
	out := make([]Install, 0, len(list.Installs))
	for _, in := range list.Installs {
		out = append(out, Install{AppName: in.AppName, Dir: in.Location})
	}
	return out
}

// AppNameAt is the AppName the launcher lists for the install in dir (its folder, or the one above its game folder),
// else the first of names: a game sold as several products is started as the one installed.
func AppNameAt(dir string, names []string) string {
	return appNameAt(ListDirs(), dir, names)
}

func appNameAt(lists []string, dir string, names []string) string {
	for _, list := range lists {
		for _, in := range Installs(list) {
			loc := filepath.Clean(in.Dir)
			if slices.Contains(names, in.AppName) && (strings.EqualFold(loc, dir) || strings.EqualFold(loc, filepath.Dir(dir))) {
				return in.AppName
			}
		}
	}
	if len(names) == 0 {
		return ""
	}
	return names[0]
}
