// Pure domain: CLI argument parsing. No I/O.

import { UsageSignal } from "./usage.js";

function splitPluginsDirFlag(arg) {
  if (arg.startsWith("--plugins-dir=")) {
    return { found: true, value: arg.slice("--plugins-dir=".length) };
  }
  if (arg.startsWith("-plugins-dir=")) {
    return { found: true, value: arg.slice("-plugins-dir=".length) };
  }
  return { found: false, value: "" };
}

// COMMAND_ALIASES maps flag-style spellings to their command so every
// spelling behaves identically (a bare --help used to die as "unknown flag").
const COMMAND_ALIASES = {
  "--help": "help",
  "-h": "help",
  "--version": "version",
  "-V": "version",
};

// takePluginsDirValue returns the value following a --plugins-dir flag.
function takePluginsDirValue(args, i) {
  if (i + 1 >= args.length) {
    throw new Error(`flag ${args[i]} needs a value`);
  }
  return args[i + 1];
}

function parseArgs(args) {
  let command = "";
  let pluginsDir = "";
  let i = 0;
  while (i < args.length) {
    const arg = args[i];
    if (arg === "--plugins-dir" || arg === "-plugins-dir") {
      pluginsDir = takePluginsDirValue(args, i);
      i += 2;
      continue;
    }
    const split = splitPluginsDirFlag(arg);
    if (split.found) {
      pluginsDir = split.value;
      i++;
      continue;
    }
    const alias = COMMAND_ALIASES[arg];
    if (alias !== undefined) {
      if (command !== "") {
        throw new Error(`unexpected argument ${arg}`);
      }
      command = alias;
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
    case "version":
    case "help":
      return { command, pluginsDir };
    default:
      throw new Error(
        `unknown command "${command}" (want install|uninstall|doctor|test|upgrade|version)`,
      );
  }
}

export { parseArgs };
