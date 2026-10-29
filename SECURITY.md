# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's **private vulnerability reporting** on this repository (Security → Report a
vulnerability). It stays private until a fix ships.

Please do not open a public issue for a suspected vulnerability first.

## What is in scope

- **Steam's text VDF files** (`libraryfolders.vdf`, `loginusers.vdf`, `localconfig.vdf`, `config.vdf`, app
  manifests), read by `Libraries`, `InstallDir`, `BuildID`, `CurrentAccount`, `LaunchOptions` and `CompatTool`. A file
  must not cause a fatal stack overflow or an allocation past 8 MiB: depth past 32 and size past the cap return errors.
- **`shortcuts.vdf`**, binary, read by `AddShortcut` and `RemoveShortcuts`. Nesting past 16 returns `ErrTooDeep`, and
  any accepted file re-encodes to the bytes read (fuzzed).
- **App ids and Heroic app names** are joined into paths: `HeroArt`, `ClientIcon` and `BuildID` refuse an id that
  would leave its folder (`filepath.IsLocal`), and `heroic.Prefix` refuses anything but one plain file name.
- **Heroic, Epic, Minigalaxy and Lutris files** are plain JSON or YAML text; a malformed one yields no installs, never
  a panic.
- **Registry values** (`GalaxyDir`, `ea.RegistryDir`, `xbox.PackageRoot`, Steam paths) are returned as paths only; a
  Galaxy product id containing a separator is refused before it reaches a key path.

## What is out of scope

- The caller passes a resolved `home` and trusted extra roots; the module does not decide which folders a user may add.
- Paths read from game files (a library path in `libraryfolders.vdf`, an install path in Heroic's list) are returned
  as they are, and checking them before use is the caller's job.
- Edits write inside the Steam folder the caller named (the signed-in account's for launch options and `AddShortcut`,
  every account's for `RemoveShortcuts`), and trust the caller to have closed Steam.
- `GrantFilesystem` widens a Flatpak sandbox on request; deciding whether to ask is the caller's.
