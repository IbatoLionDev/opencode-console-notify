// Fullscreen customs event picker: scrollable filterable catalog of
// subscribable events with type-to-filter search plus click and wheel.
// Text input lives in customs_input.js; add/edit flows in
// customs_editor.js.
import { t } from '../i18n/dictionary.js';
import { clampSelection, moveSteps, frameHeight, hitRowToIndex, Key } from '../screen/screen.js';
import { knownEvents } from '../config/events.js';
import { renderFrame } from './fullscreen_views.js';

function catalogItems(lang) {
  return knownEvents().map((e) => ({
    label: e.type,
    detail: lang === 'es' ? e.descEs : e.descEn,
    state: e.noisy ? '!' : '',
  }));
}

// pickerRows filters the catalog rows for one query, returning the rows
// plus the event index behind each row. Matching is a strict contiguous
// substring (case-insensitive) from any position; empty keeps everything.
function pickerRows(lang, query) {
  const all = catalogItems(lang);
  const q = query.toLowerCase();
  const items = [];
  const indexMap = [];
  all.forEach((item, i) => {
    const label = (item.label || '').toLowerCase();
    const detail = (item.detail || '').toLowerCase();
    if (q !== '' && !label.includes(q) && !detail.includes(q)) {
      return;
    }
    items.push(item);
    indexMap.push(i);
  });
  return { items, indexMap };
}

// pickEventType updates the filter query from one keypress: printable
// runes append (q included, which types here instead of quitting),
// Backspace trims rune-aware. Returns the new query and whether it typed.
function pickEventType(pressed, query) {
  if (pressed.key !== Key.RUNE && pressed.key !== Key.QUIT) {
    return { query, handled: false };
  }
  const rune = pressed.rune || '';
  if (rune === '\x7f' || rune === '\b') {
    if (query === '') {
      return { query, handled: false };
    }
    return { query: [...query].slice(0, -1).join(''), handled: true };
  }
  if (rune.length !== 1 || rune < ' ') {
    return { query, handled: false };
  }
  return { query: query + rune, handled: true };
}

function pickEventMouse(pressed, items, indexMap, sel, height) {
  if (pressed.key === Key.WHEEL_UP) {
    return { sel: moveSteps(items.length, sel, -1), pick: -1, handled: true };
  }
  if (pressed.key === Key.WHEEL_DOWN) {
    return { sel: moveSteps(items.length, sel, 1), pick: -1, handled: true };
  }
  if (pressed.key === Key.MOTION) {
    const idx = hitRowToIndex(items, sel, height, pressed.y);
    if (idx < 0) {
      return { sel, pick: -1, handled: true };
    }
    return { sel: idx, pick: -1, handled: true };
  }
  if (pressed.key !== Key.CLICK) {
    return { sel, pick: -1, handled: false };
  }
  const idx = hitRowToIndex(items, sel, height, pressed.y);
  if (idx < 0) {
    return { sel, pick: -1, handled: true };
  }
  return { sel: idx, pick: indexMap[idx], handled: true };
}

// pickEventNav moves the selection or picks the highlighted row for one
// navigation key. Enter/Space on an empty list stays put.
function pickEventNav(pressed, sel, items, indexMap) {
  if (pressed.key === Key.UP) {
    return { sel: moveSteps(items.length, sel, -1), pick: -1, handled: true };
  }
  if (pressed.key === Key.DOWN) {
    return { sel: moveSteps(items.length, sel, 1), pick: -1, handled: true };
  }
  if (pressed.key === Key.ENTER || pressed.key === Key.SPACE) {
    if (items.length === 0) {
      return { sel, pick: -1, handled: true };
    }
    return { sel, pick: indexMap[sel], handled: true };
  }
  return { sel, pick: -1, handled: false };
}

// pickerFooter renders the picker hint plus the live filter status.
function pickerFooter(lang, query, shown, total) {
  const base = t(lang, 'customs.pickFooter');
  if (query === '') {
    return base;
  }
  return `${base} | Filter "${query}" ${shown}/${total}`;
}

// pickEventStep routes one picker keypress through mouse, typing, Esc
// and navigation. Returns the query, selection, picked event index and
// whether the picker is done (picked, or backed out with pick -1).
function pickEventStep(pressed, query, sel, items, indexMap, height) {
  const mouse = pickEventMouse(pressed, items, indexMap, sel, height);
  if (mouse.handled) {
    if (mouse.pick >= 0) {
      return { query, sel: mouse.sel, picked: mouse.pick, done: true };
    }
    return { query, sel: mouse.sel, picked: -1, done: false };
  }
  const typed = pickEventType(pressed, query);
  if (typed.handled) {
    return { query: typed.query, sel: typed.query !== query ? 0 : sel, picked: -1, done: false };
  }
  if (pressed.key === Key.ESC) {
    if (query !== '') {
      return { query: '', sel: 0, picked: -1, done: false };
    }
    return { query, sel, picked: -1, done: true };
  }
  const nav = pickEventNav(pressed, sel, items, indexMap);
  if (nav.handled) {
    if (nav.pick >= 0) {
      return { query, sel: nav.sel, picked: nav.pick, done: true };
    }
    return { query, sel: nav.sel, picked: -1, done: false };
  }
  return { query, sel, picked: -1, done: false };
}

async function pickEvent(ctx, ask, preselect) {
  const events = knownEvents();
  let sel = 0;
  const pre = events.findIndex((e) => e.type === preselect);
  if (pre >= 0) {
    sel = pre;
  }
  const total = events.length;
  let query = '';
  for (;;) {
    const { items, indexMap } = pickerRows(ctx.lang, query);
    sel = clampSelection(items.length, sel);
    renderFrame(ctx, t(ctx.lang, 'customs.pickEvent'), items, sel, pickerFooter(ctx.lang, query, items.length, total));
    const pressed = await ask();
    const step = pickEventStep(pressed, query, sel, items, indexMap, frameHeight(ctx.rows));
    query = step.query;
    sel = step.sel;
    if (step.done) {
      return step.picked;
    }
  }
}

async function pickEventView(ctx, ask) {
  await pickEvent(ctx, ask, '');
  return 'customs';
}

export {
  catalogItems,
  pickerRows,
  pickEvent,
  pickEventView,
};
