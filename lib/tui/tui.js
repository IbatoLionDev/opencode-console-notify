// Domain: config command interactive menu (zero deps, node builtins only).
// Line-based so it works on PowerShell 5.1. Every view prints its keys.

import readline from 'node:readline';
import { loadSettings, saveSettings } from '../config/settings.js';
import { normalizeLang, t } from '../i18n/dictionary.js';

const NPM_PAGE = 'https://www.npmjs.com/package/opencode-console-notify';
const REPO_URL = 'https://github.com/IbatoLionDev/opencode-console-notify';
const RELEASES_URL = 'https://github.com/IbatoLionDev/opencode-console-notify/releases';

function alertItems(settings, lang) {
  return [
    { key: '1', event: 'session.idle', name: t(lang, 'alert.sessionIdle'), trigger: t(lang, 'trigger.sessionIdle'), enabled: settings.alerts.sessionIdle !== false, field: 'sessionIdle' },
    { key: '2', event: 'session.error', name: t(lang, 'alert.sessionError'), trigger: t(lang, 'trigger.sessionError'), enabled: settings.alerts.sessionError !== false, field: 'sessionError' },
    { key: '3', event: 'permission.asked', name: t(lang, 'alert.permissionAsked'), trigger: t(lang, 'trigger.permissionAsked'), enabled: settings.alerts.permissionAsked !== false, field: 'permissionAsked' },
    { key: '4', event: 'question.asked', name: t(lang, 'alert.questionAsked'), trigger: t(lang, 'trigger.questionAsked'), enabled: settings.alerts.questionAsked !== false, field: 'questionAsked' },
  ];
}

function onOff(lang, on) {
  return t(lang, on ? 'state.on' : 'state.off');
}

function renderMenu(lang) {
  return [
    t(lang, 'app.title'),
    `1. ${t(lang, 'menu.language')} (L)`,
    `2. ${t(lang, 'menu.update')} (U)`,
    `3. ${t(lang, 'menu.info')} (I)`,
    `4. ${t(lang, 'menu.alerts')} (A)`,
    `5. ${t(lang, 'menu.exit')} (Q)`,
    t(lang, 'footer.menu'),
    '',
  ].join('\n');
}

function renderLanguage(lang) {
  const markEn = lang === 'en' ? ` [${t(lang, 'lang.current')}]` : '';
  const markEs = lang === 'es' ? ` [${t(lang, 'lang.current')}]` : '';
  return [
    t(lang, 'lang.title'),
    `1. English${markEn}`,
    `2. Español${markEs}`,
    `[1/2] ${t(lang, 'lang.current')}: ${lang} | ${t(lang, 'footer.back')}`,
    '',
  ].join('\n');
}

function renderAlerts(settings, lang) {
  const lines = [t(lang, 'alerts.title')];
  for (const a of alertItems(settings, lang)) {
    lines.push(`${a.key}. ${a.name} [${onOff(lang, a.enabled)}]`);
    lines.push(`   ${a.trigger}`);
  }
  lines.push(t(lang, 'footer.toggle'));
  lines.push('');
  return lines.join('\n');
}

function renderInfo(lang) {
  return [
    t(lang, 'menu.info'),
    'install, uninstall, doctor, test, upgrade, version, config',
    `npm: ${NPM_PAGE}`,
    `repo: ${REPO_URL}`,
    `releases: ${RELEASES_URL}`,
    t(lang, 'footer.back'),
    '',
  ].join('\n');
}

function isBackKey(key) {
  return key === 'q' || key === 'esc' || key === 'b' || key === 'back' || key === 'volver';
}

function menuTransition(key) {
  if (key === '1' || key === 'l' || key === 'language' || key === 'idioma') return 'language';
  if (key === '2' || key === 'u' || key === 'update' || key === 'actualizar') return 'update';
  if (key === '3' || key === 'i' || key === 'info') return 'info';
  if (key === '4' || key === 'a' || key === 'alerts' || key === 'alertas') return 'alerts';
  if (key === '5' || key === 'q' || key === 'quit' || key === 'exit' || key === 'salir' || key === 'esc') return 'exit';
  return 'menu';
}

async function runMenuView(ctx, ask) {
  ctx.output.write(renderMenu(ctx.lang));
  return menuTransition(await ask());
}

function persistSettings(ctx) {
  try {
    saveSettings(ctx.pluginsDir, ctx.settings);
  } catch (err) {
    ctx.output.write(`Error: ${err.message}\n`);
    return 1;
  }
  return 0;
}

async function runLanguageView(ctx, ask) {
  ctx.output.write(renderLanguage(ctx.lang));
  const key = await ask();
  if (key === '1' || key === 'en' || key === '2' || key === 'es') {
    ctx.lang = key === '2' || key === 'es' ? 'es' : 'en';
    ctx.settings.lang = ctx.lang;
    if (persistSettings(ctx) !== 0) {
      return 'abort';
    }
    return 'language';
  }
  if (isBackKey(key)) {
    return 'menu';
  }
  return 'language';
}

async function runAlertsView(ctx, ask) {
  ctx.output.write(renderAlerts(ctx.settings, ctx.lang));
  const key = await ask();
  if (isBackKey(key)) {
    return 'menu';
  }
  // Space-prefixed numbers arrive trimmed, so the digit alone toggles.
  const item = alertItems(ctx.settings, ctx.lang).find((a) => a.key === key);
  if (item === undefined) {
    return 'alerts';
  }
  ctx.settings.alerts[item.field] = !item.enabled;
  try {
    saveSettings(ctx.pluginsDir, ctx.settings);
    ctx.settings = loadSettings(ctx.pluginsDir);
  } catch (err) {
    ctx.output.write(`Error: ${err.message}\n`);
    return 'abort';
  }
  return 'alerts';
}

async function runInfoView(ctx, ask) {
  ctx.output.write(renderInfo(ctx.lang));
  await ask();
  return 'menu';
}

async function runTui(pluginsDir, input, output, upgrade) {
  let settings;
  try {
    settings = loadSettings(pluginsDir);
  } catch (err) {
    output.write(`Error: ${err.message}\n`);
    return 1;
  }
  const ctx = { pluginsDir, settings, lang: normalizeLang(settings.lang), output };
  const rl = readline.createInterface({ input, output, terminal: true });
  const ask = () => new Promise((resolve) => {
    rl.question('', (answer) => resolve(String(answer ?? '').trim().toLowerCase()));
  });
  let view = 'menu';
  for (;;) {
    if (view === 'menu') {
      view = await runMenuView(ctx, ask);
    } else if (view === 'language') {
      view = await runLanguageView(ctx, ask);
    } else if (view === 'alerts') {
      view = await runAlertsView(ctx, ask);
    } else if (view === 'info') {
      view = await runInfoView(ctx, ask);
    } else if (view === 'update') {
      rl.close();
      if (typeof upgrade === 'function') {
        return await upgrade(output);
      }
      return 0;
    } else {
      rl.close();
      // 'exit' and 'abort' both leave the loop; abort already reported.
      return view === 'exit' ? 0 : 1;
    }
  }
}

export {
  NPM_PAGE,
  REPO_URL,
  RELEASES_URL,
  alertItems,
  renderMenu,
  renderLanguage,
  renderAlerts,
  renderInfo,
  runTui,
};
