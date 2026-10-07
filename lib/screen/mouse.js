// SGR mouse row mapping: terminal rows back to item indexes through
// the viewport. Mirrors internal/screen/mouse.go.
import {
  linesPerItem,
  viewport,
  clampSelection,
} from './frame.js';

// FIRST_ITEM_ROW is the 1-indexed terminal row of the first item in
// a list frame. render joins HOME+ERASE and the palette codes as
// their own lines, so they consume rows 1-2 before the visible
// frame (title=3, border=4, items start at 5). Go has no extra
// lines, so its first item row is 3 (see ItemFirstRow).
const FIRST_ITEM_ROW = 5;

function hitRowToIndex(items, selected, height, termRow) {
  if (!items.length) {
    return -1;
  }
  const rel = termRow - FIRST_ITEM_ROW;
  if (rel < 0) {
    return -1;
  }
  const clamped = clampSelection(items.length, selected);
  const { visible, offset } = viewport(items, clamped, height);
  let used = 0;
  for (let i = 0; i < visible.length; i++) {
    const need = linesPerItem(visible[i]);
    if (rel >= used && rel < used + need) {
      return offset + i;
    }
    used += need;
  }
  return -1;
}

export {
  FIRST_ITEM_ROW,
  hitRowToIndex,
};
