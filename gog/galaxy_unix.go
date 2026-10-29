//go:build !windows

package gog

func galaxyPath(string) string { return "" }

func windowsGamesDirs() []string { return nil }

// GalaxyDir is empty: GOG Galaxy runs only on Windows.
func GalaxyDir() string { return "" }
