//go:build !windows

package xbox

import "testing"

func TestPackageRootIsEmptyOffWindows(t *testing.T) {
	if got := PackageRoot("Vendor.Game"); got != "" {
		t.Fatalf("got %q", got)
	}
}
