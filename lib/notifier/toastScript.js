// Pure domain: toast payloads and notifier argument builders.
// No I/O; sending lives in sender.js.

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

const NOTIFY_SEND_APP = 'OpenCode';

// buildNotifySendArgs renders the notify-send argv for one summary plus
// one body line (exact vector, never through a shell).
function buildNotifySendArgs(title, body) {
  return ['--app-name=' + NOTIFY_SEND_APP, '--urgency=normal', '--expire-time=8000', String(title), String(body)];
}

export {
  buildToastScript,
  escapeToastXml,
  NOTIFY_SEND_APP,
  buildNotifySendArgs,
};
