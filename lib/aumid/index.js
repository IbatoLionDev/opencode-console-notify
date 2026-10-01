'use strict';

// Mirrors internal/aumid/registry.go + registry_windows.go and the plugin's
// ensureAumidRegistered(): owns the HKCU AUMID registry key shared by the
// plugin, install.ps1, and this CLI. HKCU only, no admin rights needed.
//
// Node has no dependency-free registry API, so every operation shells out to
// powershell.exe with -EncodedCommand (UTF-16LE base64), the same
// quoting-proof technique the Go CLI and the plugin use.

const { appendFileSync } = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');

// AUMID is the Windows notification identity shared by the plugin,
// install.ps1, and this CLI.
const AUMID = 'OpenCode.Notifier';

// AUMID_DISPLAY_NAME is the friendly name stored on the AUMID key.
const AUMID_DISPLAY_NAME = 'OpenCode';

// AUMID_KEY_PATH is the HKCU-relative path of the exact key this CLI owns.
// Uninstall removes this key only, never its parents.
const AUMID_KEY_PATH = 'Software\\Classes\\AppUserModelId\\OpenCode.Notifier';

// AUMID_PS_PATH is the PowerShell-drive form of the same key.
const AUMID_PS_PATH = 'HKCU:\\Software\\Classes\\AppUserModelId\\OpenCode.Notifier';

function debugLogFile() {
  // Same location the plugin uses for its debug log.
  return path.join(process.env.TEMP || process.env.TMP || '.', 'opencode-console-notify.debug.log');
}

function dbg(message) {
  try {
    appendFileSync(debugLogFile(), new Date().toISOString() + ' ' + message + '\n');
  } catch {
    // Debug logging must never break anything.
  }
}

// encodePowerShell encodes a script as UTF-16LE base64 for
// powershell.exe -EncodedCommand, avoiding every shell-quoting pitfall.
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

// EnsureAUMID creates the AUMID key and DisplayName value when missing.
// The existence guards make a re-run a no-op, exactly like install.ps1
// and the plugin's ensureAumidRegistered(). It never throws: failures are
// appended to the debug log, mirroring the plugin's fire-and-forget style.
function ensureAUMID() {
  if (process.platform !== 'win32') {
    return;
  }
  const script = [
    `if (-not (Test-Path -LiteralPath '${AUMID_PS_PATH}')) { New-Item -Path '${AUMID_PS_PATH}' -Force | Out-Null }`,
    `$d = Get-ItemProperty -LiteralPath '${AUMID_PS_PATH}' -Name DisplayName -ErrorAction SilentlyContinue`,
    `if ($null -eq $d) { Set-ItemProperty -LiteralPath '${AUMID_PS_PATH}' -Name DisplayName -Value '${AUMID_DISPLAY_NAME}' -Type String }`,
  ].join('\n');
  try {
    const result = runEncoded(encodePowerShell(script));
    if (result.status !== 0) {
      dbg(`aumid FAIL key=${AUMID_PS_PATH} exit=${result.status}`);
    } else {
      dbg(`aumid OK key=${AUMID_PS_PATH}`);
    }
  } catch (err) {
    dbg(`aumid THREW key=${AUMID_PS_PATH} err=${String((err && err.message) || err)}`);
  }
}

// AUMIDStatus reports whether the AUMID end state is present: the exact
// key exists and carries a DisplayName value.
function aumidStatus() {
  const script = [
    `$k = '${AUMID_PS_PATH}'`,
    `if (-not (Test-Path -LiteralPath $k)) { Write-Output 'ABSENT'; exit 0 }`,
    `$v = (Get-ItemProperty -LiteralPath $k -Name DisplayName -ErrorAction SilentlyContinue).DisplayName`,
    `if ([string]::IsNullOrEmpty($v)) { Write-Output 'NODISPLAY'; exit 0 }`,
    `Write-Output ('PRESENT:' + $v)`,
  ].join('\n');
  const result = runEncoded(encodePowerShell(script));
  if (result.status !== 0) {
    throw new Error(`powershell exited ${result.status}: ${(result.stderr || '').trim()}`);
  }
  const lines = (result.stdout || '').split('\n').map((line) => line.trim()).filter(Boolean);
  const marker = lines[lines.length - 1] || '';
  if (marker === 'ABSENT') {
    return { registered: false, detail: 'AUMID registry key is missing' };
  }
  if (marker === 'NODISPLAY') {
    return { registered: false, detail: 'AUMID registry key exists but DisplayName is missing' };
  }
  if (marker.startsWith('PRESENT:')) {
    const value = marker.slice('PRESENT:'.length);
    return {
      registered: true,
      detail: `registered (HKCU:\\Software\\Classes\\AppUserModelId\\${AUMID}, DisplayName=${value})`,
    };
  }
  throw new Error(`unexpected AUMID probe output: ${marker}`);
}

// RemoveAUMID deletes exactly the AUMID key, never its parent keys.
// A missing key is not an error.
function removeAUMID() {
  const script = [
    `$k = '${AUMID_PS_PATH}'`,
    `if (Test-Path -LiteralPath $k) { Remove-Item -LiteralPath $k -Recurse -Force; Write-Output 'REMOVED'; exit 0 }`,
    `Write-Output 'ABSENT'`,
  ].join('\n');
  const result = runEncoded(encodePowerShell(script));
  if (result.status !== 0) {
    throw new Error(`remove AUMID key: powershell exited ${result.status}: ${(result.stderr || '').trim()}`);
  }
  const lines = (result.stdout || '').split('\n').map((line) => line.trim()).filter(Boolean);
  const marker = lines[lines.length - 1] || '';
  if (marker === 'REMOVED') {
    return true;
  }
  if (marker === 'ABSENT') {
    return false;
  }
  throw new Error(`unexpected AUMID remove output: ${marker}`);
}

// defaultRegistry is the production HKCU backend, mirroring
// aumid.NewRegistry() on the Go side.
const defaultRegistry = {
  ensureAUMID,
  aumidStatus,
  removeAUMID,
};

module.exports = {
  AUMID,
  AUMID_DISPLAY_NAME,
  AUMID_KEY_PATH,
  AUMID_PS_PATH,
  encodePowerShell,
  ensureAUMID,
  aumidStatus,
  removeAUMID,
  defaultRegistry,
};
