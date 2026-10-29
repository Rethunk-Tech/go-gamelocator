<h1 align="center">go-gamelocator</h1>

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.27.3+-00ADD8?logo=go&logoColor=white)](go.mod)

</div>

---

Finds games installed by PC stores and launchers on Linux, Windows and Steam Deck, and reads and edits the Steam files a launcher needs. Package `gamelocator`, module `github.com/Rethunk-Tech/go-gamelocator`.

Every launcher keeps its own list of installed games in its own format. This module reads those lists and answers one
question: where is this game, per store. It reads files, never launches anything, and never contacts a network.

## Quick start

```bash
go get github.com/Rethunk-Tech/go-gamelocator
```

A worked example and the contracts: [HUMANS.md](HUMANS.md).

## Features

- **One package per store.** `steam`, `gog`, `heroic`, `epic`, `ea`, `xbox`, `lutris` and `bottles`, each returning
  install folders, with the folders the player added searched first.
- **Steam in depth.** Native and Flatpak installs, library folders, build ids, the signed-in account, cached art,
  screenshots, Proton builds and the compatibility tool a game uses.
- **Byte-preserving edits.** `SetLaunchOptions`, `AddShortcut` and `RemoveShortcuts` change only what they mean to,
  and take a one-time backup you name with `EditOptions`. Steam must be closed while they run.
- **Bounded parsers.** Both VDF readers refuse files nested deeper than any Steam writes (`ErrTooDeep`).
- **Linux, Windows and Steam Deck.** Windows registry reads run on a Windows CI leg; Flatpak sandbox grants for Steam
  need the `--talk-name=org.freedesktop.Flatpak` permission in your own manifest.

## Documentation

| Doc | Audience |
| --- | --- |
| [HUMANS.md](HUMANS.md) | Running and using the package |
| [AGENTS.md](AGENTS.md) | Internals, invariants, and why |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Process — commits, tests, docs |
| [SECURITY.md](SECURITY.md) | Vulnerability disclosure |
| [CHANGELOG.md](CHANGELOG.md) | Release notes |

## License

MIT — see [LICENSE](LICENSE). © 2026 Rethunk.Tech, LLC.
