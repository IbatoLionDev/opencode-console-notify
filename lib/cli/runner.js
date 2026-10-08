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
  const parsed = parseCliArgs(args, stdout, stderr);
  if (parsed === null) {
    return 2;
  }
  if (parsed.command === "help") {
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

  const handler = COMMAND_HANDLERS[parsed.command];
  if (handler === undefined) {
    stderr.write(
      `Error: unknown command "${parsed.command}"\n\n${usageText}`,
    );
    return 2;
  }
  return await handler({ pluginsDir, parsed, reg, stdout, stderr, stdin });
}

// parseCliArgs parses argv or returns null when usage was already printed
// (exit code is always 2 in that path).
function parseCliArgs(args, stdout, stderr) {
  let parsed;
  try {
    parsed = parseArgs(args);
  } catch (err) {
    if (err instanceof UsageSignal) {
      stdout.write(usageText);
      return null;
    }
    stderr.write(`Error: ${err.message}\n\n${usageText}`);
    return null;
  }
  if (
    parsed.command === "help" ||
    parsed.command === "--help" ||
    parsed.command === "-h"
  ) {
    return { command: "help", pluginsDir: parsed.pluginsDir };
  }
  return parsed;
}

function writeCommandError(stderr, err) {
  stderr.write(`Error: ${err.message}\n`);
  return 1;
}

async function handleInstall({ pluginsDir, reg, stdout, stderr }) {
  try {
    install(pluginsDir, reg, stdout);
  } catch (err) {
    return writeCommandError(stderr, err);
  }
  return 0;
}

async function handleUninstall({ pluginsDir, reg, stdout, stderr }) {
  try {
    uninstall(pluginsDir, reg, stdout);
  } catch (err) {
    return writeCommandError(stderr, err);
  }
  return 0;
}

async function handleDoctor({ pluginsDir, reg, stdout }) {
  return await doctor(pluginsDir, reg, stdout);
}

async function handleTest({ reg, stdout, stderr }) {
  try {
    sendTestToast(reg);
  } catch (err) {
    return writeCommandError(stderr, err);
  }
  stdout.write(`Sent test notification via ${testViaName()}.\n`);
  return 0;
}

// testViaName names the delivery backend in the test command output
// for this platform.
function testViaName(platform = process.platform) {
  if (platform === 'linux') {
    return 'notify-send';
  }
  return aumid.AUMID;
}

async function handleUpgrade({ pluginsDir, stdout, stderr }) {
  try {
    return await upgrade(pluginsDir, stdout);
  } catch (err) {
    return writeCommandError(stderr, err);
  }
}

async function handleVersion({ stdout, stderr }) {
  let version;
  try {
    version = readLocalVersion();
  } catch (err) {
    return writeCommandError(stderr, err);
  }
  stdout.write(`opencode-console-notify ${version}\n`);
  return 0;
}

async function handleConfig({ pluginsDir, parsed, stdout, stderr, stdin }) {
  const rest = Array.isArray(parsed.rest) ? parsed.rest : [];
  const upgradeHere = (out) => upgrade(pluginsDir, out || stdout);
  try {
    return await runConfig(pluginsDir, rest, upgradeHere, stdin, stdout);
  } catch (err) {
    return writeCommandError(stderr, err);
  }
}

const COMMAND_HANDLERS = {
  install: handleInstall,
  uninstall: handleUninstall,
  doctor: handleDoctor,
  test: handleTest,
  upgrade: handleUpgrade,
  version: handleVersion,
  config: handleConfig,
};

export { run };
