package winreg

import "golang.org/x/sys/windows/registry"

// String is the string value name under path below root, or "" when the key or the value is absent or empty.
func String(root registry.Key, path, name string) string {
	k, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer func() { _ = k.Close() }()
	v, _, err := k.GetStringValue(name)
	if err != nil {
		return ""
	}
	return v
}
