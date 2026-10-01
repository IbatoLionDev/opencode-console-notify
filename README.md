# opencode-console-notify

Native Windows toast notifications for OpenCode — know when the agent finishes, errors, or needs you, without watching the terminal.

## What it notifies

| Event | Meaning |
| --- | --- |
| `session.idle` | Task finished, waiting for input |
| `session.error` | Something went wrong |
| `permission.asked` | Permission needed to continue |
| `question.asked` | OpenCode asked a question |

Child (sub-agent) sessions are filtered for `session.idle` / `session.error` so runs
don't spam notifications. Blocking events (`permission.asked`, `question.asked`) always
notify, even from children, because they stop all progress until answered.

## Requirements

- Windows 10 or 11
- OpenCode
- No administrator rights needed (everything is registered per-user in HKCU)

## Install

Three independent paths — pick any one. Each path registers the Windows notification
identity (AUMID) itself, so there are no manual registry steps.

### PowerShell (one line)

```powershell
irm https://raw.githubusercontent.com/IbatoLionDev/opencode-console-notify/main/install.ps1 | iex
```

### Go CLI binary

1. Download `opencode-notify.exe` from the [GitHub Releases page](https://github.com/IbatoLionDev/opencode-console-notify/releases).
2. Run:

```powershell
opencode-notify.exe install
```

### npm / npx (no Go needed)

Zero-dependency, Windows-only:

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

The published package works with any npm-compatible manager:

```powershell
pnpm add -g opencode-console-notify
opencode-console-notify install
```

```powershell
yarn global add opencode-console-notify
opencode-console-notify install
```

```powershell
bun add -g opencode-console-notify
opencode-console-notify install
```

One-off runs without installing:

```powershell
pnpm dlx opencode-console-notify install
bunx opencode-console-notify install
```

### Upgrade

One command per path:

```powershell
opencode-console-notify upgrade
```

Checks the npm registry for a newer release. When the installed copy is
already up to date it prints both versions and changes nothing; otherwise it
installs `opencode-console-notify@latest` globally and re-runs `install` from
the fresh copy. `--plugins-dir DIR` (before or after the command) is
forwarded to that reinstall.

```powershell
.\install.ps1 -Upgrade
```

Skips the local-checkout file and always downloads fresh bytes from GitHub,
then installs exactly like a fresh install (same atomic write, AUMID
registration, and hash output). `-Uninstall` wins when both are set.

```powershell
opencode-notify.exe upgrade
```

Reinstalls the plugin from the copy embedded in the binary and runs `doctor`
to verify. The embedded copy is pinned at build time, so this refreshes the
install but cannot fetch a newer binary — newer `opencode-notify.exe` builds
come from the
[GitHub Releases page](https://github.com/IbatoLionDev/opencode-console-notify/releases).

### Then

Restart OpenCode.

### Options shared by all paths

Custom plugins directory (every path accepts it; the CLI flags work with
`install`, `uninstall`, `doctor`, and `test`):

```powershell
.\install.ps1 -PluginsDir "C:\path\to\plugins"
opencode-notify.exe install --plugins-dir "C:\path\to\plugins"
opencode-console-notify install --plugins-dir "C:\path\to\plugins"
```

Run from a local clone, `.\install.ps1` uses the checkout's `plugin\` file and
downloads nothing.

### Build the CLI from source (optional)

Requires Go 1.27+:

```powershell
go build -o opencode-notify.exe ./cmd/opencode-notify
```

## Verify

```powershell
opencode-notify.exe doctor
# or: opencode-console-notify doctor
```

(or re-run the PowerShell script — all three check the same end state: plugin file present
and AUMID registered).

Send a test toast:

```powershell
opencode-notify.exe test
# or: opencode-console-notify test
```

## Uninstall

```powershell
opencode-notify.exe uninstall
# or: opencode-console-notify uninstall
```

or re-run the PowerShell script with `-Uninstall`.

## Troubleshooting

- **No notifications** — run `opencode-notify.exe doctor`; it reports exactly what is
  missing (plugin file or AUMID registration).
- **Manual AUMID check** — the identity lives at
  `HKCU\Software\Classes\AppUserModelId\OpenCode.Notifier`
  (`DisplayName=OpenCode`). Both installers manage it; you should never need to
  touch it by hand.
- **Focus Assist / Do Not Disturb** — Windows silently suppresses toasts while these
  are enabled. Check Settings > System > Notifications.
- **Debug log** — the plugin writes debug info to
  `%TEMP%\opencode-console-notify.debug.log`.

## Scope / roadmap

v1 is Windows-only. macOS and Linux are not supported yet; they are tracked on the
roadmap and contributions are welcome.

## License

MIT — see [LICENSE](LICENSE).
