package ea

import (
	"testing"

	"github.com/Rethunk-Tech/go-gamelocator/internal/winregtest"
	"golang.org/x/sys/windows/registry"
)

func TestRegistryDirTriesThe64BitViewThenWOW6432Node(t *testing.T) {
	base := winregtest.Scratch(t, "ea-registry")
	winregtest.SetString(t, base+`\EA Games\Native`, "Install Dir", `D:\EA\Native`)
	winregtest.SetString(t, base+`\WOW6432Node\EA Games\Wow`, "Install Dir", `D:\EA\Wow`)
	winregtest.SetString(t, base+`\EA Games\Both`, "Install Dir", `D:\EA\Native`)
	winregtest.SetString(t, base+`\WOW6432Node\EA Games\Both`, "Install Dir", `D:\EA\Wow`)
	for key, want := range map[string]string{
		`SOFTWARE\EA Games\Native`: `D:\EA\Native`,
		`SOFTWARE\EA Games\Wow`:    `D:\EA\Wow`,
		`SOFTWARE\EA Games\Both`:   `D:\EA\Native`,
		`SOFTWARE\EA Games\None`:   "",
	} {
		if got := registryDir(registry.CURRENT_USER, base, key); got != want {
			t.Errorf("%s: %q, want %q", key, got, want)
		}
	}
}
