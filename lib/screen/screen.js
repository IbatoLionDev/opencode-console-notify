// Domain: fullscreen TUI primitives (palette, frames, viewport, key model).
// Pure string building, node builtins only. Mirrors internal/screen.

const RESET = '\x1b[0m';
const BOLD = '\x1b[1m';
const DIM = '\x1b[2m';
const FG_WHITE = '\x1b[97m';
const FG_GRAY = '\x1b[90m';
const FG_RED = '\x1b[31m';
const BG_BLACK = '\x1b[40m';
const BG_DARK_RED = '\x1b[48;5;88m';
const HIDE_CURSOR = '\x1b[?25l';
const SHOW_CURSOR = '\x1b[?25h';
const ALT_ENTER = '\x1b[?1049h';
const ALT_LEAVE = '\x1b[?1049l';
const HOME = '\x1b[H';
const CLEAR = '\x1b[2J';

const Key = {
  UNKNOWN: 'unknown',
  UP: 'up',
  DOWN: 'down',
  LEFT: 'left',
  RIGHT: 'right',
  ENTER: 'enter',
  SPACE: 'space',
  ESC: 'esc',
  QUIT: 'quit',
  RUNE: 'rune',
};

function linesPerItem(item) {
  return item.detail ? 2 : 1;
}

// Viewport slices items to the rows around selected so the selection stays
// visible. Returns { visible, offset }. Height <= 0 disables scrolling.
function viewport(items, selected, height) {
  if (height <= 0 || items.length === 0) {
    return { visible: items, offset: 0 };
  }
  let sel = selected;
  if (sel < 0) sel = 0;
  if (sel >= items.length) sel = items.length - 1;
  let start = 0;
  let used = 0;
  for (let i = 0; i < items.length; i++) {
    const need = linesPerItem(items[i]);
    if (i < sel) {
      if (used + need > height) {
        start = i + 1;
        used = 0;
      } else {
        used += need;
      }
      continue;
    }
    if (i === sel) {
      if (used + need > height) {
        start = i;
        used = need;
      } else {
        used += need;
      }
      continue;
    }
    if (used + need > height) {
      return { visible: items.slice(start, i), offset: start };
    }
    used += need;
  }
  return { visible: items.slice(start), offset: start };
}

function border(width) {
  const w = width < 2 ? 2 : width;
  return `${FG_RED}+${'-'.repeat(w - 2)}+${RESET}`;
}

function padCell(text, width) {
  if (text.length > width) {
    return text.slice(0, width);
  }
  return text + ' '.repeat(width - text.length);
}

// render builds one full screen: home cursor, title, bordered item list
// with the selected row highlighted, and the footer. The `>` marker stays
// visible even on terminals that ignore colors.
function render(frame, width) {
  const w = width < 10 ? 10 : width;
  const clamped = clampSelection(frame.items.length, frame.selected);
  const { visible, offset } = viewport(frame.items, clamped, frame.height);
  const lines = [HOME, `${BG_BLACK}${FG_WHITE}`, `${BOLD}${frame.title}${RESET}${BG_BLACK}${FG_WHITE}`, border(w)];
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
      if (selected) {
        lines.push(`${BG_DARK_RED}${FG_WHITE}  ${padCell(`${DIM}${item.detail}`, w - 2)}${RESET}${BG_BLACK}${FG_WHITE}`);
      } else {
        lines.push(`  ${DIM}${padCell(item.detail, w - 2)}${RESET}${BG_BLACK}${FG_WHITE}`);
      }
    }
  });
  lines.push(border(w));
  lines.push(`${FG_GRAY}${frame.footer}${RESET}`);
  lines.push('');
  return lines.join('\n');
}

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
  return `${ALT_ENTER}${CLEAR}${HOME}${HIDE_CURSOR}`;
}

function leaveFrame() {
  return `${SHOW_CURSOR}${ALT_LEAVE}`;
}

// parseKeyPress decodes one keypress event (Node keypress or raw chunk)
// into { key, rune }. Unknown CSI sequences stay unknown so the loop
// never spins.
function parseKeyPress(name, sequence) {
  if (sequence === '\r' || sequence === '\n') {
    return { key: Key.ENTER, rune: '' };
  }
  if (sequence === ' ') {
    return { key: Key.SPACE, rune: '' };
  }
  if (name === 'up' || sequence === '\x1b[A' || sequence === 'k') {
    return { key: Key.UP, rune: sequence.length === 1 ? sequence : '' };
  }
  if (name === 'down' || sequence === '\x1b[B' || sequence === 'j') {
    return { key: Key.DOWN, rune: sequence.length === 1 ? sequence : '' };
  }
  if (name === 'right' || sequence === '\x1b[C') {
    return { key: Key.RIGHT, rune: '' };
  }
  if (name === 'left' || sequence === '\x1b[D' || sequence === 'h') {
    return { key: Key.LEFT, rune: sequence.length === 1 ? sequence : '' };
  }
  if (name === 'escape' || sequence === '\x1b') {
    return { key: Key.ESC, rune: '' };
  }
  if (name === 'q' || sequence === 'q' || sequence === 'Q') {
    return { key: Key.QUIT, rune: sequence };
  }
  if (sequence && sequence.length === 1) {
    return { key: Key.RUNE, rune: sequence };
  }
  return { key: Key.UNKNOWN, rune: '' };
}

export {
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
  Key,
  viewport,
  render,
  clampSelection,
  moveSteps,
  frameHeight,
  enterFrame,
  leaveFrame,
  parseKeyPress,
};
