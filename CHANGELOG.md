# Changelog

Notable changes to `go-gamelocator`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- `epic`: `Installs` skips a malformed or incomplete entry instead of dropping the list.

## [0.4.0] - 2026-11-07

### Added

- `steam`: `CompatData` and `Steam.CompatData` find an app's compatdata folder in any library, `ProtonUserDir` names the Proton prefix's Windows profile, and `StoreHero` asks Steam's store for an app's hero image URL.
- `bottles`: `Name`, `UserDir`, `IsFlatpak` and `RunArgv` read a bottle's name and Windows profile and build the `bottles-cli run` command line.

### Changed

- Text VDF is parsed by `github.com/Rethunk-Tech/vdf/v2`, a new dependency.
- `GameDir` returns "" for a relative `dir`, and `Steam.Libraries` skips a relative library path.

### Fixed

- `steam.DepotDownloadDir`, `Steam.InstallDir` and `steam.BuildID` refuse an app or depot id that is not decimal digits, returning "" instead of building a path from it.
- `AddShortcut` and `RemoveShortcuts` refuse a shortcut field holding a NUL byte (`steam.ErrNUL`) instead of writing a `shortcuts.vdf` Steam cannot read.
- `steam.OverrideCommand` single-quotes a path with a space or shell syntax.
- `Steam.InstallDir` skips an `installdir` that leaves `steamapps/common`.
- Every launcher and Steam file read is capped (16 MiB, images 32 MiB) and refuses a pipe or device; a binary VDF string is capped at 64 KiB; the `.ico` entry bounds cannot wrap on a 32-bit build.
- A text VDF whose last line is a comment parses instead of hanging.

## [0.3.0] - 2026-11-02

### Changed

- `steam.Depot` carries `id`, `manifest` and `size` JSON tags.

## [0.2.0] - 2026-11-02

### Added

- `steam`: an installed app's depots (`InstalledDepots`) and the folder Steam's console `download_depot` writes a depot to (`DepotDownloadDir`).
- `steam`: find Steam's libraries on macOS, under `~/Library/Application Support/Steam`.

## [0.1.0] - 2026-10-29

### Added

- `steam`: locate native and Flatpak Steam from the usual folders, the registry on Windows and folders the caller
  adds; libraries, installs, build ids, the signed-in account, cached art, client icons and screenshots.
- `steam`: set launch options and add or remove non-Steam shortcuts without disturbing the rest of the file, with a
  required one-time backup suffix.
- `steam`: Proton builds, picking one, the prefix version, and the compatibility tool Steam chose for a game.
- `steam`: grant Flatpak Steam access to a folder.
- `gog`, `heroic`, `epic`, `ea`, `xbox`, `lutris`, `bottles`: install folders for each store and launcher, and
  Heroic's Wine prefixes.
- `gamelocator.GameDir`: confirm an install by a marker file at its root or in its `game` folder.

[Unreleased]: https://github.com/Rethunk-Tech/go-gamelocator/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/Rethunk-Tech/go-gamelocator/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/Rethunk-Tech/go-gamelocator/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/Rethunk-Tech/go-gamelocator/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Rethunk-Tech/go-gamelocator/releases/tag/v0.1.0
