# lib/notifier

Sends a real Windows toast for the test command. Node mirror of
`internal/notifier/toast.go` (no new semantics: same XML payload, same
UTF-16LE base64 handoff, same shared AUMID).

## API

- `sendTestToast(reg)` — ensures the AUMID registration (idempotent), then shows one toast via `Windows.UI.Notifications`. Throws an English error on failure.
- `buildToastScript(title, line)` / `escapeToastXml(value)` / `encodePowerShell(script)` — exported building blocks, mirroring the Go helpers.

The child PowerShell runs with `windowsHide: true` and `shell: false`
(Node's `CREATE_NO_WINDOW` equivalent), so the user's console never flashes.
