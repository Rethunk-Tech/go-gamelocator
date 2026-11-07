//go:build unix

package steam

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// the line, run by a real shell with printf in place of flatpak, hands the grant over as one argument and runs
// nothing it names.
func TestOverrideCommandSurvivesAShell(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "pwned")
	path := dir + "/My Games/it's $(touch " + marker + ")"
	line := strings.Replace(OverrideCommand(path), "flatpak", `printf '%s\n'`, 1)
	sh := exec.CommandContext(t.Context(), "sh")
	sh.Stdin = strings.NewReader(line)
	out, err := sh.Output()
	if err != nil {
		t.Fatal(err)
	}
	want := "override\n--user\n--filesystem=" + path + ":ro\n" + FlatpakID + "\n"
	if string(out) != want {
		t.Fatalf("args = %q, want %q", out, want)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the command substitution ran")
	}
}
