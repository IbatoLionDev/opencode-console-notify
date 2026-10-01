# lib/config

Resolves where the plugin file lives. Node mirror of `internal/config/paths.go`
(no new semantics: same file name, same default directory, same override rule).

## API

- `PLUGIN_FILE_NAME` — always `console-notify.js`, the only file this CLI writes or removes.
- `defaultPluginsDir()` — `<HOME>/.config/opencode/plugins` via `os.homedir()`; throws when the home directory cannot be determined.
- `resolvePluginsDir(override)` — returns `path.normalize(override)` when it is non-blank after trimming, otherwise `defaultPluginsDir()`.
- `pluginPath(pluginsDir)` — joins the plugins directory with the plugin file name.
