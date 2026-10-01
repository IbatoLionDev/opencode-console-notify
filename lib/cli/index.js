'use strict';

// Mirrors internal/cli/cli.go: the opencode-notify commands (install,
// uninstall, doctor, test) without admin rights.

const aumid = require('../aumid');
const config = require('../config');
const { doctor } = require('../doctor');
const { install, uninstall } = require('../installer');
const { sendTestToast } = require('../notifier');

const usageText = `Usage: opencode-notify [--plugins-dir DIR] <command>

Commands:
  install     Copy the embedded plugin into the plugins directory and
              register the OpenCode.Notifier AUMID (HKCU, no admin).
  uninstall   Remove the plugin file and the AUMID key only.
  doctor      Verify the install end state (plugin file + AUMID).
              Exit 0 when healthy, 1 when something is missing.
  test        Send a real Windows toast notification.

Options:
  --plugins-dir DIR   Override the plugins directory.
                      Default: <HOME>/.config/opencode/plugins
`;

// UsageSignal marks the empty-args case: print usage on stdout, exit 2.
// It mirrors Go's io.EOF signal from parseArgs.
class UsageSignal extends Error {}

// splitPluginsDirFlag extracts the value from the "--plugins-dir=DIR"
// and "-plugins-dir=DIR" equals forms.
function splitPluginsDirFlag(arg) {
  if (arg.startsWith('--plugins-dir=')) {
    return { found: true, value: arg.slice('--plugins-dir='.length) };
  }
  if (arg.startsWith('-plugins-dir=')) {
    return { found: true, value: arg.slice('-plugins-dir='.length) };
  }
  return { found: false, value: '' };
}

// parseArgs extracts the subcommand and the --plugins-dir override.
// The flag is accepted before or after the command, as
// "--plugins-dir DIR" or "--plugins-dir=DIR" (single-dash form too).
function parseArgs(args) {
  let command = '';
  let pluginsDir = '';
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === '--plugins-dir' || arg === '-plugins-dir') {
      if (i + 1 >= args.length) {
        throw new Error(`flag ${arg} needs a value`);
      }
      i++;
      pluginsDir = args[i];
      continue;
    }
    const split = splitPluginsDirFlag(arg);
    if (split.found) {
      pluginsDir = split.value;
      continue;
    }
    if (arg.startsWith('-')) {
      throw new Error(`unknown flag ${arg}`);
    }
    if (command === '') {
      command = arg;
      continue;
    }
    throw new Error(`unexpected argument ${arg}`);
  }
  switch (command) {
    case '':
      throw new UsageSignal('print usage');
    case 'install':
    case 'uninstall':
    case 'doctor':
    case 'test':
    case 'help':
    case '--help':
    case '-h':
      return { command, pluginsDir };
    default:
      throw new Error(`unknown command "${command}" (want install|uninstall|doctor|test)`);
  }
}

// Run executes the CLI; it returns the process exit code so tests can
// assert on it without spawning a subprocess.
function run(args, reg = aumid.defaultRegistry, stdout = process.stdout, stderr = process.stderr) {
  let parsed;
  try {
    parsed = parseArgs(args);
  } catch (err) {
    if (err instanceof UsageSignal) {
      stdout.write(usageText);
      return 2;
    }
    stderr.write(`Error: ${err.message}\n\n${usageText}`);
    return 2;
  }
  if (parsed.command === 'help' || parsed.command === '--help' || parsed.command === '-h') {
    stdout.write(usageText);
    return 0;
  }

  let pluginsDir;
  try {
    pluginsDir = config.resolvePluginsDir(parsed.pluginsDir);
  } catch (err) {
    stderr.write(`Error: ${err.message}\n`);
    return 2;
  }

  switch (parsed.command) {
    case 'install':
      try {
        install(pluginsDir, reg, stdout);
      } catch (err) {
        stderr.write(`Error: ${err.message}\n`);
        return 1;
      }
      return 0;
    case 'uninstall':
      try {
        uninstall(pluginsDir, reg, stdout);
      } catch (err) {
        stderr.write(`Error: ${err.message}\n`);
        return 1;
      }
      return 0;
    case 'doctor':
      return doctor(pluginsDir, reg, stdout);
    case 'test':
      try {
        sendTestToast(reg);
      } catch (err) {
        stderr.write(`Error: ${err.message}\n`);
        return 1;
      }
      stdout.write(`Sent test notification via ${aumid.AUMID}.\n`);
      return 0;
    default:
      stderr.write(`Error: unknown command "${parsed.command}"\n\n${usageText}`);
      return 2;
  }
}

module.exports = {
  parseArgs,
  run,
};
