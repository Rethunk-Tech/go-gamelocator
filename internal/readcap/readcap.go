// Package readcap reads files a launcher or the player wrote, with a size bound and without opening a pipe or device.
package readcap

import "github.com/Rethunk-Tech/go-safefs"

// Config bounds a launcher's config, manifest or Steam VDF file; real ones are well under a megabyte.
const Config = 16 << 20

// Image bounds an icon or cover image.
const Image = 32 << 20

// File reads the regular file at path, failing with safefs.ErrTooLarge when it holds more than limit bytes.
func File(path string, limit int64) ([]byte, error) {
	f, err := safefs.OpenRegular(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return safefs.ReadCapped(f, limit)
}
