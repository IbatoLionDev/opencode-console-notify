'use strict';

// Mirrors internal/notifier/toast.go: sends a real Windows toast for the
// test command. It shells out to PowerShell with a UTF-16LE base64 payload,
// the same dependency-free technique the plugin uses, so quoting can never
// break.

const { spawnSync } = require('node:child_process');
const aumid = require('../aumid');

// buildToastScript renders the PowerShell snippet that shows one toast
// through the shared AUMID identity.
function buildToastScript(title, line) {
  const xml =
    '<toast duration="long"><visual><binding template="ToastGeneric">' +
    `<text>${escapeToastXml(title)}</text>` +
    `<text>${escapeToastXml(line)}</text>` +
    '</binding></visual><audio src="ms-winsoundevent:Notification.Default"/></toast>';
  const scriptLines = [
    '[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null',
    '[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null',
    '$x = New-Object Windows.Data.Xml.Dom.XmlDocument',
    `$x.LoadXml('${xml}')`,
    '$t = [Windows.UI.Notifications.ToastNotification]::new($x)',
    "$t.Tag = 'opencode-test'",
    "$t.Group = 'opencode'",
    `[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('${aumid.AUMID}').Show($t)`,
  ];
  return scriptLines.join('\n');
}

// escapeToastXml escapes the five XML metacharacters for toast payloads.
function escapeToastXml(value) {
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}

// encodePowerShell encodes a script as UTF-16LE base64 for
// powershell.exe -EncodedCommand, avoiding every shell-quoting pitfall.
function encodePowerShell(script) {
  return Buffer.from(script, 'utf16le').toString('base64');
}

// SendTestToast ensures the AUMID registration exists (idempotent) and
// then shows a real toast so a manual run visibly notifies.
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
    throw new Error(`show test toast: powershell exited ${result.status}${output ? `: ${output}` : ''}`);
  }
}

module.exports = {
  buildToastScript,
  escapeToastXml,
  encodePowerShell,
  sendTestToast,
};
