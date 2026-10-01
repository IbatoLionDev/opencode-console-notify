// Upgrade: compare the local package version against the npm registry,
// then reinstall globally and re-run install from the fresh copy.

import { spawnSync } from 'node:child_process';
import { get } from 'node:https';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const PACKAGE_NAME = 'opencode-console-notify';
const REGISTRY_URL = 'https://registry.npmjs.org/opencode-console-notify/latest';
const FETCH_TIMEOUT_MS = 10000;

// Fetch the latest published version from the npm registry.
function fetchLatest(timeoutMs = FETCH_TIMEOUT_MS) {
  return new Promise((resolve, reject) => {
    const req = get(REGISTRY_URL, (res) => {
      const status = res.statusCode || 0;
      if (status < 200 || status >= 300) {
        res.resume();
        reject(new Error(`registry request failed with status ${status}`));
        return;
      }
      let body = '';
      res.setEncoding('utf8');
      res.on('data', (chunk) => {
        body += chunk;
      });
      res.on('end', () => {
        let data;
        try {
          data = JSON.parse(body);
        } catch (err) {
          reject(new Error(`parse registry response: ${err.message}`));
          return;
        }
        if (!data || typeof data.version !== 'string' || data.version === '') {
          reject(new Error('registry response has no version field'));
          return;
        }
        resolve(data.version);
      });
    });
    req.on('error', (err) => {
      reject(new Error(`registry request failed: ${err.message}`));
    });
    req.setTimeout(timeoutMs, () => {
      req.destroy(new Error(`registry request timed out after ${timeoutMs}ms`));
    });
  });
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

// Numeric semver comparison: negative when a < b, 0 when equal,
// positive when a > b.
function compareVersions(a, b) {
  const core = (v) => v
    .split('+')[0]
    .split('-')[0]
    .split('.')
    .map((part) => {
      const n = Number.parseInt(part, 10);
      return Number.isNaN(n) ? 0 : n;
    });
  const pa = core(a);
  const pb = core(b);
  const len = Math.max(pa.length, pb.length);
  for (let i = 0; i < len; i++) {
    const x = pa[i] || 0;
    const y = pb[i] || 0;
    if (x !== y) {
      return x - y;
    }
  }
  // A pre-release (e.g. 1.2.0-beta) sorts below its plain release.
  const preA = a.includes('-');
  const preB = b.includes('-');
  if (preA && !preB) {
    return -1;
  }
  if (!preA && preB) {
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
  fetchLatest,
  readLocalVersion,
  compareVersions,
  upgrade,
};
