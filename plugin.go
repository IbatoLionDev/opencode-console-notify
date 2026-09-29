package main

// plugin.go embeds the plugin file so the CLI installs byte-identical
// content without a network download.

import _ "embed"

// pluginSource is the exact content of plugin/console-notify.js.
//
//go:embed plugin/console-notify.js
var pluginSource []byte
