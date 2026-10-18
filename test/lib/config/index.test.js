// Domain: plugins directory resolution. Pure (os.homedir is read-only here).
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import os from 'node:os';
import path from 'node:path';
import {
  PLUGIN_FILE_NAME,
  defaultPluginsDir,
  resolvePluginsDir,
  pluginPath,
} from '../../../lib/config/index.js';

test('config index: PLUGIN_FILE_NAME', () => {
  assert.strictEqual(PLUGIN_FILE_NAME, 'console-notify.js');
});

test('config index: resolvePluginsDir normaliza el override', () => {
  assert.strictEqual(resolvePluginsDir('D:\\plugins'), 'D:\\plugins');
  assert.strictEqual(resolvePluginsDir('/a/b'), path.normalize('/a/b'));
});

test('config index: resolvePluginsDir por defecto vive bajo el home directory', () => {
  const d = defaultPluginsDir();
  assert.ok(typeof d === 'string' && d.length > 0);
  assert.ok(d.includes(path.join(os.homedir(), '.config', 'opencode', 'plugins')));
});

test('config index: pluginPath une el plugins dir y el nombre de archivo', () => {
  assert.strictEqual(pluginPath('D:\\plugins'), 'D:\\plugins\\console-notify.js');
  assert.ok(pluginPath('/a/b').endsWith('console-notify.js'));
});
