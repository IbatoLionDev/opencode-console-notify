// Fullscreen customs editor: event picker plus single-line text input
// with echo and backspace, staying inside the alternate screen.
import { t } from '../i18n/dictionary.js';
import { Key } from '../screen/screen.js';
import { knownEvents } from '../config/events.js';
import { addCustom, updateCustom } from '../config/customs.js';
import { persistSettings, renderFrame } from './fullscreen_views.js';

async function readLineInput(ctx, ask, prompt) {
  ctx.output.write(prompt);
  let chars = [];
  for (;;) {
    const pressed = await ask();
    if (pressed.key === Key.ENTER) {
      ctx.output.write('\r\n');
      return { text: chars.join('').trim(), ok: true };
    }
    if (pressed.key === Key.ESC) {
      ctx.output.write('\r\n');
      return { text: '', ok: false };
    }
    if (pressed.key !== Key.RUNE) {
      continue;
    }
    if (pressed.rune === '\x7f' || pressed.rune === '\b') {
      if (chars.length > 0) {
        chars = [...chars.join('')].slice(0, -1);
        ctx.output.write('\b \b');
      }
      continue;
    }
    if (pressed.rune >= ' ') {
      chars = [...chars.join(''), pressed.rune];
      ctx.output.write(pressed.rune);
    }
  }
}

function catalogItems(lang) {
  return knownEvents().map((e) => ({
    label: e.type,
    detail: lang === 'es' ? e.descEs : e.descEn,
    state: e.noisy ? '!' : '',
  }));
}

async function pickEvent(ctx, ask, preselect) {
  const events = knownEvents();
  let sel = 0;
  const pre = events.findIndex((e) => e.type === preselect);
  if (pre >= 0) {
    sel = pre;
  }
  for (;;) {
    renderFrame(ctx, t(ctx.lang, 'customs.pickEvent'), catalogItems(ctx.lang), sel, t(ctx.lang, 'footer.back'));
    const pressed = await ask();
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

async function customsAdd(ctx, ask) {
  const events = knownEvents();
  const idx = await pickEvent(ctx, ask, '');
  if (idx < 0) {
    return 'customs';
  }
  const title = await readLineInput(ctx, ask, `${t(ctx.lang, 'customs.titleLabel')}: `);
  if (!title.ok || title.text === '') {
    return 'customs';
  }
  const body = await readLineInput(ctx, ask, `${t(ctx.lang, 'customs.bodyLabel')} (${t(ctx.lang, 'customs.optional')}): `);
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
  const title = await readLineInput(ctx, ask, `${t(ctx.lang, 'customs.titleLabel')} [${current.title}]: `);
  if (!title.ok) {
    return 'customs';
  }
  const body = await readLineInput(ctx, ask, `${t(ctx.lang, 'customs.bodyLabel')} [${current.body}]: `);
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
  readLineInput,
  pickEvent,
  pickEventView,
  customsAdd,
  customsEdit,
};
