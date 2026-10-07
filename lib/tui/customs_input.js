// Fullscreen customs text input: single-line echo with backspace,
// staying inside the alternate screen. Picker lives in
// customs_picker.js; add/edit flows in customs_editor.js.
import { t } from '../i18n/dictionary.js';
import { moveCursor, inputCursorCol, cursorBlock, INPUT_ROW, Key, HIDE_CURSOR } from '../screen/screen.js';
import { renderFrame } from './fullscreen_views.js';

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

export {
  promptFrame,
  readLineInput,
};
