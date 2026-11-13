# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's private vulnerability reporting on this repository: <https://github.com/Rethunk-Tech/go-gamelocator/security/advisories/new>, or the **Security** tab, then **Report a vulnerability**. The report stays private between you and us until a fix ships.

If private reporting is unavailable, email [oss@rethunk.tech](mailto:oss@rethunk.tech) with **[SECURITY]** in the subject.

Please do not open a public issue, pull request or discussion for a suspected vulnerability.

A useful report has the affected repository and version (or commit), steps to reproduce, what an attacker gets and under which conditions, and whether anyone else knows.

## What counts as a vulnerability

A vulnerability is a flaw that lets someone cross a boundary they should not: run code, read or change data they have no right to, escape a sandbox or path restriction, or leak a credential. A bug is anything else that misbehaves, such as a crash on your own input, wrong output or a slow path; open an ordinary issue for those. If you are unsure which one you have, report it privately and we will say.

## Coordinated disclosure

We handle the report inside the private advisory, so you see each status change and can comment on the fix. The advisory goes public when the fix ships, or 90 days after your report, whichever comes first. If a fix is close, either of us can ask for more time in the thread. You are credited in the advisory unless you would rather not be.

## Response timelines

- We acknowledge your report within 3 business days.
- We give you an assessment (accepted, needs more information or declined, with reasons) within 10 business days.
- We aim to ship a fix within 90 days of the report.

## What is in scope

- **Steam's text VDF files** (`libraryfolders.vdf`, `loginusers.vdf`, `localconfig.vdf`, `config.vdf`, app
  manifests), read by `Libraries`, `InstallDir`, `BuildID`, `CurrentAccount`, `LaunchOptions` and `CompatTool`. A file
  must not cause a fatal stack overflow or an allocation past 8 MiB: depth past 32 and size past the cap return errors.
- **`shortcuts.vdf`**, binary, read by `AddShortcut` and `RemoveShortcuts`. Nesting past 16 returns `ErrTooDeep`, and
  any accepted file re-encodes to the bytes read (fuzzed). A string over 64 KiB is refused, and a shortcut field with a NUL byte is refused on write.
- **App ids and Heroic app names** are joined into paths: `HeroArt`, `ClientIcon` and `BuildID` refuse an id that
  is not decimal digits (`numericID`), and `heroic.Prefix` refuses anything but one plain file name.
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

## Safe harbour

We will not pursue legal action against good-faith research that avoids privacy violations, data destruction and disruption of services, stops at a proof of concept, and keeps the details private until the advisory is published. Rethunk.Tech, LLC maintains this repository.
