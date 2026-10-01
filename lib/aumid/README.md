# lib/aumid

Owns the Windows notification identity (AUMID) registry boundary. Node mirror
of `internal/aumid/registry.go` + `registry_windows.go` and the plugin's
`ensureAumidRegistered()` (no new semantics: same key, same idempotency).

## API

- `AUMID` — `OpenCode.Notifier`, shared by the plugin, install.ps1, and this CLI.
- `AUMID_DISPLAY_NAME` — `OpenCode`, the friendly name stored on the key.
- `AUMID_KEY_PATH` — HKCU-relative path of the exact key this CLI owns.
- `ensureAUMID()` — creates the key and DisplayName when missing; re-running is a no-op. Never throws: failures go to `%TEMP%/opencode-console-notify.debug.log`, like the plugin.
- `aumidStatus()` — returns `{ registered, detail }`; throws only on genuine probe failures.
- `removeAUMID()` — deletes exactly the AUMID key; returns `true` when it existed, `false` when already absent.
- `defaultRegistry` — production HKCU backend (mirrors `aumid.NewRegistry()`).

All operations spawn `powershell.exe -NoProfile -NonInteractive -EncodedCommand`
with a UTF-16LE base64 payload, so quoting can never break. HKCU only, no admin.
