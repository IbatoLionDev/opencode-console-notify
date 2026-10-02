// Facade: stable export map (checkHealth, healthy, doctor). Health rules
// and orchestration live in health.js, the update notice in
// updateNotice.js (imported by health.js, not re-exported here).

export { checkHealth, healthy, doctor } from './health.js';
