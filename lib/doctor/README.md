# lib/doctor

Parity gate: verifies the SAME end state as install.ps1 (plugin file present
+ AUMID registered) regardless of which install path produced it. Node mirror
of `internal/doctor/doctor.go` (no new semantics).

## API

- `checkHealth(pluginsDir, reg)` — returns `{ pluginsDir, pluginPath, pluginOK, pluginNote, aumidOK, aumidNote }`. Missing state is reported, never thrown.
- `healthy(report)` — true only when both halves of the end state hold.
- `doctor(pluginsDir, reg, out)` — prints the `Plugin file` / `AUMID registration` lines plus the healthy/issues verdict; returns `0` when healthy, `1` otherwise.
