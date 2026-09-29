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

Two independent paths — pick either one. Each path registers the Windows notification
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

### Then

Restart OpenCode.

## Verify

```powershell
opencode-notify.exe doctor
```

(or re-run the PowerShell script — both check the same end state: plugin file present
and AUMID registered).

Send a test toast:

```powershell
opencode-notify.exe test
```

## Uninstall

```powershell
opencode-notify.exe uninstall
```

or re-run the PowerShell script with `-Uninstall`.

## Troubleshooting

- **No notifications** — run `opencode-notify.exe doctor`; it reports exactly what is
  missing (plugin file or AUMID registration).
- **Focus Assist / Do Not Disturb** — Windows silently suppresses toasts while these
  are enabled. Check Settings > System > Notifications.
- **Debug log** — the plugin writes debug info to
  `%TEMP%\opencode-console-notify.debug.log`.

## Scope / roadmap

v1 is Windows-only. macOS and Linux are not supported yet; they are tracked on the
roadmap and contributions are welcome.

## License

MIT — see [LICENSE](LICENSE).
