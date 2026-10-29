package ea

import (
	"strings"

	"github.com/Rethunk-Tech/go-gamelocator/internal/winreg"
	"golang.org/x/sys/windows/registry"
)

// RegistryDir is the "Install Dir" a game's installer recorded under HKLM key (SOFTWARE\...), in the 64-bit view and
// then the 32-bit one, or "" when neither has it. Vortex's Sims 4 extension reads it the same way
// (Nexus-Mods/Vortex, extensions/games/game-sims4/src/index.js, findGame), which finds a game on any drive, whichever
// EA launcher installed it.
func RegistryDir(key string) string { return registryDir(registry.LOCAL_MACHINE, "SOFTWARE", key) }

// registryDir reads the key below base in root; the real call passes HKLM and SOFTWARE, and tests pass a scratch key.
func registryDir(root registry.Key, base, key string) string {
	rest := strings.TrimPrefix(key, `SOFTWARE\`)
	for _, k := range []string{base + `\` + rest, base + `\WOW6432Node\` + rest} {
		if p := winreg.String(root, k, "Install Dir"); p != "" {
			return p
		}
	}
	return ""
}
