// Package plugin exposes the opencode JS plugin file as embedded bytes so
// the CLI installs byte-identical content without a network download.
package plugin

import _ "embed"

// Source is the exact content of console-notify.js in this directory.
//
//go:embed console-notify.js
var Source []byte
