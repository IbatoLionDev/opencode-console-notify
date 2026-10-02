// Adapter: config command (flags for scripts + interactive TUI).
// Update reuses the existing upgrade flow for this distribution.

import { loadSettings, saveSettings } from '../config/settings.js';
import { normalizeLang } from '../i18n/dictionary.js';
import { renderAlerts, renderInfo, runTui } from '../tui/tui.js';

function printAlerts(pluginsDir, out) {
  let settings;
  try {
    settings = loadSettings(pluginsDir);
  } catch (err) {
    out.write(`Error: ${err.message}\n`);
    return 1;
  }
  out.write(renderAlerts(settings, normalizeLang(settings.lang)));
  return 0;
}

function applyLang(pluginsDir, lang, out) {
  const want = String(lang ?? '').trim().toLowerCase();
  if (want !== 'en' && want !== 'es') {
    out.write(`Error: --lang must be en or es (got "${lang}")\n`);
    return 2;
  }
  let settings;
  try {
    settings = loadSettings(pluginsDir);
  } catch (err) {
    out.write(`Error: ${err.message}\n`);
    return 1;
  }
  settings.lang = want;
  try {
    saveSettings(pluginsDir, settings);
  } catch (err) {
    out.write(`Error: ${err.message}\n`);
    return 1;
  }
  out.write(`Language: ${want}\n`);
  return 0;
}

function applyToggle(pluginsDir, spec, out) {
  const eq = String(spec ?? '').indexOf('=');
  if (eq < 0) {
    out.write(`Error: --toggle needs KEY=on|off (got "${spec}")\n`);
    return 2;
  }
  const key = String(spec.slice(0, eq)).trim().toLowerCase();
  const value = String(spec.slice(eq + 1)).trim().toLowerCase();
  const on = value === 'on' || value === 'true' || value === '1';
  const off = value === 'off' || value === 'false' || value === '0';
  if (!on && !off) {
    out.write(`Error: --toggle value must be on or off (got "${value}")\n`);
    return 2;
  }
  let settings;
  try {
    settings = loadSettings(pluginsDir);
  } catch (err) {
    out.write(`Error: ${err.message}\n`);
    return 1;
  }
  switch (key) {
    case 'sessionidle':
    case 'session.idle':
    case 'idle':
      settings.alerts.sessionIdle = on;
      break;
    case 'sessionerror':
    case 'session.error':
    case 'error':
      settings.alerts.sessionError = on;
      break;
    case 'permissionasked':
    case 'permission.asked':
    case 'permission':
    case 'perm':
      settings.alerts.permissionAsked = on;
      break;
    case 'questionasked':
    case 'question.asked':
    case 'question':
      settings.alerts.questionAsked = on;
      break;
    default:
      out.write(`Error: unknown alert "${key}" (want sessionIdle|sessionError|permissionAsked|questionAsked)\n`);
      return 2;
  }
  try {
    saveSettings(pluginsDir, settings);
  } catch (err) {
    out.write(`Error: ${err.message}\n`);
    return 1;
  }
  return printAlerts(pluginsDir, out);
}

async function runConfig(pluginsDir, configArgs, upgrade, stdin = process.stdin, out = process.stdout) {
  const rest = Array.isArray(configArgs) ? configArgs : [];
  for (let i = 0; i < rest.length; i++) {
    const arg = rest[i];
    if (arg === '--list-alerts' || arg === '-list-alerts') {
      return printAlerts(pluginsDir, out);
    }
    if (arg === '--info' || arg === '-info') {
      let settings;
      try {
        settings = loadSettings(pluginsDir);
      } catch (err) {
        out.write(`Error: ${err.message}\n`);
        return 1;
      }
      out.write(renderInfo(normalizeLang(settings.lang)));
      return 0;
    }
    if (arg === '--lang' || arg === '-lang') {
      if (i + 1 >= rest.length) {
        out.write(`Error: flag ${arg} needs a value\n`);
        return 2;
      }
      return applyLang(pluginsDir, rest[i + 1], out);
    }
    if (arg.startsWith('--lang=')) {
      return applyLang(pluginsDir, arg.slice('--lang='.length), out);
    }
    if (arg === '--toggle' || arg === '-toggle') {
      if (i + 1 >= rest.length) {
        out.write(`Error: flag ${arg} needs a value\n`);
        return 2;
      }
      return applyToggle(pluginsDir, rest[i + 1], out);
    }
    if (arg.startsWith('--toggle=')) {
      return applyToggle(pluginsDir, arg.slice('--toggle='.length), out);
    }
    out.write(`Error: unknown config flag "${arg}" (want --lang|--toggle|--list-alerts|--info)\n`);
    return 2;
  }
  return await runTui(pluginsDir, stdin, out, upgrade);
}

export { runConfig };
