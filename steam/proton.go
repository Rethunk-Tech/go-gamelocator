package steam

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/Rethunk-Tech/go-safefs"
)

// ProtonBuild is a Proton a native Steam has: its folder (holding the proton script) and that Steam's root.
type ProtonBuild struct {
	Dir, SteamRoot string
}

var protonVersion = regexp.MustCompile(`^Proton (\d+)\.(\d+)`)

// protonRank orders Proton folder names, best first: the newest numbered Valve release, then Experimental, then any
// other build (GE-Proton and friends) by name, newest first.
func protonRank(a, b string) int {
	ma, mb := protonVersion.FindStringSubmatch(a), protonVersion.FindStringSubmatch(b)
	switch {
	case ma != nil && mb != nil:
		for i := 1; i <= 2; i++ {
			x, _ := strconv.Atoi(ma[i])
			y, _ := strconv.Atoi(mb[i])
			if x != y {
				return y - x
			}
		}
		return 0
	case ma != nil:
		return -1
	case mb != nil:
		return 1
	}
	ea, eb := strings.Contains(a, "Experimental"), strings.Contains(b, "Experimental")
	if ea != eb {
		if ea {
			return -1
		}
		return 1
	}
	return strings.Compare(padNumbers(b), padNumbers(a))
}

var digits = regexp.MustCompile(`\d+`)

// padNumbers zero-pads each run of digits so names compare by their numbers: GE-Proton10 after GE-Proton9.
func padNumbers(s string) string {
	return digits.ReplaceAllStringFunc(s, func(d string) string { return strings.Repeat("0", max(0, 20-len(d))) + d })
}

// ProtonBuilds are the Proton builds a native Steam on this computer has, in its libraries' steamapps/common or its
// compatibilitytools.d, best first. Flatpak Steam's builds need Flatpak's runtime, so they are left out. An empty home
// is the user's home, and extra roots are searched as LocateAll searches them.
func ProtonBuilds(home string, roots ...string) []ProtonBuild {
	var found []ProtonBuild
	for _, st := range LocateAll(home, roots...) {
		if st.Kind != KindNative {
			continue
		}
		dirs := []string{filepath.Join(st.Root, "compatibilitytools.d")}
		libs, _ := st.Libraries()
		for _, lib := range append([]string{st.Root}, libs...) {
			dirs = append(dirs, filepath.Join(lib, "steamapps", "common"))
		}
		for _, d := range dirs {
			entries, _ := os.ReadDir(d)
			for _, e := range entries {
				p := filepath.Join(d, e.Name())
				if safefs.IsFile(filepath.Join(p, "proton")) && !slices.ContainsFunc(found, func(b ProtonBuild) bool { return b.Dir == p }) {
					found = append(found, ProtonBuild{Dir: p, SteamRoot: st.Root})
				}
			}
		}
	}
	slices.SortStableFunc(found, func(a, b ProtonBuild) int { return protonRank(filepath.Base(a.Dir), filepath.Base(b.Dir)) })
	return found
}

// FindProton is the build named want (its folder name) when Steam has it, else the best one.
func FindProton(home, want string, roots ...string) (ProtonBuild, bool) {
	return PickProton(ProtonBuilds(home, roots...), want, "")
}

// foldTool reduces a Proton build's folder name or a Steam compatibility tool name to letters and digits, so
// "proton_9" and "Proton 9.0" compare as a prefix.
func foldTool(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
}

// PickProton is the build named want (a folder name), else the build matching tool (the compatibility tool Steam runs
// the game with, as CompatTool names it), else the best one.
func PickProton(found []ProtonBuild, want, tool string) (ProtonBuild, bool) {
	if i := slices.IndexFunc(found, func(b ProtonBuild) bool { return want != "" && filepath.Base(b.Dir) == want }); i >= 0 {
		return found[i], true
	}
	if t := foldTool(tool); t != "" {
		if i := slices.IndexFunc(found, func(b ProtonBuild) bool { return foldTool(filepath.Base(b.Dir)) == t }); i >= 0 {
			return found[i], true
		}
		if i := slices.IndexFunc(found, func(b ProtonBuild) bool { return strings.HasPrefix(foldTool(filepath.Base(b.Dir)), t) }); i >= 0 {
			return found[i], true
		}
	}
	if len(found) == 0 {
		return ProtonBuild{}, false
	}
	return found[0], true
}

// LibraryOf is the steamapps folder holding the install at dir, found by walking up; ok is false when dir is not in
// one.
func LibraryOf(dir string) (steamapps string, ok bool) {
	steamapps = dir
	for filepath.Base(steamapps) != "steamapps" {
		parent := filepath.Dir(steamapps)
		if parent == steamapps {
			return "", false
		}
		steamapps = parent
	}
	return steamapps, true
}

// PrefixVersion is the first line of the Proton prefix folder's config_info (the tool version that created it), or ""
// when the file is absent or its first line is blank. compatdata is the app's folder under steamapps/compatdata.
func PrefixVersion(compatdata string) string {
	b, err := safefs.ReadFile(filepath.Join(compatdata, "config_info"))
	if err != nil {
		return ""
	}
	line, _, _ := strings.Cut(string(b), "\n")
	return strings.TrimSpace(line)
}

// CompatTool is the compatibility tool Steam's config/config.vdf maps appID to, exactly: "" when the file, the
// mapping or the entry is absent. There is no fallback to the "0" entry (all games), so a caller searching several
// Steam roots asks for the game's own id in every root before it asks for "0".
func (s Steam) CompatTool(appID string) string {
	m, err := parseVDF(filepath.Join(s.Root, "config", "config.vdf"))
	if err != nil {
		return ""
	}
	node := m
	for _, key := range []string{"InstallConfigStore", "Software", "Valve", "Steam", "CompatToolMapping", appID} {
		next, ok := child(node, key)
		if !ok {
			return ""
		}
		node = next
	}
	name, _ := vdfKey(node, "name").(string)
	return name
}
