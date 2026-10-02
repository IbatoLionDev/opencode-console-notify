// Facade: stable export map for the AUMID identity. Pure constants live
// in constants.js, PowerShell execution in powershell.js, registry
// operations in registry.js.

export {
  AUMID,
  AUMID_DISPLAY_NAME,
  AUMID_KEY_PATH,
  AUMID_PS_PATH,
} from './constants.js';
export { encodePowerShell } from './powershell.js';
export {
  ensureAUMID,
  aumidStatus,
  removeAUMID,
  defaultRegistry,
} from './registry.js';
