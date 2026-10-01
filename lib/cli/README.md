# lib/cli

Command surface of the npm install path. Node mirror of `internal/cli/cli.go`
(no new semantics: same commands, same flag forms, same exit codes).

## API

- `parseArgs(args)` — extracts `{ command, pluginsDir }`. Accepts `install | uninstall | doctor | test | help | --help | -h`; `--plugins-dir DIR`, `--plugins-dir=DIR`, and the single-dash forms, before or after the command. Unknown flags and unknown commands throw; empty args throw a usage signal.
- `run(args, reg, stdout, stderr)` — executes the CLI and returns the exit code: `2` for usage/errors, `1` for failed operations or unhealthy doctor, `0` on success.
