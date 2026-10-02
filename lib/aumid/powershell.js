// Adapter: PowerShell execution via UTF-16LE base64 (-EncodedCommand),
// quoting-proof. Debug logging must never break anything.

import { appendFileSync } from 'node:fs';
import path from 'node:path';
import { spawnSync } from 'node:child_process';

function debugLogFile() {
  return path.join(process.env.TEMP || process.env.TMP || '.', 'opencode-console-notify.debug.log');
}

function dbg(message) {
  try {
    appendFileSync(debugLogFile(), new Date().toISOString() + ' ' + message + '\n');
  } catch {
    // Debug logging must never break anything.
  }
}

function encodePowerShell(script) {
  return Buffer.from(script, 'utf16le').toString('base64');
}

function runEncoded(encoded) {
  const result = spawnSync(
    'powershell.exe',
    ['-NoProfile', '-NonInteractive', '-EncodedCommand', encoded],
    { encoding: 'utf8', windowsHide: true, shell: false },
  );
  if (result.error) {
    throw result.error;
  }
  return result;
}

export {
  debugLogFile,
  dbg,
  encodePowerShell,
  runEncoded,
};
