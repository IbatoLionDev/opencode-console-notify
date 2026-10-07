// Fullscreen customs event picker: scrollable catalog of subscribable
// events with click and wheel support. Text input lives in
// customs_input.js; add/edit flows in customs_editor.js.
import { t } from '../i18n/dictionary.js';
import { moveSteps, frameHeight, hitRowToIndex, Key } from '../screen/screen.js';
import { knownEvents } from '../config/events.js';
import { renderFrame } from './fullscreen_views.js';

function catalogItems(lang) {
  return knownEvents().map((e) => ({
    label: e.type,
    detail: lang === 'es' ? e.descEs : e.descEn,
    state: e.noisy ? '!' : '',
  }));
}

function pickEventMouse(pressed, events, catalog, sel, height) {
  if (pressed.key === Key.WHEEL_UP) {
    return { sel: moveSteps(events.length, sel, -1), pick: -1, handled: true };
  }
  if (pressed.key === Key.WHEEL_DOWN) {
    return { sel: moveSteps(events.length, sel, 1), pick: -1, handled: true };
  }
  if (pressed.key === Key.MOTION) {
    const idx = hitRowToIndex(catalog, sel, height, pressed.y);
    if (idx < 0) {
      return { sel, pick: -1, handled: true };
    }
    return { sel: idx, pick: -1, handled: true };
  }
  if (pressed.key !== Key.CLICK) {
    return { sel, pick: -1, handled: false };
  }
  const idx = hitRowToIndex(catalog, sel, height, pressed.y);
  if (idx < 0) {
    return { sel, pick: -1, handled: true };
  }
  return { sel: idx, pick: idx, handled: true };
}

async function pickEvent(ctx, ask, preselect) {
  const events = knownEvents();
  let sel = 0;
  const pre = events.findIndex((e) => e.type === preselect);
  if (pre >= 0) {
    sel = pre;
  }
  for (;;) {
    const catalog = catalogItems(ctx.lang);
    renderFrame(ctx, t(ctx.lang, 'customs.pickEvent'), catalog, sel, t(ctx.lang, 'footer.back'));
    const pressed = await ask();
    const mouse = pickEventMouse(pressed, events, catalog, sel, frameHeight(ctx.rows));
    if (mouse.handled) {
      if (mouse.pick >= 0) {
        return mouse.pick;
      }
      sel = mouse.sel;
      continue;
    }
    if (pressed.key === Key.UP) {
      sel = moveSteps(events.length, sel, -1);
    } else if (pressed.key === Key.DOWN) {
      sel = moveSteps(events.length, sel, 1);
    } else if (pressed.key === Key.ENTER || pressed.key === Key.SPACE) {
      return sel;
    } else if (pressed.key === Key.QUIT || pressed.key === Key.ESC) {
      return -1;
    }
  }
}

async function pickEventView(ctx, ask) {
  await pickEvent(ctx, ask, '');
  return 'customs';
}

export {
  catalogItems,
  pickEvent,
  pickEventView,
};
