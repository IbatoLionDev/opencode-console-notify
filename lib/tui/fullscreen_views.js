// Fullscreen menu and info views. Alerts and language live in their own
// files; the loop and TTY gate live in fullscreen.js.
import { saveSettings } from '../config/settings.js';
import { t } from '../i18n/dictionary.js';
import {
  render,
  clampSelection,
  moveSteps,
  frameHeight,
  Key,
} from '../screen/screen.js';
import { NPM_PAGE, REPO_URL, RELEASES_URL } from './tui.js';

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

async function runInfoView(ctx, ask) {
  renderFrame(ctx, t(ctx.lang, 'menu.info'), infoScreenItems(), 0, t(ctx.lang, 'footer.back'));
  await ask();
  return 'menu';
}

export {
  menuScreenItems,
  menuTarget,
  menuShortcut,
  infoScreenItems,
  isBackKey,
  persistSettings,
  renderFrame,
  runMenuView,
  runInfoView,
};
