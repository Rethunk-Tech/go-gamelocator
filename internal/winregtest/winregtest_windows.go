package winregtest

import (
	"strings"
	"testing"

	"golang.org/x/sys/windows/registry"
)

const testRoot = `Software\go-gamelocator-test`

// Scratch creates HKCU\Software\go-gamelocator-test\name, empties it first, deletes it when the test ends, and returns
// its path below HKCU. Nothing outside that tree is written.
func Scratch(t testing.TB, name string) string {
	t.Helper()
	base := testRoot + `\` + name
	deleteTree(base)
	k, _, err := registry.CreateKey(registry.CURRENT_USER, base, registry.ALL_ACCESS)
	if err != nil {
		t.Fatal(err)
	}
	_ = k.Close()
	t.Cleanup(func() {
		deleteTree(base)
		_ = registry.DeleteKey(registry.CURRENT_USER, testRoot)
	})
	return base
}

// SetString writes the string value name under the key at path below HKCU, creating the key.
func SetString(t testing.TB, path, name, value string) {
	t.Helper()
	k, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = k.Close() }()
	if err := k.SetStringValue(name, value); err != nil {
		t.Fatal(err)
	}
}

func deleteTree(path string) {
	k, err := registry.OpenKey(registry.CURRENT_USER, path, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return
	}
	names, _ := k.ReadSubKeyNames(-1)
	_ = k.Close()
	for _, n := range names {
		deleteTree(strings.TrimRight(path, `\`) + `\` + n)
	}
	_ = registry.DeleteKey(registry.CURRENT_USER, path)
}
