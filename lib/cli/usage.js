// Pure domain: CLI usage text and the usage-signal sentinel.
// No I/O, no imports.

const usageText = `Usage: opencode-notify [--plugins-dir DIR] <command>

Commands:
  install     Copy the embedded plugin into the plugins directory and
              register the OpenCode.Notifier AUMID (HKCU, no admin).
  uninstall   Remove the plugin file and the AUMID key only.
  doctor      Verify the install end state (plugin file + AUMID).
              Exit 0 when healthy, 1 when something is missing.
  test        Send a real Windows toast notification.
  upgrade     Check the npm registry for a newer release; when newer,
              install it globally and reinstall the plugin from it.
  config      Open the config TUI (language, update, info, alerts,
              customs). Flags for scripts: --lang en|es,
              --toggle KEY=on|off, --list-alerts, --list-events,
              --list-customs, --info.
  version     Print the installed package version.

Options:
  --plugins-dir DIR   Override the plugins directory.
                      Default: <HOME>/.config/opencode/plugins
  --version, -V       Print the version and exit (same as version).
`;

class UsageSignal extends Error {}

export { usageText, UsageSignal };
