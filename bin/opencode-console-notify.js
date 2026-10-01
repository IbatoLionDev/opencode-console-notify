#!/usr/bin/env node

// Entry point: all logic lives in lib/cli. run() is async and always
// resolves to the exit code.
import { run } from '../lib/cli/index.js';

try {
  process.exitCode = await run(process.argv.slice(2));
} catch (err) {
  console.error(`Error: ${err?.message ?? err}`);
  process.exitCode = 1;
}
