// Domain: config settings persistence. Pure functions plus I/O exercised
// against an isolated temporary mirror of the plugins directory.
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {
  SETTINGS_VERSION,
  defaultSettings,
  normalizeSettings,
  settingsPath,
  loadSettings,
  saveSettings,
} from '../../../lib/config/settings.js';

function tempPluginsDir() {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'ocn-settings-test-'));
}
function removeDir(dir) {
  fs.rmSync(dir, { recursive: true, force: true });
}

test('settings: defaultSettings shape and version', () => {
  const s = defaultSettings();
  assert.strictEqual(s.version, SETTINGS_VERSION);
  assert.strictEqual(s.lang, 'en');
  assert.deepStrictEqual(s.alerts, {
    sessionIdle: true,
    sessionError: true,
    permissionAsked: true,
    questionAsked: true,
  });
  assert.deepStrictEqual(s.customAlerts, []);
});

test('settings: normalizeSettings keeps known langs and coerces alerts', () => {
  const s = normalizeSettings({ version: 2, lang: 'es', alerts: { sessionIdle: false } });
  assert.strictEqual(s.lang, 'es');
  assert.strictEqual(s.alerts.sessionIdle, false);
  assert.strictEqual(s.alerts.sessionError, true);
  assert.strictEqual(s.alerts.permissionAsked, true);
  assert.strictEqual(s.alerts.questionAsked, true);
});

test('settings: normalizeSettings rejects unknown lang and falls back to en', () => {
  assert.strictEqual(normalizeSettings({ lang: 'fr' }).lang, 'en');
  assert.strictEqual(normalizeSettings({}).lang, 'en');
});

test('settings: settingsPath joins plugins dir and file name', () => {
  assert.strictEqual(settingsPath('D:\\plugins'), path.join('D:\\plugins', 'console-notify.config.json'));
});

test('settings: loadSettings missing file returns defaults without error', () => {
  const dir = tempPluginsDir();
  try {
    const s = loadSettings(dir);
    assert.strictEqual(s.lang, 'en');
    assert.strictEqual(s.version, SETTINGS_VERSION);
  } finally {
    removeDir(dir);
  }
});

test('settings: saveSettings then loadSettings round-trips in an isolated mirror dir', () => {
  const dir = tempPluginsDir();
  try {
    const saved = saveSettings(dir, { version: 2, lang: 'es', alerts: { sessionIdle: false } });
    assert.strictEqual(saved.lang, 'es');
    assert.strictEqual(saved.alerts.sessionIdle, false);
    const loaded = loadSettings(dir);
    assert.deepStrictEqual(loaded, saved);
    assert.ok(fs.existsSync(settingsPath(dir)));
  } finally {
    removeDir(dir);
  }
});

test('settings: loadSettings corrupt JSON throws', () => {
  const dir = tempPluginsDir();
  try {
    fs.writeFileSync(settingsPath(dir), '{ not json', 'utf8');
    assert.throws(() => loadSettings(dir), /parse settings file/);
  } finally {
    removeDir(dir);
  }
});

test('settings: saveSettings creates the plugins directory when missing', () => {
  const root = tempPluginsDir();
  const nested = path.join(root, 'deep', 'plugins');
  try {
    saveSettings(nested, { version: 2 });
    assert.ok(fs.existsSync(settingsPath(nested)));
  } finally {
    removeDir(root);
  }
});
