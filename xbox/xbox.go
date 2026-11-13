// Package xbox finds games installed by the Xbox app on Windows.
package xbox

import (
	"cmp"
	"os"
	"strings"
)

// packageOK is a package family name: one component, never empty, so it cannot match every repository key.
func packageOK(pkg string) bool {
	return pkg != "" && !strings.ContainsAny(pkg, `\/`)
}

// DefaultLibrary is the Xbox app's default library folder, %SystemDrive%\XboxGames.
func DefaultLibrary() string {
	return cmp.Or(os.Getenv("SystemDrive"), "C:") + `\XboxGames`
}
