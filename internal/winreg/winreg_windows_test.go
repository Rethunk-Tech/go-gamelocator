package winreg

import (
	"testing"

	"github.com/Rethunk-Tech/go-gamelocator/internal/winregtest"
	"golang.org/x/sys/windows/registry"
)

func TestStringReadsOneValueAndAnswersEmptyForAnythingAbsent(t *testing.T) {
	base := winregtest.Scratch(t, "winreg-string")
	winregtest.SetString(t, base+`\k`, "v", "value")
	if got := String(registry.CURRENT_USER, base+`\k`, "v"); got != "value" {
		t.Fatalf("got %q", got)
	}
	if got := String(registry.CURRENT_USER, base+`\k`, "absent"); got != "" {
		t.Fatalf("absent value: %q", got)
	}
	if got := String(registry.CURRENT_USER, base+`\absent`, "v"); got != "" {
		t.Fatalf("absent key: %q", got)
	}
}
