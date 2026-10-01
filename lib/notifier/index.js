// Test toast via PowerShell UTF-16LE base64 (-EncodedCommand), quoting-proof.

import { spawnSync } from 'node:child_process';
import * as aumid from '../aumid/index.js';

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

function escapeToastXml(value) {
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}

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
  buildToastScript,
  escapeToastXml,
  encodePowerShell,
  sendTestToast,
};
