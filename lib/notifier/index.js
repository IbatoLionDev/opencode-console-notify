// Facade: stable export map (buildToastScript, escapeToastXml,
// NOTIFY_SEND_APP, buildNotifySendArgs, encodePowerShell,
// sendTestToast). Pure builders live in toastScript.js, delivery in
// sender.js.

export { buildToastScript, escapeToastXml, NOTIFY_SEND_APP, buildNotifySendArgs } from './toastScript.js';
export { encodePowerShell, sendTestToast } from './sender.js';
