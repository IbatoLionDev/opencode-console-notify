// Facade: stable export map (buildToastScript, escapeToastXml,
// encodePowerShell, sendTestToast). Pure builders live in toastScript.js,
// delivery in sender.js.

export { buildToastScript, escapeToastXml } from './toastScript.js';
export { encodePowerShell, sendTestToast } from './sender.js';
