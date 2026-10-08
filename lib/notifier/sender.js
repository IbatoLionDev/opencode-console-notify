// Adapter: test-toast delivery. Windows via PowerShell UTF-16LE base64
// (-EncodedCommand), Linux via notify-send argv, both quoting-proof.

import { spawnSync } from 'node:child_process';
import * as aumid from '../aumid/index.js';
import { buildToastScript, buildNotifySendArgs } from './toastScript.js';

function encodePowerShell(script) {
  return Buffer.from(script, 'utf16le').toString('base64');
}

function sendTestToast(reg = aumid.defaultRegistry) {
  reg.ensureAUMID();
  if (process.platform === 'linux') {
    return sendTestToastLinux();
  }
  if (process.platform !== 'win32') {
    throw new Error(`test toast not supported on ${process.platform}`);
  }
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

// sendTestToastLinux shells to notify-send with an exact argv (no
// shell). A missing binary is a clear error pointing at libnotify-bin.
function sendTestToastLinux() {
  const args = buildNotifySendArgs('OpenCode', 'Test notification from opencode-notify');
  const result = spawnSync('notify-send', args, { encoding: 'utf8', shell: false });
  if (result.error) {
    throw new Error(`show test toast: ${result.error.message} (is libnotify-bin installed?)`);
  }
  if (result.status !== 0) {
    const output = `${result.stdout || ''}${result.stderr || ''}`.trim();
    const suffix = output ? `: ${output}` : '';
    throw new Error(`show test toast: notify-send exited ${result.status}${suffix}`);
  }
}

export {
  encodePowerShell,
  sendTestToast,
};
