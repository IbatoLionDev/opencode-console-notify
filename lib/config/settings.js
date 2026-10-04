// Domain: config command settings persistence (schema v2 adds customs).
// Missing file means current behavior: everything on, English.

import fs from 'node:fs';
import path from 'node:path';
import { writeFileAtomic } from '../installer/atomicWrite.js';
import { normalizeCustoms } from './customs.js';

const SETTINGS_FILE_NAME = 'console-notify.config.json';
const SETTINGS_VERSION = 2;

function defaultSettings() {
  return {
    version: SETTINGS_VERSION,
    lang: 'en',
    alerts: {
      sessionIdle: true,
      sessionError: true,
      permissionAsked: true,
      questionAsked: true,
    },
    customAlerts: [],
  };
}

function normalizeSettings(s) {
  const alerts = (s && typeof s === 'object' && s.alerts && typeof s.alerts === 'object')
    ? s.alerts
    : {};
  const lang = (s && s.lang === 'es') || (s && s.lang === 'en') ? s.lang : 'en';
  return {
    version: SETTINGS_VERSION,
    lang,
    alerts: {
      sessionIdle: alerts.sessionIdle !== false,
      sessionError: alerts.sessionError !== false,
      permissionAsked: alerts.permissionAsked !== false,
      questionAsked: alerts.questionAsked !== false,
    },
    customAlerts: normalizeCustoms(s && Array.isArray(s.customAlerts) ? s.customAlerts : []),
  };
}

function settingsPath(pluginsDir) {
  return path.join(pluginsDir, SETTINGS_FILE_NAME);
}

// Missing file returns defaults with no error; corrupt JSON throws
// an English error and leaves the install untouched.
function loadSettings(pluginsDir) {
  const dest = settingsPath(pluginsDir);
  let raw;
  try {
    raw = fs.readFileSync(dest, 'utf8');
  } catch (err) {
    if (err && err.code === 'ENOENT') {
      return defaultSettings();
    }
    throw new Error(`read settings file: ${err.message}`);
  }
  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch (err) {
    throw new Error(`parse settings file: ${err.message}`);
  }
  return normalizeSettings(parsed);
}

function saveSettings(pluginsDir, settings) {
  const normalized = normalizeSettings(settings);
  try {
    fs.mkdirSync(pluginsDir, { recursive: true });
  } catch (err) {
    throw new Error(`create plugins directory: ${err.message}`);
  }
  const dest = settingsPath(pluginsDir);
  try {
    writeFileAtomic(dest, `${JSON.stringify(normalized, null, 2)}\n`);
  } catch (err) {
    throw new Error(`write settings file: ${err.message}`);
  }
  return normalized;
}

export {
  SETTINGS_FILE_NAME,
  SETTINGS_VERSION,
  defaultSettings,
  normalizeSettings,
  settingsPath,
  loadSettings,
  saveSettings,
};
