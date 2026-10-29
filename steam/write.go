package steam

import (
	"errors"
	"os"

	"github.com/Rethunk-Tech/go-safefs"
)

// Sentinels a caller can test with errors.Is.
var (
	// ErrNoAccount means no Steam account has signed in.
	ErrNoAccount = errors.New("no Steam account is signed in")
	// ErrNoAppsSection means localconfig.vdf has no apps section, which Steam writes once a game has been launched.
	ErrNoAppsSection = errors.New("localconfig.vdf has no apps section")
	// ErrNoShortcuts means shortcuts.vdf has no shortcuts list.
	ErrNoShortcuts = errors.New("shortcuts.vdf has no shortcuts list")
	// ErrTooDeep means a VDF file nests deeper than any file Steam writes, which the readers refuse before they
	// recurse because a stack overflow cannot be recovered.
	ErrTooDeep = errors.New("VDF nests deeper than any Steam file")
)

var errNoBackupSuffix = errors.New("EditOptions.BackupSuffix is required")

// EditOptions configures the methods that rewrite a Steam file.
type EditOptions struct {
	// BackupSuffix is required: the one-time copy taken before a file's first edit is the path plus this suffix. An
	// empty suffix is an error rather than a default, so a caller that forgets it cannot leave a second backup under
	// another name beside the one it already has.
	BackupSuffix string
}

func (o EditOptions) validate() error {
	if o.BackupSuffix == "" {
		return errNoBackupSuffix
	}
	return nil
}

func backupBeforeEdit(path, suffix string, perm os.FileMode) error {
	backup := path + suffix
	if _, err := os.Stat(backup); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	body, err := safefs.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		// An account that never had the file (no non-Steam game yet) has nothing to restore.
		return nil
	}
	if err != nil {
		return err
	}
	return safefs.WriteFile(backup, body, perm)
}
