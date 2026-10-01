# Changelog

All notable changes to this project are documented here, newest first.
Version headers link to the GitHub release when one exists.

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
