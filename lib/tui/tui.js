// Domain: config command interactive menu (zero deps, node builtins only).
// Line-based so it works on PowerShell 5.1. Every view prints its keys.
// Pure renders live in tui_renders.js; this file owns the loop.
import readline from 'node:readline';
import { loadSettings, saveSettings } from '../config/settings.js';
import { normalizeLang } from '../i18n/dictionary.js';
import {
  alertItems,
  renderMenu,
  renderLanguage,
  renderAlerts,
  renderInfo,
} from './tui_renders.js';

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
  runTui,
};
export {
  NPM_PAGE,
  REPO_URL,
  RELEASES_URL,
  alertItems,
  renderMenu,
  renderLanguage,
  renderAlerts,
  renderInfo,
} from './tui_renders.js';
