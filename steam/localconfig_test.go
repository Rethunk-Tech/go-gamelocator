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

func scanLaunchOpts(t *testing.T, b []byte, app string) (text string, span []byte) {
	t.Helper()
	toks, err := scanVDF(b)
	if err != nil {
		t.Fatal(err)
	}
	loc, err := locateLaunchOptions(toks, app)
	if err != nil {
		t.Fatal(err)
	}
	if loc.optVal == nil {
		t.Fatal("no LaunchOptions token")
	}
	span = b[loc.optVal.start:loc.optVal.end]
	if len(span) < 2 || span[0] != '"' || span[len(span)-1] != '"' {
		t.Fatalf("LaunchOptions span %d..%d is not a quoted token: %q", loc.optVal.start, loc.optVal.end, span)
	}
	return loc.optVal.text, span
}

func parsedLaunchOpts(t *testing.T, b []byte, app string) string {
	t.Helper()
	m, err := parseTextVDF(b)
	if err != nil {
		t.Fatalf("parseTextVDF: %v\n%s", err, b)
	}
	node := m
	for _, k := range []string{"UserLocalConfigStore", "Software", "Valve", "Steam", "apps", app} {
		next, ok := child(node, k)
		if !ok {
			t.Fatalf("missing %s", k)
		}
		node = next
	}
	for k, v := range node {
		if s, ok := v.(string); ok && strings.EqualFold(k, "LaunchOptions") {
			return s
		}
	}
	t.Fatal("no LaunchOptions")
	return ""
}

// quoteVDF leaves newline and tab raw, matching Valve's writer, so a write-then-read of those
// bytes (and of a Windows path, a quote, or a trailing backslash) must come back identical.
func TestSetLaunchOptionsRoundTripsSpecialBytes(t *testing.T) {
	cases := []struct {
		name, value string
	}{
		{"newline", "a\nb"},
		{"tab", "a\tb"},
		{"backslash-n path", `C:\new\saves`},
		{"quote", `say "hi"`},
		{"trailing backslash", `dir\`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, value, err := setLaunchOptions([]byte(localConfig), "413150", func(string) string { return tc.value })
			if err != nil {
				t.Fatal(err)
			}
			if value != tc.value {
				t.Fatalf("returned %q, want %q", value, tc.value)
			}
			got, span := scanLaunchOpts(t, out, "413150")
			if got != tc.value {
				t.Fatalf("scanVDF = %q, want %q", got, tc.value)
			}
			if strings.ContainsRune(tc.value, '\n') && !bytes.Contains(span, []byte{'\n'}) {
				t.Fatalf("quoted span escaped the newline: %q", span)
			}
			if strings.ContainsRune(tc.value, '\t') && !bytes.Contains(span, []byte{'\t'}) {
				t.Fatalf("quoted span escaped the tab: %q", span)
			}
			if got := parsedLaunchOpts(t, out, "413150"); got != tc.value {
				t.Fatalf("LaunchOptions = %q, want %q", got, tc.value)
			}
		})
	}
}
