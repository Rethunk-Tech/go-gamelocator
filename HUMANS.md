# HUMANS.md

Run and use `go-gamelocator`. Internals: [AGENTS.md](AGENTS.md).

## Quick start

```bash
go get github.com/Rethunk-Tech/go-gamelocator
```

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Rethunk-Tech/go-gamelocator/gog"
	"github.com/Rethunk-Tech/go-gamelocator/steam"
)

func main() {
	// A scratch home: a real program passes the user's home folder here.
	home, err := os.MkdirTemp("", "gamelocator-home")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(home) }()

	// A Steam folder the player added, searched before the usual ones.
	added := filepath.Join(home, "Games", "Steam")
	if err := os.MkdirAll(filepath.Join(added, "steamapps"), 0o750); err != nil {
		panic(err)
	}
	st, status := steam.Locate(home, added)
	fmt.Println(status, st.Root == added)

	// Edits need a signed-in account, and Steam closed; this scratch Steam has no loginusers.vdf.
	_, err = st.SetLaunchOptions("413150", func(current string) string { return "-x " + current },
		steam.EditOptions{BackupSuffix: ".bak"})
	fmt.Println(err != nil)

	game := gog.Game{ProductID: "1453375253", Folder: "Stardew Valley", Marker: "Stardew Valley.dll"}
	fmt.Println(len(gog.Locate(home, game, gog.Roots{})))
}
```

## Contracts

| Entry point | Contract |
| --- | --- |
| `steam.Roots`, `LocateAll`, `Locate`, `ProtonBuilds`, `FindProton` | Extra roots come first, then the usual folders, native before Flatpak. A root without `steamapps` is skipped. An empty `home` is the user's home. |
| `Steam.Libraries`, `InstallDir`, `BuildID` | An app id that would leave its folder finds nothing. `InstallDir` is `""`, not an error, for an app that is not installed. |
| `Steam.CurrentAccount` | The `MostRecent` login, else the newest sign-in (ties to the lower id), else `ErrNoAccount`. |
| `Steam.SetLaunchOptions`, `AddShortcut`, `RemoveShortcuts` | `EditOptions.BackupSuffix` is required: `""` fails before any file is touched. The first edit of a file copies it to `path+suffix`, once, never again. Writes are atomic. Steam must be closed. |
| `AddShortcut` | Matches an existing entry by exe and `Key` and updates it in place, keeping its `appid`, playtime and tags. An unchanged entry writes nothing. |
| `Steam.CompatTool` | The exact entry for the app id, with no fallback to `"0"`: ask for the game, then `"0"`, across all roots. |
| `steam.ShowOverride`, `GrantFilesystem` | Run `flatpak override --user` on the host. They need the Flatpak permission `--talk-name=org.freedesktop.Flatpak` when the caller is itself sandboxed. |
| VDF parsers | A file past 8 MiB or nested past 32 text levels (16 binary) fails with an error wrapping `ErrTooDeep`. Nothing recurses unbounded. |
| `gog`, `lutris`, `bottles`, `heroic` | `home` must be resolved (not `""`). Folders in the caller's roots come first. Lutris and Bottles find nothing off Linux. |
| `epic.ListDirs`, `ea.Dirs` | Windows only; elsewhere they return just the caller's folders. |
| `ea.RegistryDir`, `xbox.PackageRoot`, `gog.GalaxyDir` | Read the Windows registry; `""` off Windows or when absent. |
| `heroic.Prefix` | `""` unless `appName` is one plain file name, so it cannot leave `GamesConfig`. |
| Every function | Reads files only. None launches a process but the two `flatpak override` calls, and none opens a network connection. |

## Requirements

Go 1.27.3 or newer, plus `lefthook` and `gitleaks` on PATH if you want the git hooks.

```bash
lefthook install
```

That is per clone; git does not commit hooks. Pre-commit scans staged changes for secrets, and
pre-push runs `gate`.

## Verify

```bash
gate
```
