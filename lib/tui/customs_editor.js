// Fullscreen customs editor: create and update flows. Text input lives
// in customs_input.js, the event picker in customs_picker.js; this file
// keeps the import path working via re-exports.
import { t } from '../i18n/dictionary.js';
import { knownEvents } from '../config/events.js';
import { addCustom, updateCustom } from '../config/customs.js';
import { persistSettings } from './fullscreen_views.js';
import { readLineInput } from './customs_input.js';
import { pickEvent } from './customs_picker.js';

async function customsAdd(ctx, ask) {
  const events = knownEvents();
  const idx = await pickEvent(ctx, ask, '');
  if (idx < 0) {
    return 'customs';
  }
  const title = await readLineInput(ctx, ask, t(ctx.lang, 'customs.title'), `${t(ctx.lang, 'customs.titleLabel')}: `);
  if (!title.ok || title.text === '') {
    return 'customs';
  }
  const body = await readLineInput(ctx, ask, t(ctx.lang, 'customs.title'), `${t(ctx.lang, 'customs.bodyLabel')} (${t(ctx.lang, 'customs.optional')}): `);
  if (!body.ok) {
    return 'customs';
  }
  try {
    addCustom(ctx.settings, events[idx].type, title.text, body.text);
  } catch (err) {
    ctx.output.write(`Error: ${err.message}\n`);
    return 'abort';
  }
  if (!persistSettings(ctx)) {
    return 'abort';
  }
  return 'customs';
}

async function customsEdit(ctx, ask, sel) {
  const list = ctx.settings.customAlerts || [];
  if (sel < 0 || sel >= list.length) {
    return 'customs';
  }
  const current = list[sel];
  const title = await readLineInput(ctx, ask, t(ctx.lang, 'customs.title'), `${t(ctx.lang, 'customs.titleLabel')} [${current.title}]: `);
  if (!title.ok) {
    return 'customs';
  }
  const body = await readLineInput(ctx, ask, t(ctx.lang, 'customs.title'), `${t(ctx.lang, 'customs.bodyLabel')} [${current.body}]: `);
  if (!body.ok) {
    return 'customs';
  }
  const idx = await pickEvent(ctx, ask, current.event);
  if (idx < 0) {
    return 'customs';
  }
  const events = knownEvents();
  try {
    const kept = updateCustom(ctx.settings, current.id, events[idx].type, title.text || current.title, body.text || current.body || '');
    if (!kept) {
      return 'customs';
    }
  } catch (err) {
    ctx.output.write(`Error: ${err.message}\n`);
    return 'abort';
  }
  if (!persistSettings(ctx)) {
    return 'abort';
  }
  return 'customs';
}

export {
  customsAdd,
  customsEdit,
};
export {
  readLineInput,
} from './customs_input.js';
export {
  pickEvent,
  pickEventView,
  catalogItems,
} from './customs_picker.js';
