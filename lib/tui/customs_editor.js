// Fullscreen customs editor: event picker plus single-line text input
// with echo and backspace, staying inside the alternate screen.
import { t } from '../i18n/dictionary.js';
import { moveSteps, moveCursor, inputCursorCol, cursorBlock, INPUT_ROW, Key, HIDE_CURSOR } from '../screen/screen.js';
import { knownEvents } from '../config/events.js';
import { addCustom, updateCustom } from '../config/customs.js';
import { persistSettings, renderFrame } from './fullscreen_views.js';

// promptFrame renders the text prompt as a frame so input never leaves
// the alternate screen; every keystroke redraws with the buffer.
function promptFrame(ctx, title, prompt, text) {
  renderFrame(ctx, title, [{ label: prompt + text }], 0, t(ctx.lang, 'customs.inputFooter'));
  // renderFrame ends with footer + trailing newline, so reposition the
  // hardware cursor inside the input row after prompt+typed text.
  const col = inputCursorCol(prompt, [...text].length, ctx.cols);
  ctx.output.write(moveCursor(INPUT_ROW, col));
  // Drawn white block marks the insertion point; the hardware cursor
  // stays hidden (enterFrame hides, leaveFrame restores).
  ctx.output.write(cursorBlock());
}

async function readLineInput(ctx, ask, frameTitle, prompt) {
  try {
    let chars = [];
    for (;;) {
      promptFrame(ctx, frameTitle, prompt, chars.join(''));
      const pressed = await ask();
      // Printable runes type even when they double as shortcuts
      // elsewhere (q/j/k/h navigate outside input, type inside it).
      const rune = pressed.rune || '';
      if (pressed.key === Key.ENTER) {
        return { text: chars.join('').trim(), ok: true };
      }
      if (pressed.key === Key.ESC) {
        return { text: '', ok: false };
      }
      if (rune === '\x7f' || rune === '\b') {
        if (chars.length > 0) {
          chars = [...chars.join('')].slice(0, -1);
        }
        continue;
      }
      if (rune.length === 1 && rune >= ' ') {
        chars = [...chars.join(''), rune];
      }
    }
  } finally {
    ctx.output.write(HIDE_CURSOR);
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
  readLineInput,
  pickEvent,
  pickEventView,
  customsAdd,
  customsEdit,
};
