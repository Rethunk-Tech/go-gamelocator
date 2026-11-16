package steam

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Rethunk-Tech/vdf/v2"
)

const localConfig = `"UserLocalConfigStore"
{
	"Software"
	{
		"Valve"
		{
			"Steam"
			{
				"apps"
				{
					"413150"
					{
						"LastPlayed"		"1"
						"LaunchOptions"		"-old"
					}
					"440"
					{
						"cloud"
						{
							"last_sync_state"		"synchronized"
						}
					}
				}
			}
		}
	}
	"friends"
	{
		"x"		"y"
	}
}
`

func launchOpts(t *testing.T, b []byte, app string) string {
	t.Helper()
	m, err := vdf.NewParser(bytes.NewReader(b)).Parse()
	if err != nil {
		t.Fatalf("result does not parse: %v\n%s", err, b)
	}
	node := m
	for _, k := range []string{"UserLocalConfigStore", "Software", "Valve", "Steam", "apps", app} {
		next, ok := node[k].(map[string]any)
		if !ok {
			return "<missing>"
		}
		node = next
	}
	v, _ := node["LaunchOptions"].(string)
	return v
}

func TestSetLaunchOptionsEditsOnlyTheOneValue(t *testing.T) {
	line := `"C:\Games\Stardew Valley\StardewModdingAPI.exe" %command%`
	merge := func(old string) string { return strings.TrimSpace(line + " " + old) }

	out, value, err := setLaunchOptions([]byte(localConfig), "413150", merge)
	if err != nil {
		t.Fatal(err)
	}
	if got := launchOpts(t, out, "413150"); got != line+" -old" || value != got {
		t.Fatalf("replaced value = %q (returned %q)", got, value)
	}
	if !bytes.Contains(out, []byte(`"friends"`)) || !bytes.Contains(out, []byte(`"last_sync_state"		"synchronized"`)) {
		t.Fatal("other content must be kept")
	}

	out, _, err = setLaunchOptions([]byte(localConfig), "440", merge)
	if err != nil {
		t.Fatal(err)
	}
	if got := launchOpts(t, out, "440"); got != line {
		t.Fatalf("added key = %q", got)
	}

	out, _, err = setLaunchOptions([]byte(localConfig), "999", merge)
	if err != nil {
		t.Fatal(err)
	}
	if got := launchOpts(t, out, "999"); got != line {
		t.Fatalf("new app block = %q", got)
	}
	if got := launchOpts(t, out, "413150"); got != "-old" {
		t.Fatalf("another app changed: %q", got)
	}

	if _, _, err := setLaunchOptions([]byte(`"UserLocalConfigStore" { }`), "413150", merge); err == nil {
		t.Fatal("a file without an apps section must be refused")
	}
}

// A merge that keeps the value it was given changes nothing, so the file must come back byte for byte and the caller
// must see the path it launched: "C:\new\saves" holds one backslash, not an escaped n.
func TestSetLaunchOptionsIdentityMergeKeepsBytes(t *testing.T) {
	in := []byte(strings.Replace(localConfig, `"-old"`, `"C:\new\saves -old"`, 1))

	out, value, err := setLaunchOptions(in, "413150", func(o string) string { return o })
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, in) {
		t.Fatalf("identity merge rewrote the file:\n%s", out)
	}
	if want := `C:\new\saves -old`; value != want {
		t.Fatalf("value = %q, want %q", value, want)
	}
}

// A merged newline must reach the file escaped: written raw it ends the token, so Steam would load a short value.
func TestSetLaunchOptionsEscapesMergedNewline(t *testing.T) {
	in := []byte(strings.Replace(localConfig, `"-old"`, `"C:\new\saves -old"`, 1))
	want := `C:\new\saves -old` + "\nEXTRA"

	out, value, err := setLaunchOptions(in, "413150", func(o string) string { return o + "\nEXTRA" })
	if err != nil {
		t.Fatal(err)
	}
	if value != want {
		t.Fatalf("value = %q, want %q", value, want)
	}
	if !bytes.Contains(out, []byte(`\nEXTRA"`)) {
		t.Fatalf("the newline was not written as an escape:\n%s", out)
	}
	if !bytes.Contains(out, []byte(`"friends"`)) {
		t.Fatal("other content must be kept")
	}
}
