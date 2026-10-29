//go:build !windows

package ea

// RegistryDir is empty off Windows: only Windows has the registry the EA installers write to.
func RegistryDir(string) string { return "" }
