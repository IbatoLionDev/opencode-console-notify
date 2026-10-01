'use strict';

// Mirrors internal/config/paths.go: resolves where the plugin file lives.
// Semantics are identical to the Go side: the --plugins-dir override wins
// when non-blank (after trimming), otherwise the default is used.

const os = require('node:os');
const path = require('node:path');

// PLUGIN_FILE_NAME is the only file this CLI ever writes or removes.
const PLUGIN_FILE_NAME = 'console-notify.js';

// DefaultPluginsDir returns OpenCode's Windows plugin directory:
// <HOME>/.config/opencode/plugins. It mirrors install.ps1's default.
function defaultPluginsDir() {
  const home = os.homedir();
  if (!home || home.trim() === '') {
    throw new Error('could not determine the home directory');
  }
  return path.join(home, '.config', 'opencode', 'plugins');
}

// ResolvePluginsDir returns the --plugins-dir override when non-blank,
// otherwise the default plugins directory.
function resolvePluginsDir(override) {
  if (typeof override === 'string' && override.trim() !== '') {
    // path.normalize is the Node equivalent of Go's filepath.Clean.
    return path.normalize(override);
  }
  return defaultPluginsDir();
}

// PluginPath joins the plugins directory with the plugin file name.
function pluginPath(pluginsDir) {
  return path.join(pluginsDir, PLUGIN_FILE_NAME);
}

module.exports = {
  PLUGIN_FILE_NAME,
  defaultPluginsDir,
  resolvePluginsDir,
  pluginPath,
};
