//go:build !windows

package ea

import "testing"

func TestRegistryDirIsEmptyOffWindows(t *testing.T) {
	if got := RegistryDir(`SOFTWARE\EA Games\X`); got != "" {
		t.Fatalf("got %q", got)
	}
}
