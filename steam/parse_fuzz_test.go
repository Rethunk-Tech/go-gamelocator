package steam

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

var seedLocalConfig = []byte(`"UserLocalConfigStore"
{
	"Software"
	{
		"Valve"
		{
			"Steam"
			{
				"apps"
				{
					"123"
					{
						"LaunchOptions"		"-foo"
					}
				}
			}
		}
	}
}
`)

func FuzzScanVDF(f *testing.F) {
	f.Add(seedLocalConfig)
	f.Add([]byte(`"k" "v"`))
	f.Add([]byte(`"k" { }`))
	f.Add([]byte(`"k" { // c
}`))
	f.Add([]byte(`"C:\\saves" "ok"`))
	f.Add([]byte("\"k\"\t\"a\\\"b\\\\\""))
	for _, name := range []string{
		"testdata/localconfig-client.vdf",
		"testdata/lib/steamapps/libraryfolders.vdf",
		"testdata/lib/config/loginusers.vdf",
	} {
		b, err := os.ReadFile(name)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 64<<10 {
			return
		}
		toks, err := scanVDF(b)
		if err != nil {
			return
		}
		prev := 0
		for _, tk := range toks {
			if tk.start < prev || tk.end > len(b) || tk.start >= tk.end {
				t.Fatalf("span %d..%d after %d in %d bytes", tk.start, tk.end, prev, len(b))
			}
			prev = tk.end
			if tk.kind != 's' {
				continue
			}
			got, err := scanVDF([]byte(quoteVDF(tk.text)))
			if err != nil || len(got) != 1 || got[0].kind != 's' || got[0].text != tk.text {
				t.Fatalf("quote round-trip of %q: %v %#v", tk.text, err, got)
			}
		}
	})
}

func FuzzParseTextVDF(f *testing.F) {
	f.Add(seedLocalConfig)
	f.Add([]byte(`"libraryfolders"
{
	"0"
	{
		"path"		"/home/steam"
	}
}
`))
	f.Add([]byte(`"AppState"
{
	"appid"		"123"
	"installdir"		"Game"
	"buildid"		"1"
}
`))
	f.Add([]byte(`"users"
{
	"76561197960265729"
	{
		"AccountName"		"a"
		"PersonaName"		"p"
		"MostRecent"		"1"
	}
}
`))
	f.Add([]byte("// \"\n" + strings.Repeat(`"a"{`, 40)))
	f.Add([]byte("\"k\" { }\n// last line, no line ending"))
	for _, name := range []string{
		"testdata/localconfig-client.vdf",
		"testdata/lib/steamapps/libraryfolders.vdf",
		"testdata/lib/steamapps/appmanifest_413150.acf",
		"testdata/lib/config/loginusers.vdf",
		"testdata/appmanifest_depots.acf",
	} {
		b, err := os.ReadFile(name)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 64<<10 {
			return
		}
		m, err := parseTextVDF(b)
		if bytes.IndexByte(b, 0) >= 0 {
			if err == nil {
				t.Fatal("NUL accepted")
			}
			return
		}
		if depth, derr := textVDFDepth(b); derr == nil && depth > maxTextVDFDepth {
			if err == nil || !errors.Is(err, ErrTooDeep) {
				t.Fatalf("depth %d: %v", depth, err)
			}
			return
		}
		if err == nil && m == nil {
			t.Fatal("nil map")
		}
	})
}

func FuzzBinaryVDF(f *testing.F) {
	var empty, one bytes.Buffer
	_ = writeVDFMap(&empty, []vdfNode{{Kind: vdfMap, Key: "shortcuts"}})
	_ = writeVDFMap(&one, []vdfNode{{Kind: vdfMap, Key: "shortcuts", Child: []vdfNode{
		{Kind: vdfMap, Key: "0", Child: []vdfNode{
			{Kind: vdfString, Key: "AppName", Str: "G"},
			{Kind: vdfString, Key: "Exe", Str: `"C:\g.exe"`},
			{Kind: vdfInt32, Key: "appid", Int: 1},
			{Kind: vdfUint64, Key: "LastPlayTime", Big: 2},
		}},
	}}})
	f.Add(empty.Bytes())
	f.Add(one.Bytes())
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 64<<10 {
			return
		}
		root, err := readVDFMap(bufio.NewReader(bytes.NewReader(b)))
		if err != nil {
			return
		}
		boundVDF(t, root, 0)
		var out bytes.Buffer
		if err := writeVDFMap(&out, root); err != nil {
			t.Fatalf("write of parsed tree: %v", err)
		}
		root2, err := readVDFMap(bufio.NewReader(bytes.NewReader(out.Bytes())))
		if err != nil {
			t.Fatalf("re-read: %v", err)
		}
		var out2 bytes.Buffer
		if err := writeVDFMap(&out2, root2); err != nil {
			t.Fatalf("second write: %v", err)
		}
		if !bytes.Equal(out.Bytes(), out2.Bytes()) {
			t.Fatal("binary VDF write is not stable")
		}
	})
}

func boundVDF(t *testing.T, nodes []vdfNode, depth int) {
	t.Helper()
	if depth > maxBinaryVDFDepth {
		t.Fatalf("tree deeper than %d", maxBinaryVDFDepth)
	}
	for _, n := range nodes {
		if len(n.Key) > maxCString || len(n.Str) > maxCString {
			t.Fatalf("string longer than %d", maxCString)
		}
		if n.Kind == vdfMap {
			boundVDF(t, n.Child, depth+1)
		}
	}
}

func FuzzAddShortcut(f *testing.F) {
	var empty bytes.Buffer
	_ = writeVDFMap(&empty, []vdfNode{{Kind: vdfMap, Key: "shortcuts"}})
	f.Add(empty.Bytes(), "G", `C:\g.exe`, `C:\`, "-foo", "")
	f.Fuzz(func(t *testing.T, body []byte, name, exe, startDir, opts, key string) {
		if len(body) > 64<<10 {
			return
		}
		next, _, _, err := addShortcut(body, Shortcut{
			Name: name, Exe: exe, StartDir: startDir, LaunchOptions: opts, Key: key,
		})
		if err != nil || len(next) == 0 {
			return
		}
		if _, err := readVDFMap(bufio.NewReader(bytes.NewReader(next))); err != nil {
			t.Fatalf("written shortcuts.vdf does not parse: %v", err)
		}
	})
}

func FuzzRemoveShortcuts(f *testing.F) {
	var empty bytes.Buffer
	_ = writeVDFMap(&empty, []vdfNode{{Kind: vdfMap, Key: "shortcuts"}})
	f.Add(empty.Bytes(), `"C:\g.exe"`)
	f.Fuzz(func(t *testing.T, body []byte, prefix string) {
		if len(body) > 64<<10 {
			return
		}
		next, n, err := removeShortcuts(body, func(exe, opts string) bool {
			return exe == prefix || opts == prefix
		})
		if err != nil {
			return
		}
		if n == 0 {
			if !bytes.Equal(next, body) {
				t.Fatal("removed nothing but changed the file")
			}
			return
		}
		if _, err := readVDFMap(bufio.NewReader(bytes.NewReader(next))); err != nil {
			t.Fatalf("rewritten shortcuts.vdf does not parse: %v", err)
		}
	})
}
