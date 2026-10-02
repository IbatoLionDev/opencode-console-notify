// Facade: stable export map for upgrades. Pure version comparison lives
// in versioning.js, the registry client in registryClient.js, the local
// read plus orchestration in upgrader.js.

export { compareVersions } from './versioning.js';
export { fetchLatest } from './registryClient.js';
export { readLocalVersion, upgrade } from './upgrader.js';
