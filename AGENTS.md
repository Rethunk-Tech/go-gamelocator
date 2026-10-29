# AGENTS.md

`gamelocator` finds games installed by PC stores and launchers, and edits the Steam files a launcher needs. Launchers,
mod managers and save tools import its store packages. It never imports a consumer, a GUI toolkit, a game catalog or a
process helper: callers describe a game with a marker file and ids, and map the results onto their own types.

## Gate

`gate` is the command. Each override in `.gate.toml` is there because a runner
cannot resolve that check from PATH. The `windows-latest` test job is the only place
Windows-only behaviour runs; the Windows-only tests are the registry readers (`steam` Steam keys, `gog` Galaxy keys, `ea.RegistryDir`,
`xbox.PackageRoot`, `internal/winreg`), which write a layout under `HKCU\Software\go-gamelocator-test` and delete it
afterwards (HKLM is never written), plus the Windows branches of the path tests. The fuzz targets run for 20 seconds each in the Linux test job.

## Shape

- **A caller's roots are extra arguments, searched first.** `Roots`, `LocateAll`, `ProtonBuilds`, `FindProton`,
  `gog.Roots`, `lutris.Locate` and `bottles.Dirs` take the folders a player added, because Windows Steam is found
  through the registry and a custom library lives nowhere a default can guess. Never make a store ignore them.
- **Reads edit nothing, edits keep every other byte.** `setLaunchOptions` splices one value or inserts one block, and
  the shortcuts list keeps file order, so Steam sees only the change.
- **`EditOptions.BackupSuffix` is required.** An implicit default would leave a second backup beside the caller's
  existing one under another name.
- **Depth bounds guard fatal errors.** A stack overflow cannot be recovered, so `parseVDF` counts depth over
  `scanVDF`'s tokens (the tokenizer the edits use, so a comment or quoted brace cannot hide nesting) before the
  third-party parser runs, and `readVDFLevel` stops at 16 levels. Both wrap `ErrTooDeep`.
- **An empty `home` means the user's home, in `steam` only.** `homeOr` resolves it; the other stores take a resolved
  home, because `userdirs` answers `""` with `""`.
- **`CompatTool` does not fall back to `"0"`.** One root's default must not win over another root's per-game entry.
- **`heroic.Prefix` has its own rule, not `wineprefix.Root`.** A `"proton"` wine type means `prefix/pfx` only, any
  other type the prefix only, as Heroic itself does.
- **Windows readers take `(root, basePath)`.** The exported name passes the real root; tests pass a scratch key.
- **`(goos, env)` hooks stay unexported** (`epic.listDirs`, `ea.dirs`) so the Windows-only rule is tested on Linux.
- **Tests use temp homes** and set `HOME` and `USERPROFILE` together, since `userdirs` honours `XDG_*` and `APPDATA`
  only for the process's own home.

## Files

| File | Holds |
| --- | --- |
| `gamelocator.go` | `GameDir`, the marker check every store uses |
| `steam/steam.go` | Roots, `Steam`, libraries, accounts, art, text VDF parsing and its bounds, Flatpak grants |
| `steam/localconfig.go` | `scanVDF`, `setLaunchOptions` and `SetLaunchOptions` |
| `steam/shortcuts.go` | Binary VDF reader and writer, `AddShortcut`, `RemoveShortcuts` |
| `steam/grid.go`, `icon.go` | Grid art for new shortcuts, the client icon |
| `steam/write.go` | `EditOptions`, the error sentinels, the one-time backup |
| `steam/proton.go` | Proton builds and ranking, `LibraryOf`, `PrefixVersion`, `CompatTool` |
| `steam/locate_unix.go`, `locate_windows.go` | The usual Steam folders, and the registry keys on Windows |
| `gog/gog.go` | `Locate`, `Source`, GOG games folders, Minigalaxy |
| `gog/galaxy_unix.go`, `galaxy_windows.go` | Galaxy's folders and registry entries |
| `heroic/heroic.go`, `epic.go`, `prefix.go` | Heroic config folders, GOG and Epic lists, Wine prefix |
| `epic/epic.go` | The Epic Games Launcher install list |
| `ea/ea.go`, `registry_*.go` | EA App folders, installer registry entries |
| `xbox/xbox.go`, `package_*.go` | Xbox library folder, AppX package roots |
| `lutris/lutris.go` | Lutris config matching and `Locate` |
| `bottles/bottles.go` | Bottles folders, bottle list, game folders, `Of` |
| `internal/winreg`, `winregtest` | One registry string read; a scratch HKCU tree for tests |
| `internal/wineprefix/wineprefix.go` | The folder holding `drive_c` |
