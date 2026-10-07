// Fullscreen alerts view: the four default toggles rendered as a frame.
import { t } from '../i18n/dictionary.js';
import {
  clampSelection,
  moveSteps,
  frameHeight,
  hitRowToIndex,
  Key,
} from '../screen/screen.js';
import { allAlertItems } from './tui.js';
import { toggleCustom } from '../config/customs.js';
import { renderFrame, persistSettings } from './fullscreen_views.js';

function stateKey(on) {
  return on ? 'state.on' : 'state.off';
}

// alertScreenItems maps all alerts (defaults plus customs) over screen
// rows; the shared mapping lives in tui_renders.js.
function alertScreenItems(settings, lang) {
  return allAlertItems(settings, lang).map((a) => ({
    label: a.name,
    detail: a.trigger,
    state: t(lang, stateKey(a.enabled)),
  }));
}

function digitIndex(rune, n) {
  const idx = rune.charCodeAt(0) - '1'.charCodeAt(0);
  if (rune.length !== 1 || idx < 0 || idx >= n) {
    return -1;
  }
  return idx;
}

function toggleAlert(ctx, idx) {
  const items = allAlertItems(ctx.settings, ctx.lang);
  if (idx < 0 || idx >= items.length) {
    return true;
  }
  const item = items[idx];
  if (item.field) {
    ctx.settings.alerts[item.field] = !item.enabled;
  } else if (!toggleCustom(ctx.settings, item.id)) {
    return true;
  }
  return persistSettings(ctx);
}

function alertsMouse(ctx, pressed, items, clamped, sel) {
  if (pressed.key === Key.WHEEL_UP) {
    sel.value = moveSteps(items.length, clamped, -1);
    return 'alerts';
  }
  if (pressed.key === Key.WHEEL_DOWN) {
    sel.value = moveSteps(items.length, clamped, 1);
    return 'alerts';
  }
  if (pressed.key === Key.MOTION) {
    const idx = hitRowToIndex(items, clamped, frameHeight(ctx.rows), pressed.y);
    if (idx >= 0) {
      sel.value = idx;
    }
    return 'alerts';
  }
  if (pressed.key !== Key.CLICK) {
    return null;
  }
  const idx = hitRowToIndex(items, clamped, frameHeight(ctx.rows), pressed.y);
  if (idx < 0) {
    return 'alerts';
  }
  sel.value = idx;
  if (!toggleAlert(ctx, idx)) {
    return 'abort';
  }
  return 'alerts';
}

async function runAlertsView(ctx, ask, sel) {
  const items = alertScreenItems(ctx.settings, ctx.lang);
  const clamped = clampSelection(items.length, sel.value);
  sel.value = clamped;
  renderFrame(ctx, t(ctx.lang, 'alerts.title'), items, clamped, t(ctx.lang, 'footer.toggle'));
  const pressed = await ask();
  const mouseNext = alertsMouse(ctx, pressed, items, clamped, sel);
  if (mouseNext) {
    return mouseNext;
  }
  if (pressed.key === Key.RUNE) {
    const idx = digitIndex(pressed.rune, items.length);
    if (idx >= 0 && !toggleAlert(ctx, idx)) {
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

export {
  alertScreenItems,
  digitIndex,
  runAlertsView,
  persistSettings,
};
