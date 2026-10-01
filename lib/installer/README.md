# lib/installer

Implements the install and uninstall commands. Node mirror of
`internal/installer/install.go` (no new semantics: same atomic write,
same output lines, same narrow uninstall scope).

## API

- `install(pluginsDir, reg, out)` — creates the plugins directory, copies `plugin/console-notify.js` via temp-file-plus-rename, ensures the AUMID key, and prints the `Installed` / `SHA256` / `Source` / `AUMID` lines.
- `uninstall(pluginsDir, reg, out)` — removes only the plugin file and exactly the AUMID key, never parent keys or anything else in the directory.
