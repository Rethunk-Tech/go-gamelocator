package xbox

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Rethunk-Tech/go-gamelocator/internal/winregtest"
)

func TestPackageRootTakesTheFirstPackageWhoseFolderExists(t *testing.T) {
	base := winregtest.Scratch(t, "xbox-appx")
	root := t.TempDir()
	gone := filepath.Join(root, "gone")
	here := filepath.Join(root, "here")
	if err := os.Mkdir(here, 0o750); err != nil {
		t.Fatal(err)
	}
	winregtest.SetString(t, base+`\Vendor.Game_1.0_x64__abc`, "PackageRootFolder", gone)
	winregtest.SetString(t, base+`\Vendor.Game_2.0_x64__abc`, "PackageRootFolder", here)
	winregtest.SetString(t, base+`\Vendor.GameTwo_1.0_x64__abc`, "PackageRootFolder", root)
	if got := packageRoot(base, "Vendor.Game"); got != here {
		t.Fatalf("got %q, want %q", got, here)
	}
	if got := packageRoot(base, "Vendor.Missing"); got != "" {
		t.Fatalf("missing package: %q", got)
	}
	if got := packageRoot(base+`\absent`, "Vendor.Game"); got != "" {
		t.Fatalf("absent repository: %q", got)
	}
}

func TestPackageRootIgnoresAnEmptyNameAndARelativeFolder(t *testing.T) {
	base := winregtest.Scratch(t, "xbox-appx-hostile")
	here := t.TempDir()
	winregtest.SetString(t, base+`\_weird_1.0_x64__abc`, "PackageRootFolder", here)
	if got := packageRoot(base, ""); got != "" {
		t.Fatalf("empty pkg = %q", got)
	}
	rel, err := os.ReadFile(filepath.Join("testdata", "relative-root.txt"))
	if err != nil {
		t.Fatal(err)
	}
	winregtest.SetString(t, base+`\Vendor.Game_1.0_x64__abc`, "PackageRootFolder", strings.TrimSpace(string(rel)))
	if got := packageRoot(base, "Vendor.Game"); got != "" {
		t.Fatalf("relative folder = %q", got)
	}
}
