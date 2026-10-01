'use strict';

const os = require('node:os');
const path = require('node:path');

const PLUGIN_FILE_NAME = 'console-notify.js';

function defaultPluginsDir() {
  const home = os.homedir();
  if (!home || home.trim() === '') {
    throw new Error('could not determine the home directory');
  }
  return path.join(home, '.config', 'opencode', 'plugins');
}

function resolvePluginsDir(override) {
  if (typeof override === 'string' && override.trim() !== '') {
    return path.normalize(override);
  }
  return defaultPluginsDir();
}

function pluginPath(pluginsDir) {
  return path.join(pluginsDir, PLUGIN_FILE_NAME);
}

module.exports = {
  PLUGIN_FILE_NAME,
  defaultPluginsDir,
  resolvePluginsDir,
  pluginPath,
};
