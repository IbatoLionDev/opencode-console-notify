// Drawn cursor helpers: the hardware cursor stays hidden during text
// input, so the prompt frame draws a white block at the insertion point
// and parks the hardware cursor inside the input row after every render.
// Mirrors internal/screen/cursor.go.
import {
  BG_WHITE,
  RESET,
} from './palette.js';

// INPUT_ROW is the 1-indexed terminal row of the input row in a
// single-item prompt frame. render joins HOME+ERASE and the palette codes
// as their own newline-terminated lines, so they consume rows 1-2 before
// the visible frame starts (title=3, top border=4, input=5).
const INPUT_ROW = 5;

// cursorBlock returns the drawn insertion-point block: a white-bg ASCII
// space (solid white cell, zero wide-rune risk) followed by reset. The
// hardware cursor stays hidden during text input; this block is written
// at the insertion point after every prompt-frame render instead.
function cursorBlock() {
  return `${BG_WHITE} ${RESET}`;
}

// moveCursor returns the ANSI CUP sequence placing the hardware cursor
// at the given 1-indexed row and column, clamping each to >= 1.
function moveCursor(row, col) {
  const r = row < 1 ? 1 : row;
  const c = col < 1 ? 1 : col;
  return `\x1b[${r};${c}H`;
}

// inputCursorCol returns the 1-indexed column just after prompt+typed in
// the input row: 1 + marker width ("> ") + prompt runes + typed runes,
// clamped to [1, effectiveWidth] where effectiveWidth mirrors render.
function inputCursorCol(prompt, typedLen, width) {
  const w = width < 10 ? 10 : width;
  let col = 1 + 2 + [...prompt].length + typedLen;
  if (col < 1) col = 1;
  if (col > w) col = w;
  return col;
}

export {
  INPUT_ROW,
  moveCursor,
  inputCursorCol,
  cursorBlock,
};
