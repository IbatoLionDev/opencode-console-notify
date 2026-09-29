package cli

// plugin.go re-exports the embedded plugin bytes. The bytes live in the
// plugin package so the go:embed pattern stays next to the JS file
// (patterns cannot contain "..").

import "github.com/IbatoLionDev/opencode-console-notify/plugin"

// pluginSource is the exact content of plugin/console-notify.js.
var pluginSource = plugin.Source
