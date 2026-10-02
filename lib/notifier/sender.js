// Adapter: test-toast delivery via PowerShell UTF-16LE base64
// (-EncodedCommand), quoting-proof.

import { spawnSync } from 'node:child_process';
import * as aumid from '../aumid/index.js';
import { buildToastScript } from './toastScript.js';

function encodePowerShell(script) {
  return Buffer.from(script, 'utf16le').toString('base64');
}

function sendTestToast(reg = aumid.defaultRegistry) {
  reg.ensureAUMID();
  const script = buildToastScript('OpenCode', 'Test notification from opencode-notify');
  const encoded = encodePowerShell(script);
  // windowsHide (Win32 CREATE_NO_WINDOW) is load-bearing, NOT cosmetic:
  // without it PowerShell inherits our console window and -WindowStyle
  // Hidden would hide the USER's console. shell:false keeps the argument
  // vector exact. The synchronous wait mirrors Go's CombinedOutput.
  const result = spawnSync(
    'powershell.exe',
    ['-NoProfile', '-NonInteractive', '-EncodedCommand', encoded],
    { encoding: 'utf8', windowsHide: true, shell: false },
  );
  if (result.error) {
    throw new Error(`show test toast: ${result.error.message}`);
  }
  if (result.status !== 0) {
    const output = `${result.stdout || ''}${result.stderr || ''}`.trim();
    const suffix = output ? `: ${output}` : '';
    throw new Error(`show test toast: powershell exited ${result.status}${suffix}`);
  }
}

export {
  encodePowerShell,
  sendTestToast,
};
