// Facade: stable export map (install, uninstall, identity lines). The
// source path resolution lives in source.js, the atomic write in
// atomicWrite.js, the command orchestration in commands.js.

export { install, uninstall, identityInstalledLine, identityRemovedLine } from './commands.js';
