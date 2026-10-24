// Adapter: byte comparison between the installed plugin file and the
// packaged copy shipped inside this package. Shared by doctor (health
// content check) and upgrade (fast-path self-heal) so the comparison is
// defined exactly once. The hash follows the same createHash('sha256')
// pattern as the installer's install output.

import crypto from 'node:crypto';
import fs from 'node:fs';
import * as config from '../config/index.js';
import { pluginSourcePath } from './source.js';

// sha256OrNull hashes one file's contents, returning null when the file
// does not exist (nothing to compare against) and propagating any other
// read failure; the label names the file group for the error message.
function sha256OrNull(file, label) {
  let data;
  try {
    data = fs.readFileSync(file);
  } catch (err) {
    if (err && err.code === 'ENOENT') {
      return null;
    }
    throw new Error(`read ${label}: ${err.message}`);
  }
  return crypto.createHash('sha256').update(data).digest('hex');
}

// comparePluginBytes hashes the installed plugin file at
// config.pluginPath(pluginsDir) and the packaged copy resolved by
// pluginSourcePath(), and reports whether they match byte for byte.
// Missing files on either side are never equal; an unreadable (non
// ENOENT) file throws so callers surface the state instead of guessing.
function comparePluginBytes(pluginsDir) {
  const installedPath = config.pluginPath(pluginsDir);
  const sourcePath = pluginSourcePath();
  const installedSha = sha256OrNull(installedPath, 'installed plugin file');
  const sourceSha = sha256OrNull(sourcePath, 'plugin source');
  return {
    equal: installedSha !== null && sourceSha !== null && installedSha === sourceSha,
    installedPath,
    sourcePath,
    installedPresent: installedSha !== null,
    sourcePresent: sourceSha !== null,
  };
}

export { comparePluginBytes };
