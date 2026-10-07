// Frame rendering and viewport scrolling: pure string building over the
// palette, node builtins only. Mirrors internal/screen frame helpers.
import {
  RESET,
  BOLD,
  DIM,
  FG_WHITE,
  FG_GRAY,
  FG_RED,
  BG_BLACK,
  BG_DARK_RED,
  HIDE_CURSOR,
  SHOW_CURSOR,
  ALT_ENTER,
  ALT_LEAVE,
  HOME,
  CLEAR,
  ERASE_BELOW,
  MOUSE_ENABLE,
  MOUSE_DISABLE,
} from './palette.js';

function linesPerItem(item) {
  return item.detail ? 2 : 1;
}

// fitBeforeWindow finds where the visible window starts so the selected
// row fits: rows accumulate until one overflows, restarting at (or just
// after) the overflowed row. Returns the start index and rows used.
function fitBeforeWindow(items, sel, height) {
  let start = 0;
  let used = 0;
  for (let i = 0; i <= sel; i++) {
    const need = linesPerItem(items[i]);
    if (used + need <= height) {
      used += need;
      continue;
    }
    if (i === sel) {
      start = i;
      used = need;
    } else {
      start = i + 1;
      used = 0;
    }
  }
  return { start, used };
}

function viewport(items, selected, height) {
  if (height <= 0 || items.length === 0) {
    return { visible: items, offset: 0 };
  }
  const sel = clampSelection(items.length, selected);
  const window = fitBeforeWindow(items, sel, height);
  let used = window.used;
  let end = items.length;
  for (let i = sel + 1; i < items.length; i++) {
    const need = linesPerItem(items[i]);
    if (used + need > height) {
      end = i;
      break;
    }
    used += need;
  }
  return { visible: items.slice(window.start, end), offset: window.start };
}

function border(width) {
  const w = width < 2 ? 2 : width;
  return `${FG_RED}+${'-'.repeat(w - 2)}+${RESET}`;
}

function padCell(text, width) {
  const chars = [...text];
  if (chars.length > width) {
    return chars.slice(0, width).join('');
  }
  return text + ' '.repeat(width - chars.length);
}

// render builds one full screen: home cursor, title, bordered item list
// with the selected row highlighted, and the footer. The `>` marker stays
// visible even on terminals that ignore colors.
function render(frame, width) {
  const w = width < 10 ? 10 : width;
  const clamped = clampSelection(frame.items.length, frame.selected);
  const { visible, offset } = viewport(frame.items, clamped, frame.height);
  const lines = [HOME + ERASE_BELOW, `${BG_BLACK}${FG_WHITE}`, `${BOLD}${frame.title}${RESET}${BG_BLACK}${FG_WHITE}`, border(w)];
  visible.forEach((item, i) => {
    const selected = offset + i === clamped;
    const marker = selected ? '> ' : '  ';
    const label = item.state ? `${item.label} [${item.state}]` : item.label;
    if (selected) {
      lines.push(`${BG_DARK_RED}${FG_WHITE}${BOLD}${marker}${padCell(label, w - marker.length)}${RESET}${BG_BLACK}${FG_WHITE}`);
    } else {
      lines.push(`${marker}${padCell(label, w - marker.length)}`);
    }
    if (item.detail) {
      // Dim wraps the padded text so every row ends on the same column.
      const detail = padCell(item.detail, w - 2);
      if (selected) {
        lines.push(`${BG_DARK_RED}${FG_WHITE}  ${DIM}${detail}${RESET}${BG_BLACK}${FG_WHITE}`);
      } else {
        lines.push(`  ${DIM}${detail}${RESET}${BG_BLACK}${FG_WHITE}`);
      }
    }
  });
  lines.push(border(w));
  lines.push(`${FG_GRAY}${frame.footer}${RESET}`);
  lines.push('');
  return lines.join('\n');
}

// INPUT_ROW and the drawn cursor live in cursor.js; mouse row mapping
// lives in mouse.js (both mirror their Go twins). This file keeps the
// pure frame: viewport, render, selection, geometry and frame bytes.

function clampSelection(n, selected) {
  if (n <= 0) return 0;
  if (selected < 0) return 0;
  if (selected >= n) return n - 1;
  return selected;
}

function moveSteps(n, selected, steps) {
  if (n <= 0) return 0;
  let next = (selected + steps) % n;
  if (next < 0) next += n;
  return next;
}

function frameHeight(termRows) {
  const h = termRows - 6;
  return h < 1 ? 1 : h;
}

function enterFrame() {
  return `${ALT_ENTER}${CLEAR}${HOME}${HIDE_CURSOR}${MOUSE_ENABLE}`;
}

function leaveFrame() {
  return `${MOUSE_DISABLE}${SHOW_CURSOR}${ALT_LEAVE}`;
}

export {
  linesPerItem,
  viewport,
  render,
  clampSelection,
  moveSteps,
  frameHeight,
  enterFrame,
  leaveFrame,
};
