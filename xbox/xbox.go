// Package xbox finds games installed by the Xbox app on Windows.
package xbox

import (
	"cmp"
	"os"
)

// DefaultLibrary is the Xbox app's default library folder, %SystemDrive%\XboxGames.
func DefaultLibrary() string {
	return cmp.Or(os.Getenv("SystemDrive"), "C:") + `\XboxGames`
}
