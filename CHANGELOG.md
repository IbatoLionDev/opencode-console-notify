# Changelog

All notable changes to this project are documented here, newest first.
Version headers link to the GitHub release when one exists.

## [2.2.0] - 2026-10-04

### Added

- Custom notifications: create your own toasts from the config TUI by
  picking one of 48 subscribable OpenCode events and writing the text. New
  customs panel with list, create, edit, delete, toggle and a read-only
  events catalog (noisy families flagged); customs appear in alerts tagged
  as custom and toggle like defaults; the plugin fires them with the same
  child-session rule. Script flags: `--list-events`, `--list-customs`.
  Zero new dependencies.

## [2.1.2] - 2026-10-04

### Fixed

- Switching to a shorter fullscreen view no longer leaves ghost lines (or
  title tails) from the previous one: every frame starts with Home plus
  erase-below, in both runtimes.

## [2.1.1] - 2026-10-03

### Fixed

- Fullscreen rows now end on the same column: the dim detail line measured
  the ANSI dim code as visible text (background ended 4 cells short) and
  padding byte-counted accented runes. Padding is rune-aware in both
  runtimes, covered by a Go regression test and a JS probe.

## [2.1.0] - 2026-10-03

### Added

- `config` TUI is now fullscreen on a real console (Go and npm,
  English/Spanish): in-place redraw with j/k/arrows navigation, Enter
  select, Space toggle, scrollable viewport, and the black/red/white
  palette. On pipes and scripts the previous line mode runs unchanged, as
  do all flags (`--lang`, `--toggle`, `--list-alerts`, `--info`). Zero new
  dependencies. Mouse click support is planned for 2.2.

## [2.0.0] - 2026-10-02

### Added

- `config` command (Go and npm, English/Spanish): interactive TUI with
  language (applies to the TUI and default notification texts), update
  (reuses the existing `upgrade` flow per distribution), info (commands
  plus npm/repo/Releases links), and alerts (the four default events with
  their triggers, toggleable with Space). Every view prints its keys in a
  visible footer. Script flags: `--lang en|es`, `--toggle KEY=on|off`,
  `--list-alerts`, `--info`. Settings persist in
  `console-notify.config.json` next to the plugin file; missing file keeps
  v1 behavior (all alerts on, English). Zero new dependencies.

## [1.3.1] - 2026-10-02

Internal restructuring under clean-architecture layering; no behavior,
message, flag, or exit-code changes.

## [1.3.0] - 2026-10-01

### Added

- `doctor` (npm and Go) prints an automatic update notice
  (`Update available: <local> -> <latest>`) when the npm registry reports a
  newer release. The check is best-effort with a short timeout: offline or
  unreachable registries stay silent and never change the exit code.

## [1.2.1] - 2026-10-01

### Added

- npm: `opencode-console-notify version` (also `--version` / `-V`) prints the
  installed package version, so a stale global install is diagnosable.

## [1.2.0] - 2026-10-01

Upgrade on all three install paths, zero new dependencies.

### Added

- npm: `opencode-console-notify upgrade` — compares the local version against
  the npm registry; when a newer release exists it installs
  `opencode-console-notify@latest` globally and re-runs `install` from the
  fresh copy (`--plugins-dir DIR` is forwarded). Already up to date prints
  both versions and changes nothing.
- PowerShell: `.\install.ps1 -Upgrade` — skips the local-checkout file and
  always downloads fresh bytes from GitHub, then installs exactly like a
  fresh install (same atomic write, AUMID registration, and hash output).
  `-Uninstall` wins when both switches are set.
- Go: `opencode-notify.exe upgrade` — reinstalls the plugin from the embedded
  copy and runs `doctor` to verify. The copy is build-pinned, so the output
  points at the GitHub Releases page for newer binaries.

## [1.1.0] - 2026-10-01

Third install path: zero-dependency npm/npx CLI — no Go needed, Windows-only.

```powershell
npm i -g opencode-console-notify
opencode-console-notify install
```

Or without installing anything:

```powershell
npx opencode-console-notify install
npx opencode-console-notify doctor
npx opencode-console-notify test
npx opencode-console-notify uninstall
```

### Added

- `bin/` + `lib/` mirror folders (`config`, `aumid`, `installer`, `doctor`,
  `notifier`, `cli`) — same behavior as the Go CLI, zero dependencies.
- Parity `doctor`: verifies the same end state (plugin file present + AUMID
  registration) regardless of which install path produced it.
- `--plugins-dir` override on every npm CLI command
  (`install` | `uninstall` | `doctor` | `test`).

## [1.0.0] - 2026-09-29

First public release: native Windows toast notifications for OpenCode.
No administrator rights needed — everything is registered per-user in HKCU.

### Install (pick either path)

One-line PowerShell (downloads nothing else):

```powershell
irm https://raw.githubusercontent.com/IbatoLionDev/opencode-console-notify/main/install.ps1 | iex
```

Go CLI binary (`opencode-notify.exe`, attached to the
[1.0.0 release](https://github.com/IbatoLionDev/opencode-console-notify/releases/tag/v1.0.0)):

```powershell
opencode-notify.exe install
```

Both paths leave the same verifiable state (plugin file + AUMID
registration); `opencode-notify.exe doctor` checks it either way.

### Commands

`install` | `uninstall` | `doctor` | `test` — every command accepts
`--plugins-dir` to override the plugins directory
(default `<HOME>/.config/opencode/plugins`).

### What it notifies

- `session.idle` — task finished, waiting for input
- `session.error` — something went wrong
- `permission.asked` / `question.asked` — OpenCode is blocked and needs you
  (these always notify, even from child sessions)

Child sub-agent sessions are filtered for idle/error so runs never spam.
A 4-second per-kind cooldown keeps repeated events to a single toast.

### Contents

- `plugin/console-notify.js` — the OpenCode plugin (auto-registers the
  `OpenCode.Notifier` AUMID on load, XML 1.0-safe payloads)
- Go CLI (`cmd/` + `internal/` domain packages, unit-tested) with the plugin
  embedded, so installs work fully offline
- `install.ps1` — standalone PowerShell installer (`-PluginsDir`, `-Uninstall`)
- CI on `windows-latest`: Go build/vet/gofmt/tests + `install.ps1` smoke test

### Requirements

- Windows 10 or 11, OpenCode, Go 1.27+ (only to build from source)
- v1 is Windows-only; macOS/Linux are roadmap
