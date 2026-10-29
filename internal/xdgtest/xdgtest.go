// Package xdgtest clears the XDG variables a test package's locators would read.
package xdgtest

import (
	"os"
	"testing"
)

// Main runs m with no XDG or HOST_XDG folder set: tests pass their own temporary home, and a CI runner's XDG folders
// would otherwise win over it.
func Main(m *testing.M) {
	for _, k := range []string{"XDG_CONFIG_HOME", "XDG_DATA_HOME", "HOST_XDG_CONFIG_HOME", "HOST_XDG_DATA_HOME"} {
		if err := os.Unsetenv(k); err != nil {
			panic(err)
		}
	}
	os.Exit(m.Run())
}
