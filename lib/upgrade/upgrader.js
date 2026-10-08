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

// quoteArg quotes one argv token for cmd.exe (doubled inner quotes).
// Tokens here are project constants, never user input; quoting is
// belt-and-braces.
function quoteArg(arg) {
  return `"${String(arg).replace(/"/g, '""')}"`;
}

// runNpm runs npm with an exact argv on POSIX and one pre-quoted command
// line on win32. npm on Windows is a .cmd shim that needs a shell, but
// shell:true with an args array concatenates without escaping (Node
// DEP0190, later a throw), so the string form carries constant-only
// tokens instead.
function runNpm(args, opts) {
  if (process.platform !== 'win32') {
    return spawnSync('npm', args, { ...opts, shell: false });
  }
  return spawnSync(`npm ${args.map(quoteArg).join(' ')}`, { ...opts, shell: true });
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
  const npmInstall = runNpm(['i', '-g', `${PACKAGE_NAME}@latest`], { stdio: 'inherit' });
  if (npmInstall.error || npmInstall.status !== 0) {
    const detail = npmInstall.error ? npmInstall.error.message : `exit code ${npmInstall.status}`;
    throw new Error(`npm install failed (${detail}); current install untouched`);
  }
  // Resolve the fresh CLI from the global node_modules root.
  const rootProbe = runNpm(['root', '-g'], { encoding: 'utf8' });
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
  runNpm,
  upgrade,
};
