// Package cli testexport exposes identifiers needed by black-box tests
// living in test/internal/cli. Production code must not depend on this
// file; it exists only so the test package can probe internals without
// being in package cli.

package cli

// ParseArgs exposes parseArgs for black-box tests.
func ParseArgs(args []string) (command string, pluginsDir string, err error) {
	return parseArgs(args)
}

// ParseArgsFull exposes parseArgsFull for black-box tests.
func ParseArgsFull(args []string) (command string, pluginsDir string, rest []string, err error) {
	return parseArgsFull(args)
}
