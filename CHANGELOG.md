# Changelog

Notable changes to `go-gamelocator`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/Rethunk-Tech/go-gamelocator/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Rethunk-Tech/go-gamelocator/releases/tag/v0.1.0
