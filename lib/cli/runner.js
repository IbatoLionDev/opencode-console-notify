// Adapter: CLI command orchestration (writes to stdout/stderr, delegates
// to domain adapters). Parsing and usage text live in args.js/usage.js.

import * as aumid from "../aumid/index.js";
import * as config from "../config/index.js";
import { doctor } from "../doctor/index.js";
import { install, uninstall } from "../installer/index.js";
import { sendTestToast } from "../notifier/index.js";
import { readLocalVersion, upgrade } from "../upgrade/index.js";
import { usageText, UsageSignal } from "./usage.js";
import { parseArgs } from "./args.js";
import { runConfig } from "./config.js";

// Run executes the CLI and always resolves to the process exit code
// (async so every command, including upgrade, returns the same type).
async function run(
  args,
  reg = aumid.defaultRegistry,
  stdout = process.stdout,
  stderr = process.stderr,
  stdin = process.stdin,
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
      return await doctor(pluginsDir, reg, stdout);
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
    case "version": {
      let version;
      try {
        version = readLocalVersion();
      } catch (err) {
        stderr.write(`Error: ${err.message}\n`);
        return 1;
      }
      stdout.write(`opencode-console-notify ${version}\n`);
      return 0;
    }
    case "config": {
      const rest = Array.isArray(parsed.rest) ? parsed.rest : [];
      const upgradeHere = (out) => upgrade(pluginsDir, out || stdout);
      try {
        return await runConfig(pluginsDir, rest, upgradeHere, stdin, stdout);
      } catch (err) {
        stderr.write(`Error: ${err.message}\n`);
        return 1;
      }
    }
    default:
      stderr.write(
        `Error: unknown command "${parsed.command}"\n\n${usageText}`,
      );
      return 2;
  }
}

export { run };
