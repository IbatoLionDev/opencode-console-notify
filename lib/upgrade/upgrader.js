// Adapter: local version read plus the upgrade orchestration
// (global npm install, then reinstall from the fresh copy).

import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { compareVersions } from './versioning.js';
import { PACKAGE_NAME, fetchLatest } from './registryClient.js';
import { comparePluginBytes } from '../installer/pluginBytes.js';

// Stale-plugin repair reason: printed (without the trailing newline
// being doubled) before the reinstall when the versions are already
// current but the installed plugin bytes differ from this release.
const STALE_PLUGIN_REASON =
  'Already up to date, but the plugin file differs from this release; reinstalling it.\n';

// Pure decision: what upgrade should do, given the registry version, the
// local version, and whether the installed plugin bytes match the
// packaged copy. 'already' keeps the "Already up to date." fast path,
// 'repair' reruns the same reinstall path used after a real upgrade, and
// 'upgrade' runs the normal flow (a newer release supersedes the byte
// check; its post-upgrade reinstall covers the repair either way).
function upgradeDecision(latest, local, pluginBytesEqual) {
  if (compareVersions(latest, local) > 0) {
    return { action: 'upgrade' };
  }
  if (pluginBytesEqual) {
    return { action: 'already' };
  }
  return { action: 'repair', reason: STALE_PLUGIN_REASON };
}

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

// reinstallPlugin installs the plugin from the fresh CLI copy inside the
// global npm store. Shared by the post-upgrade reinstall and the
// stale-repair on the up-to-date fast path. postUpgrade selects the
// wording of the failure guidance: after a real upgrade the lines name
// the new version (unchanged wording), a repair reports its own context.
function reinstallPlugin(pluginsDir, out, postUpgrade, upgradedVersion) {
  const rootProbe = runNpm(['root', '-g'], { encoding: 'utf8' });
  if (rootProbe.error || rootProbe.status !== 0) {
    out.write(postUpgrade
      ? `Upgrade downloaded but the fresh CLI could not be located; run "${PACKAGE_NAME} install" manually.\n`
      : `The fresh CLI could not be located; run "${PACKAGE_NAME} install" manually.\n`);
    return 1;
  }
  const freshCli = path.join(String(rootProbe.stdout).trim(), PACKAGE_NAME, 'bin', 'opencode-console-notify.js');
  const args = [freshCli, 'install'];
  if (pluginsDir) {
    args.push('--plugins-dir', pluginsDir);
  }
  const reinstall = spawnSync(process.execPath, args, { stdio: 'inherit' });
  if (reinstall.error || reinstall.status !== 0) {
    out.write(postUpgrade
      ? `Upgraded to ${upgradedVersion} but the reinstall did not finish; run "${PACKAGE_NAME} install" manually, then check with "${PACKAGE_NAME} doctor".\n`
      : `The version is current but the reinstall did not finish; run "${PACKAGE_NAME} install" manually, then check with "${PACKAGE_NAME} doctor".\n`);
    return 1;
  }
  return 0;
}

// Upgrade to the latest registry release, then reinstall the plugin from
// the fresh copy. Returns the process exit code; throws when the global
// npm install fails (the current install is left untouched in that case).
async function upgrade(pluginsDir, out = process.stdout) {
  const local = readLocalVersion();
  const latest = await fetchLatest();
  out.write(`Local version:  ${local}\n`);
  out.write(`Latest version: ${latest}\n`);
  // Fast path stays only when the installed plugin bytes also match this
  // release; differing bytes self-heal through the reinstall path.
  const decision = upgradeDecision(latest, local, comparePluginBytes(pluginsDir).equal);
  if (decision.action === 'already') {
    out.write('Already up to date.\n');
    return 0;
  }
  if (decision.action === 'repair') {
    out.write(decision.reason);
    return reinstallPlugin(pluginsDir, out, false, '');
  }
  out.write(`Upgrading ${PACKAGE_NAME} to ${latest}...\n`);
  const npmInstall = runNpm(['i', '-g', `${PACKAGE_NAME}@latest`], { stdio: 'inherit' });
  if (npmInstall.error || npmInstall.status !== 0) {
    const detail = npmInstall.error ? npmInstall.error.message : `exit code ${npmInstall.status}`;
    throw new Error(`npm install failed (${detail}); current install untouched`);
  }
  return reinstallPlugin(pluginsDir, out, true, latest);
}

export {
  readLocalVersion,
  runNpm,
  upgradeDecision,
  STALE_PLUGIN_REASON,
  upgrade,
};
