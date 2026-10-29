package xbox

import (
	"strings"

	"github.com/Rethunk-Tech/go-safefs"
	"golang.org/x/sys/windows/registry"
)

// packages is the per-user AppModel repository: one subkey per installed package, named by its full name
// (Name_Version_Arch_ResourceId_PublisherId), whose PackageRootFolder is what Get-AppxPackage reports as
// InstallLocation. Reading it directly spares a PowerShell start (about a second each) for every game looked up.
const packages = `Software\Classes\Local Settings\Software\Microsoft\Windows\CurrentVersion\AppModel\Repository\Packages`

// PackageRoot is the install folder of the Windows app package named pkg, as Gale
// (Kesomannen/gale, src-tauri/src/profile/launch/platform.rs) and r2modman
// (ebkr/r2modmanPlus, src/providers/generic/game/directory_resolver/win/XboxGamePassDirectoryResolver.ts) find an Xbox
// copy (Get-AppxPackage), or "" when it is not installed.
func PackageRoot(pkg string) string { return packageRoot(packages, pkg) }

// packageRoot reads the repository below HKCU at base; tests pass a scratch key.
func packageRoot(base, pkg string) string {
	k, err := registry.OpenKey(registry.CURRENT_USER, base, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return ""
	}
	defer func() { _ = k.Close() }()
	names, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return ""
	}
	for _, full := range names {
		if !strings.HasPrefix(full, pkg+"_") {
			continue
		}
		sub, err := registry.OpenKey(k, full, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		dir, _, err := sub.GetStringValue("PackageRootFolder")
		_ = sub.Close()
		if err == nil && safefs.IsDir(dir) {
			return dir
		}
	}
	return ""
}
