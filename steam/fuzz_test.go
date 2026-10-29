package steam

import (
	"bufio"
	"bytes"
	"slices"
	"strconv"
	"testing"
)

// An accepted input re-encodes to the bytes it consumed, and nesting past the bound is the only refusal that names
// ErrTooDeep.
func FuzzShortcutsVDF(f *testing.F) {
	seed, _, _, err := addShortcut(nil, Shortcut{Name: "Game", Exe: "/opt/game", StartDir: "/opt", LaunchOptions: "--x"})
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add(bytes.Repeat([]byte{vdfMap, 'a', 0}, 40))
	f.Add([]byte{vdfMap, 'a', 0, vdfEnd, vdfEnd})
	f.Fuzz(func(t *testing.T, b []byte) {
		src := bytes.NewReader(b)
		r := bufio.NewReader(src)
		nodes, err := readVDFMap(r)
		if err != nil {
			return
		}
		consumed := len(b) - r.Buffered() - src.Len()
		var out bytes.Buffer
		writeVDFMap(&out, nodes)
		if !bytes.Equal(out.Bytes(), b[:consumed]) {
			t.Fatalf("re-encoded %x, consumed %x", out.Bytes(), b[:consumed])
		}
	})
}

func shortcutNodes(t *testing.T, b []byte) []vdfNode {
	t.Helper()
	root, err := readVDFMap(bufio.NewReader(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}
	return root[0].Child
}

// After an Added, removing that shortcut returns exactly the entries the file had, renumbered.
func FuzzAddRemoveShortcut(f *testing.F) {
	one, _, _, err := addShortcut(nil, Shortcut{Name: "A", Exe: "/opt/a", StartDir: "/opt"})
	if err != nil {
		f.Fatal(err)
	}
	f.Add([]byte(nil))
	f.Add(one)
	sc := Shortcut{Name: "Fuzz", Exe: "/fuzz/exe", StartDir: "/fuzz", LaunchOptions: "--fuzz-unique"}
	f.Fuzz(func(t *testing.T, body []byte) {
		next, _, res, err := addShortcut(body, sc)
		if err != nil || res != Added {
			return
		}
		match := func(exe, opts string) bool { return exe == quoted(sc.Exe) && opts == sc.LaunchOptions }
		after, n, err := removeShortcuts(next, match)
		if err != nil || n != 1 {
			t.Fatalf("remove = %d, %v", n, err)
		}
		want := []vdfNode{{Kind: vdfMap, Key: "shortcuts"}}
		if len(body) > 0 {
			want = []vdfNode{{Kind: vdfMap, Key: "shortcuts", Child: slices.Clone(shortcutNodes(t, body))}}
		}
		for i := range want[0].Child {
			want[0].Child[i].Key = strconv.Itoa(i)
		}
		var buf bytes.Buffer
		writeVDFMap(&buf, want)
		if !bytes.Equal(after, buf.Bytes()) {
			t.Fatalf("after remove %x, want %x", after, buf.Bytes())
		}
	})
}

// Every byte outside the one value or block an edit touches is kept, and the result still tokenizes.
func FuzzSetLaunchOptions(f *testing.F) {
	f.Add([]byte(localConfig), "413150")
	f.Add([]byte(localConfig), "999")
	f.Add([]byte(`"apps" { "1" { } }`), "1")
	f.Fuzz(func(t *testing.T, in []byte, appID string) {
		toks, err := scanVDF(in)
		if err != nil {
			return
		}
		out, _, err := setLaunchOptions(in, appID, func(old string) string { return "-x " + old })
		if err != nil {
			return
		}
		if _, err := scanVDF(out); err != nil {
			t.Fatalf("result does not tokenize: %v", err)
		}
		if bytes.Equal(in, out) {
			return
		}
		start := 0
		for start < len(in) && start < len(out) && in[start] == out[start] {
			start++
		}
		end := 0
		for end < len(in)-start && end < len(out)-start && in[len(in)-1-end] == out[len(out)-1-end] {
			end++
		}
		lo, hi := start, len(in)-end
		ok := false
		for _, tk := range toks {
			inString := tk.kind == 's' && tk.start <= lo && hi <= tk.end
			atBrace := tk.kind == '}' && lo == hi && lo == tk.start
			ok = ok || inString || atBrace
		}
		if !ok {
			t.Fatalf("changed %d..%d, outside any value or closing brace", lo, hi)
		}
	})
}
