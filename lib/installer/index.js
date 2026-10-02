// Facade: stable export map (install, uninstall). The source path
// resolution lives in source.js, the atomic write in atomicWrite.js,
// the command orchestration in commands.js.

export { install, uninstall } from './commands.js';
