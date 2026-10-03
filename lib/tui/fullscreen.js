// Adapter: fullscreen config loop (keyboard navigation, in-place redraw).
// Line mode (tui.js runTui) stays as the fallback for pipes and scripts.

import readline from 'node:readline';
import { loadSettings, saveSettings } from '../config/settings.js';
import { normalizeLang, t } from '../i18n/dictionary.js';
import {
  render,
  clampSelection,
  moveSteps,
  frameHeight,
  enterFrame,
  leaveFrame,
  parseKeyPress,
  Key,
} from '../screen/screen.js';
import { alertItems, runTui, NPM_PAGE, REPO_URL, RELEASES_URL } from './tui.js';

function menuScreenItems(lang) {
  return [
    { label: t(lang, 'menu.language') },
    { label: t(lang, 'menu.update') },
    { label: t(lang, 'menu.info') },
    { label: t(lang, 'menu.alerts') },
    { label: t(lang, 'menu.exit') },
  ];
}

function menuTarget(sel) {
  switch (sel) {
    case 0: return 'language';
    case 1: return 'update';
    case 2: return 'info';
    case 3: return 'alerts';
    default: return 'exit';
  }
}

function menuShortcut(rune) {
  switch (rune) {
    case '1': case 'l': case 'L': return 'language';
    case '2': case 'u': case 'U': return 'update';
    case '3': case 'i': case 'I': return 'info';
    case '4': case 'a': case 'A': return 'alerts';
    case '5': return 'exit';
    default: return '';
  }
}

function digitIndex(rune, n) {
  const idx = rune.charCodeAt(0) - '1'.charCodeAt(0);
  if (rune.length !== 1 || idx < 0 || idx >= n) {
    return -1;
  }
  return idx;
}

function stateKey(on) {
  return on ? 'state.on' : 'state.off';
}

function alertField(idx) {
  switch (idx) {
    case 0: return 'sessionIdle';
    case 1: return 'sessionError';
    case 2: return 'permissionAsked';
    default: return 'questionAsked';
  }
}

function alertScreenItems(settings, lang) {
  return alertItems(settings, lang).map((a) => ({
    label: a.name,
    detail: a.trigger,
    state: t(lang, stateKey(a.enabled)),
  }));
}

function langIndex(lang) {
  return lang === 'es' ? 1 : 0;
}

function currentMark(active, lang) {
  return active ? t(lang, 'lang.current') : '';
}

function infoScreenItems() {
  return [
    { label: 'install, uninstall, doctor, test, upgrade, version, config' },
    { label: `npm: ${NPM_PAGE}` },
    { label: `repo: ${REPO_URL}` },
    { label: `releases: ${RELEASES_URL}` },
  ];
}

function isBackKey(key) {
  return key === 'q' || key === 'esc' || key === 'b' || key === 'back' || key === 'volver';
}

function persistSettings(ctx) {
  try {
    saveSettings(ctx.pluginsDir, ctx.settings);
  } catch (err) {
    ctx.output.write(`Error: ${err.message}\n`);
    return false;
  }
  return true;
}

function renderFrame(ctx, title, items, selected, footer) {
  ctx.output.write(render({
    title,
    items,
    selected,
    footer,
    height: frameHeight(ctx.rows),
  }, ctx.cols));
}

async function runMenuView(ctx, ask, sel) {
  const items = menuScreenItems(ctx.lang);
  const clamped = clampSelection(items.length, sel.value);
  sel.value = clamped;
  renderFrame(ctx, t(ctx.lang, 'app.title'), items, clamped, t(ctx.lang, 'footer.menu'));
  const pressed = await ask();
  if (pressed.key === Key.RUNE) {
    return menuShortcut(pressed.rune) || 'menu';
  }
  if (pressed.key === Key.UP) {
    sel.value = moveSteps(items.length, clamped, -1);
  } else if (pressed.key === Key.DOWN) {
    sel.value = moveSteps(items.length, clamped, 1);
  } else if (pressed.key === Key.ENTER) {
    return menuTarget(clamped);
  } else if (pressed.key === Key.QUIT || pressed.key === Key.ESC) {
    return 'exit';
  }
  return 'menu';
}

async function runAlertsView(ctx, ask, sel) {
  const items = alertScreenItems(ctx.settings, ctx.lang);
  const clamped = clampSelection(items.length, sel.value);
  sel.value = clamped;
  renderFrame(ctx, t(ctx.lang, 'alerts.title'), items, clamped, t(ctx.lang, 'footer.toggle'));
  const pressed = await ask();
  if (pressed.key === Key.RUNE) {
    const idx = digitIndex(pressed.rune, items.length);
    if (idx >= 0 && toggleAlert(ctx, idx)) {
      return 'abort';
    }
    return 'alerts';
  }
  if (pressed.key === Key.UP) {
    sel.value = moveSteps(items.length, clamped, -1);
  } else if (pressed.key === Key.DOWN) {
    sel.value = moveSteps(items.length, clamped, 1);
  } else if (pressed.key === Key.ENTER || pressed.key === Key.SPACE) {
    if (!toggleAlert(ctx, clamped)) {
      return 'abort';
    }
  } else if (pressed.key === Key.QUIT || pressed.key === Key.ESC) {
    return 'menu';
  }
  return 'alerts';
}

function toggleAlert(ctx, idx) {
  ctx.settings.alerts[alertField(idx)] = !ctx.settings.alerts[alertField(idx)];
  return persistSettings(ctx);
}

async function runLanguageView(ctx, ask) {
  const items = [
    { label: 'English', state: currentMark(ctx.lang === 'en', ctx.lang) },
    { label: 'Español', state: currentMark(ctx.lang === 'es', ctx.lang) },
  ];
  renderFrame(ctx, t(ctx.lang, 'lang.title'), items, langIndex(ctx.lang), t(ctx.lang, 'footer.back'));
  const pressed = await ask();
  if (pressed.key === Key.RUNE) {
    if (pressed.rune === '1' || pressed.rune === 'e' || pressed.rune === 'E') {
      return setLanguage(ctx, 'en');
    }
    if (pressed.rune === '2') {
      return setLanguage(ctx, 'es');
    }
    return 'language';
  }
  if (pressed.key === Key.ENTER || pressed.key === Key.SPACE) {
    return setLanguage(ctx, ctx.lang === 'es' ? 'en' : 'es');
  }
  if (pressed.key === Key.QUIT || pressed.key === Key.ESC) {
    return 'menu';
  }
  return 'language';
}

function setLanguage(ctx, lang) {
  ctx.lang = lang;
  ctx.settings.lang = lang;
  if (!persistSettings(ctx)) {
    return 'abort';
  }
  return 'language';
}

async function runInfoView(ctx, ask) {
  renderFrame(ctx, t(ctx.lang, 'menu.info'), infoScreenItems(), 0, t(ctx.lang, 'footer.back'));
  await ask();
  return 'menu';
}

async function runFullscreen(pluginsDir, input, output, upgrade) {
  let settings;
  try {
    settings = loadSettings(pluginsDir);
  } catch (err) {
    output.write(`Error: ${err.message}\n`);
    return 1;
  }
  const ctx = {
    pluginsDir,
    settings,
    lang: normalizeLang(settings.lang),
    output,
    cols: output.columns || 80,
    rows: output.rows || 24,
  };
  input.setRawMode(true);
  readline.emitKeypressEvents(input);
  output.write(enterFrame());
  const ask = () => new Promise((resolve) => {
    input.once('keypress', (ch, key) => resolve(parseKeyPress(key?.name || '', key?.sequence || ch || '')));
  });
  const menuSel = { value: 0 };
  const alertsSel = { value: 0 };
  let view = 'menu';
  try {
    for (;;) {
      if (view === 'menu') {
        view = await runMenuView(ctx, ask, menuSel);
      } else if (view === 'alerts') {
        view = await runAlertsView(ctx, ask, alertsSel);
      } else if (view === 'language') {
        view = await runLanguageView(ctx, ask);
      } else if (view === 'info') {
        view = await runInfoView(ctx, ask);
      } else if (view === 'update') {
        output.write(leaveFrame());
        input.setRawMode(false);
        if (typeof upgrade === 'function') {
          return await upgrade(output);
        }
        return 0;
      } else if (view === 'exit') {
        return 0;
      } else {
        return 1;
      }
    }
  } finally {
    try {
      input.setRawMode(false);
    } catch {
    }
    output.write(leaveFrame());
  }
}

// runSmartTui opens the fullscreen loop on a real console and the line
// mode everywhere else (pipes, scripts, probes).
async function runSmartTui(pluginsDir, input, output, upgrade) {
  if (input?.isTTY && output?.isTTY) {
    return await runFullscreen(pluginsDir, input, output, upgrade);
  }
  return await runTui(pluginsDir, input, output, upgrade);
}

export {
  menuTarget,
  menuShortcut,
  digitIndex,
  alertField,
  runFullscreen,
  runSmartTui,
};
