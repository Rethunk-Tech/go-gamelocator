package gog

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Rethunk-Tech/go-gamelocator/internal/winreg"
	"golang.org/x/sys/windows/registry"
)

// galaxyKey is GOG's registry tree below HKLM; Galaxy is a 32-bit program, so it lives under WOW6432Node.
const galaxyKey = `SOFTWARE\WOW6432Node\GOG.com`

func windowsGamesDirs() []string {
	dirs := []string{`C:\GOG Games`}
	if pf := os.Getenv("ProgramFiles(x86)"); pf != "" {
		dirs = append(dirs, filepath.Join(pf, "GOG Galaxy", "Games"))
	}
	return dirs
}

// GalaxyDir is GOG Galaxy's install folder from its registry entry, else its default folder, whether or not it exists.
func GalaxyDir() string {
	if p := galaxyClient(registry.LOCAL_MACHINE, galaxyKey); p != "" {
		return p
	}
	if pf := os.Getenv("ProgramFiles(x86)"); pf != "" {
		return filepath.Join(pf, "GOG Galaxy")
	}
	return ""
}

func galaxyClient(root registry.Key, base string) string {
	if p := winreg.String(root, base+`\GalaxyClient\paths`, "client"); p != "" {
		return filepath.Clean(p)
	}
	return ""
}

// lookupGalaxy is the Galaxy registry read; tests replace it so Locate never reads the real registry.
var lookupGalaxy = readGalaxyRegistry

func readGalaxyRegistry(productID string) string {
	return readGalaxyGame(registry.LOCAL_MACHINE, galaxyKey, productID)
}

// readGalaxyGame is the install folder Galaxy recorded for productID below base in root. A product id is one key name,
// never a path.
func readGalaxyGame(root registry.Key, base, productID string) string {
	if productID == "" || strings.ContainsAny(productID, `\/`) {
		return ""
	}
	return winreg.String(root, base+`\Games\`+productID, "path")
}

func galaxyPath(productID string) string { return lookupGalaxy(productID) }
