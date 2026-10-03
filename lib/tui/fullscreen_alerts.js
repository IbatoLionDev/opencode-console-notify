// Fullscreen alerts view: the four default toggles rendered as a frame.
import { t } from '../i18n/dictionary.js';
import {
  clampSelection,
  moveSteps,
  Key,
} from '../screen/screen.js';
import { alertItems } from './tui.js';
import { renderFrame, persistSettings } from './fullscreen_views.js';

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

// alertScreenItems maps the shared alertItems domain over screen rows so
// the four alerts are defined once (tui.js owns the mapping).
function alertScreenItems(settings, lang) {
  return alertItems(settings, lang).map((a) => ({
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
  ctx.settings.alerts[alertField(idx)] = !ctx.settings.alerts[alertField(idx)];
  return persistSettings(ctx);
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

export {
  alertField,
  alertScreenItems,
  digitIndex,
  runAlertsView,
  persistSettings,
};
