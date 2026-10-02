// Adapter: local version read plus the upgrade orchestration
// (global npm install, then reinstall from the fresh copy).

import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { compareVersions } from './versioning.js';
import { PACKAGE_NAME, fetchLatest } from './registryClient.js';

// Read the version of this checkout from package.json above lib/.
function readLocalVersion() {
  const here = path.dirname(fileURLToPath(import.meta.url));
  const pkgPath = path.join(here, '..', '..', 'package.json');
  let raw;
  try {
    raw = fs.readFileSync(pkgPath, 'utf8');
  } catch (err) {
    throw new Error(`read local package.json: ${err.message}`);
  }
  let pkg;
  try {
    pkg = JSON.parse(raw);
  } catch (err) {
    throw new Error(`parse local package.json: ${err.message}`);
  }
  if (!pkg || typeof pkg.version !== 'string' || pkg.version === '') {
    throw new Error('local package.json has no version field');
  }
  return pkg.version;
}

// Upgrade to the latest registry release, then reinstall the plugin from
// the fresh copy. Returns the process exit code; throws when the global
// npm install fails (the current install is left untouched in that case).
async function upgrade(pluginsDir, out = process.stdout) {
  const local = readLocalVersion();
  const latest = await fetchLatest();
  out.write(`Local version:  ${local}\n`);
  out.write(`Latest version: ${latest}\n`);
  if (compareVersions(latest, local) <= 0) {
    out.write('Already up to date.\n');
    return 0;
  }
  out.write(`Upgrading ${PACKAGE_NAME} to ${latest}...\n`);
  // shell:true on win32 so the npm shim resolves.
  const shell = process.platform === 'win32';
  const npmInstall = spawnSync('npm', ['i', '-g', `${PACKAGE_NAME}@latest`], {
    stdio: 'inherit',
    shell,
  });
  if (npmInstall.error || npmInstall.status !== 0) {
    const detail = npmInstall.error ? npmInstall.error.message : `exit code ${npmInstall.status}`;
    throw new Error(`npm install failed (${detail}); current install untouched`);
  }
  // Resolve the fresh CLI from the global node_modules root.
  const rootProbe = spawnSync('npm', ['root', '-g'], { encoding: 'utf8', shell });
  if (rootProbe.error || rootProbe.status !== 0) {
    out.write(`Upgrade downloaded but the fresh CLI could not be located; run "${PACKAGE_NAME} install" manually.\n`);
    return 1;
  }
  const freshCli = path.join(String(rootProbe.stdout).trim(), PACKAGE_NAME, 'bin', 'opencode-console-notify.js');
  const args = [freshCli, 'install'];
  if (pluginsDir) {
    args.push('--plugins-dir', pluginsDir);
  }
  const reinstall = spawnSync(process.execPath, args, { stdio: 'inherit' });
  if (reinstall.error || reinstall.status !== 0) {
    out.write(`Upgraded to ${latest} but the reinstall did not finish; run "${PACKAGE_NAME} install" manually, then check with "${PACKAGE_NAME} doctor".\n`);
    return 1;
  }
  return 0;
}

export {
  readLocalVersion,
  upgrade,
};
