//go:build !windows

package xbox

// PackageRoot is empty off Windows: only Windows has app packages.
func PackageRoot(string) string { return "" }
