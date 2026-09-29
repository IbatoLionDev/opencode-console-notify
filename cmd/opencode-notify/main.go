// Command opencode-notify installs, checks, and exercises the
// opencode-console-notify Windows toast plugin without admin rights.
package main

import (
	"os"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], aumid.NewRegistry(), os.Stdout, os.Stderr))
}
