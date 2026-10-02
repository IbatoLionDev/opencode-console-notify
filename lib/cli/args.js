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

function takePluginsDirFlag(state, args, i) {
  const arg = args[i];
  if (arg === "--plugins-dir" || arg === "-plugins-dir") {
    state.pluginsDir = takePluginsDirValue(args, i);
    return 2;
  }
  const split = splitPluginsDirFlag(arg);
  if (split.found) {
    state.pluginsDir = split.value;
    return 1;
  }
  return 0;
}

// stepConfigArg collects one argument after the config command: plugins-dir
// overrides still apply, everything else is a config arg for scripts.
function stepConfigArg(state, args, i) {
  const consumed = takePluginsDirFlag(state, args, i);
  if (consumed > 0) {
    return consumed;
  }
  state.rest.push(args[i]);
  return 1;
}

// stepArg processes one regular argument and returns how many were consumed.
function stepArg(state, args, i) {
  const consumed = takePluginsDirFlag(state, args, i);
  if (consumed > 0) {
    return consumed;
  }
  const arg = args[i];
  const alias = COMMAND_ALIASES[arg];
  if (alias !== undefined) {
    if (state.command !== "") {
      throw new Error(`unexpected argument ${arg}`);
    }
    state.command = alias;
    return 1;
  }
  if (arg.startsWith("-")) {
    throw new Error(`unknown flag ${arg}`);
  }
  if (state.command === "") {
    state.command = arg;
    return 1;
  }
  throw new Error(`unexpected argument ${arg}`);
}

function buildResult(state) {
  switch (state.command) {
    case "":
      throw new UsageSignal("print usage");
    case "install":
    case "uninstall":
    case "doctor":
    case "test":
    case "upgrade":
    case "version":
    case "help":
      return { command: state.command, pluginsDir: state.pluginsDir };
    case "config":
      return { command: state.command, pluginsDir: state.pluginsDir, rest: state.rest };
    default:
      throw new Error(
        `unknown command "${state.command}" (want install|uninstall|doctor|test|upgrade|version|config)`,
      );
  }
}

function parseArgs(args) {
  const state = { command: "", pluginsDir: "", rest: [] };
  let i = 0;
  while (i < args.length) {
    if (state.command === "config") {
      i += stepConfigArg(state, args, i);
    } else {
      i += stepArg(state, args, i);
    }
  }
  return buildResult(state);
}

export { parseArgs };
