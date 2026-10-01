// CLI: install | uninstall | doctor | test | upgrade (--plugins-dir override).

import * as aumid from "../aumid/index.js";
import * as config from "../config/index.js";
import { doctor } from "../doctor/index.js";
import { install, uninstall } from "../installer/index.js";
import { sendTestToast } from "../notifier/index.js";
import { upgrade } from "../upgrade/index.js";

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

Options:
  --plugins-dir DIR   Override the plugins directory.
                      Default: <HOME>/.config/opencode/plugins
`;

class UsageSignal extends Error {}

function splitPluginsDirFlag(arg) {
  if (arg.startsWith("--plugins-dir=")) {
    return { found: true, value: arg.slice("--plugins-dir=".length) };
  }
  if (arg.startsWith("-plugins-dir=")) {
    return { found: true, value: arg.slice("-plugins-dir=".length) };
  }
  return { found: false, value: "" };
}

function parseArgs(args) {
  let command = "";
  let pluginsDir = "";
  let i = 0;
  while (i < args.length) {
    const arg = args[i];
    if (arg === "--plugins-dir" || arg === "-plugins-dir") {
      if (i + 1 >= args.length) {
        throw new Error(`flag ${arg} needs a value`);
      }
      pluginsDir = args[i + 1];
      i += 2;
      continue;
    }
    const split = splitPluginsDirFlag(arg);
    if (split.found) {
      pluginsDir = split.value;
      i++;
      continue;
    }
    if (arg.startsWith("-")) {
      throw new Error(`unknown flag ${arg}`);
    }
    if (command === "") {
      command = arg;
      i++;
      continue;
    }
    throw new Error(`unexpected argument ${arg}`);
  }
  switch (command) {
    case "":
      throw new UsageSignal("print usage");
    case "install":
    case "uninstall":
    case "doctor":
    case "test":
    case "upgrade":
    case "help":
    case "--help":
    case "-h":
      return { command, pluginsDir };
    default:
      throw new Error(
        `unknown command "${command}" (want install|uninstall|doctor|test|upgrade)`,
      );
  }
}

// Run executes the CLI and always resolves to the process exit code
// (async so every command, including upgrade, returns the same type).
async function run(
  args,
  reg = aumid.defaultRegistry,
  stdout = process.stdout,
  stderr = process.stderr,
) {
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
  if (
    parsed.command === "help" ||
    parsed.command === "--help" ||
    parsed.command === "-h"
  ) {
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
    case "install":
      try {
        install(pluginsDir, reg, stdout);
      } catch (err) {
        stderr.write(`Error: ${err.message}\n`);
        return 1;
      }
      return 0;
    case "uninstall":
      try {
        uninstall(pluginsDir, reg, stdout);
      } catch (err) {
        stderr.write(`Error: ${err.message}\n`);
        return 1;
      }
      return 0;
    case "doctor":
      return doctor(pluginsDir, reg, stdout);
    case "test":
      try {
        sendTestToast(reg);
      } catch (err) {
        stderr.write(`Error: ${err.message}\n`);
        return 1;
      }
      stdout.write(`Sent test notification via ${aumid.AUMID}.\n`);
      return 0;
    case "upgrade":
      return upgrade(pluginsDir, stdout).then(
        (code) => code,
        (err) => {
          stderr.write(`Error: ${err.message}\n`);
          return 1;
        },
      );
    default:
      stderr.write(
        `Error: unknown command "${parsed.command}"\n\n${usageText}`,
      );
      return 2;
  }
}

export { parseArgs, run };
