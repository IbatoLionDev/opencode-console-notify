'use strict';

// Mirrors internal/installer/install.go: implements the install and
// uninstall commands against the shipped plugin/ directory.

const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');
const aumid = require('../aumid');
const config = require('../config');

function pluginSourcePath() {
  // The plugin ships inside this package (see package.json "files"),
  // so the source is read from disk instead of a Go embed.
  return path.join(__dirname, '..', '..', 'plugin', config.PLUGIN_FILE_NAME);
}

// writeFileAtomic writes data to dest via a temp file in the same
// directory followed by a rename, so a running OpenCode never observes
// a half-written plugin file. Any leftover temp file is cleaned up.
function writeFileAtomic(dest, data) {
  const dir = path.dirname(dest);
  const tmpName = path.join(dir, `${config.PLUGIN_FILE_NAME}.${crypto.randomUUID()}.tmp`);
  try {
    fs.writeFileSync(tmpName, data, { mode: 0o644 });
    fs.renameSync(tmpName, dest);
  } catch (err) {
    // Best effort cleanup: after a successful rename the temp path no
    // longer exists and unlink is a harmless no-op.
    try {
      fs.unlinkSync(tmpName);
    } catch {
      // Ignore: the original error below is what matters.
    }
    throw err;
  }
}

// Install copies the shipped plugin into pluginsDir (creating it when
// needed) and idempotently registers the AUMID key.
function install(pluginsDir, reg = aumid.defaultRegistry, out = process.stdout) {
  let source;
  try {
    source = fs.readFileSync(pluginSourcePath());
  } catch (err) {
    throw new Error(`read plugin source: ${err.message}`);
  }
  try {
    fs.mkdirSync(pluginsDir, { recursive: true });
  } catch (err) {
    throw new Error(`create plugins directory: ${err.message}`);
  }

  const dest = config.pluginPath(pluginsDir);
  try {
    writeFileAtomic(dest, source);
  } catch (err) {
    throw new Error(`write plugin file: ${err.message}`);
  }

  reg.ensureAUMID();

  const sum = crypto.createHash('sha256').update(source).digest('hex');
  out.write(`Installed: ${dest}\n`);
  out.write(`SHA256:    ${sum}\n`);
  out.write(`Source:    plugin/${config.PLUGIN_FILE_NAME} (no download)\n`);
  out.write(`AUMID:     registered (HKCU:\\Software\\Classes\\AppUserModelId\\${aumid.AUMID}, DisplayName=${aumid.AUMID_DISPLAY_NAME})\n`);
  out.write('Restart OpenCode to load the plugin.\n');
}

// Uninstall removes only the plugin file and exactly the AUMID key,
// never parent keys or anything else in the plugins directory.
function uninstall(pluginsDir, reg = aumid.defaultRegistry, out = process.stdout) {
  const dest = config.pluginPath(pluginsDir);
  let present;
  try {
    present = fs.statSync(dest) !== undefined;
  } catch (err) {
    if (err && err.code === 'ENOENT') {
      present = false;
    } else {
      throw new Error(`stat plugin file: ${err.message}`);
    }
  }

  if (present) {
    try {
      fs.unlinkSync(dest);
    } catch (err) {
      throw new Error(`remove plugin file: ${err.message}`);
    }
    out.write(`Removed plugin file: ${dest}\n`);
  } else {
    out.write(`Plugin file not present, nothing to remove: ${dest}\n`);
  }

  let removed;
  try {
    removed = reg.removeAUMID();
  } catch (err) {
    throw new Error(`remove AUMID key: ${err.message}`);
  }
  if (removed) {
    out.write(`Removed AUMID registry key: HKCU:\\Software\\Classes\\AppUserModelId\\${aumid.AUMID}\n`);
  } else {
    out.write(`AUMID registry key not present, nothing to remove: HKCU:\\Software\\Classes\\AppUserModelId\\${aumid.AUMID}\n`);
  }

  out.write('Uninstall complete. Restart OpenCode to unload the plugin.\n');
}

module.exports = {
  install,
  uninstall,
};
