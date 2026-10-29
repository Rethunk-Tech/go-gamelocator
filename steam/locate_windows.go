//go:build windows

package steam

import (
	"os"
	"path/filepath"

	"github.com/Rethunk-Tech/go-gamelocator/internal/winreg"
	"golang.org/x/sys/windows/registry"
)

func candidates(string) []string {
	out := candidatesAt(registry.CURRENT_USER, registry.LOCAL_MACHINE, `Software`, `SOFTWARE`)
	if pf := os.Getenv("ProgramFiles(x86)"); pf != "" {
		out = append(out, filepath.Join(pf, "Steam"))
	}
	return out
}

// candidatesAt reads the Steam keys below userBase in user and below machineBase in machine; the real call passes
// HKCU\Software and HKLM\SOFTWARE, and tests pass a scratch key.
func candidatesAt(user, machine registry.Key, userBase, machineBase string) []string {
	var out []string
	add := func(root registry.Key, path, name string) {
		if p := winreg.String(root, path, name); p != "" {
			out = append(out, filepath.Clean(p))
		}
	}
	add(user, userBase+`\Valve\Steam`, "SteamPath")
	add(machine, machineBase+`\Wow6432Node\Valve\Steam`, "InstallPath")
	add(machine, machineBase+`\Valve\Steam`, "InstallPath")
	return out
}
