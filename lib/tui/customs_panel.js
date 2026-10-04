// Fullscreen customs panel: list, toggle and delete over user
// notifications. Add/edit flows live in customs_editor.js.
import { t } from '../i18n/dictionary.js';
import {
  clampSelection,
  moveSteps,
  Key,
} from '../screen/screen.js';
import { toggleCustom, removeCustom } from '../config/customs.js';
import { persistSettings, renderFrame } from './fullscreen_views.js';
import { customsAdd, customsEdit, pickEventView, readLineInput } from './customs_editor.js';

function customsPanelItems(settings, lang) {
  const list = (settings.customAlerts || []).map((c) => ({
    label: `${c.title} [${t(lang, 'custom.tag')}]`,
    detail: c.body ? `${c.event} — ${c.body}` : c.event,
    state: t(lang, c.enabled ? 'state.on' : 'state.off'),
  }));
  if (list.length === 0) {
    list.push({ label: t(lang, 'customs.empty') });
  }
  return list;
}

async function runCustomsView(ctx, ask, sel) {
  const customs = ctx.settings.customAlerts || [];
  sel.value = clampSelection(customs.length, sel.value);
  renderFrame(ctx, t(ctx.lang, 'customs.title'), customsPanelItems(ctx.settings, ctx.lang), sel.value, t(ctx.lang, 'customs.footer'));
  const pressed = await ask();
  if (pressed.key === Key.RUNE) {
    return customsRuneKey(ctx, ask, pressed.rune, sel.value);
  }
  if (pressed.key === Key.UP) {
    sel.value = moveSteps(customs.length, sel.value, -1);
  } else if (pressed.key === Key.DOWN) {
    sel.value = moveSteps(customs.length, sel.value, 1);
  } else if (pressed.key === Key.ENTER || pressed.key === Key.SPACE) {
    if (!toggleCustomAt(ctx, sel.value)) {
      return 'abort';
    }
  } else if (pressed.key === Key.QUIT || pressed.key === Key.ESC) {
    return 'menu';
  }
  return 'customs';
}

async function customsRuneKey(ctx, ask, rune, sel) {
  if (rune === 'a' || rune === 'A') {
    return customsAdd(ctx, ask);
  }
  if (rune === 'e' || rune === 'E') {
    return customsEdit(ctx, ask, sel);
  }
  if (rune === 'd' || rune === 'D') {
    return customsDelete(ctx, ask, sel);
  }
  if (rune === 'c' || rune === 'C') {
    await pickEventView(ctx, ask);
    return 'customs';
  }
  const idx = digitIndex(rune, (ctx.settings.customAlerts || []).length);
  if (idx >= 0 && !toggleCustomAt(ctx, idx)) {
    return 'abort';
  }
  return 'customs';
}

function toggleCustomAt(ctx, idx) {
  const list = ctx.settings.customAlerts || [];
  if (idx < 0 || idx >= list.length) {
    return true;
  }
  if (!toggleCustom(ctx.settings, list[idx].id)) {
    return true;
  }
  return persistSettings(ctx);
}

async function customsDelete(ctx, ask, sel) {
  const list = ctx.settings.customAlerts || [];
  if (sel < 0 || sel >= list.length) {
    return 'customs';
  }
  const answer = await readLineInput(ctx, ask, t(ctx.lang, 'customs.confirmDelete').replace('%s', list[sel].title));
  if (!answer.ok) {
    return 'customs';
  }
  if (answer.text !== 'y' && answer.text !== 's') {
    return 'customs';
  }
  if (!removeCustom(ctx.settings, list[sel].id)) {
    return 'customs';
  }
  if (!persistSettings(ctx)) {
    return 'abort';
  }
  return 'customs';
}

function digitIndex(rune, n) {
  const idx = rune.charCodeAt(0) - '1'.charCodeAt(0);
  if (rune.length !== 1 || idx < 0 || idx >= n) {
    return -1;
  }
  return idx;
}

export {
  customsPanelItems,
  runCustomsView,
};
