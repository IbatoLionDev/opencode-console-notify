// Key model for the fullscreen loop: arrows and j/k navigate, Enter
// selects, Space toggles, digits and letters double as shortcuts.
// Mirrors internal/screen/keys.go.
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
  CLICK: 'click',
  WHEEL_UP: 'wheel_up',
  WHEEL_DOWN: 'wheel_down',
  MOTION: 'motion',
};

// arrowKey maps vim/arrows to a navigation key, or '' when the input is
// not an arrow. Single 'l' stays a plain rune: it is the language shortcut.
function arrowKey(name, sequence) {
  if (name === 'up' || sequence === '\x1b[A' || sequence === 'k') {
    return Key.UP;
  }
  if (name === 'down' || sequence === '\x1b[B' || sequence === 'j') {
    return Key.DOWN;
  }
  if (name === 'right' || sequence === '\x1b[C') {
    return Key.RIGHT;
  }
  if (name === 'left' || sequence === '\x1b[D' || sequence === 'h') {
    return Key.LEFT;
  }
  return '';
}

// parseMouseSequence decodes one SGR mouse report
// (ESC [ < Cb ; Cx ; Cy M/m). Returns null when the sequence is
// incomplete, { key: UNKNOWN } for releases and unhandled buttons,
// CLICK with x/y for left press, WHEEL_UP/DOWN with x/y for 64/65,
// MOTION with x/y for bit-5 reports (hover moves, never activates).
function parseMouseSequence(sequence) {
  if (!sequence.startsWith('\x1b[<')) {
    return undefined;
  }
  const final = sequence[sequence.length - 1];
  if (final !== 'M' && final !== 'm') {
    return { key: Key.UNKNOWN };
  }
  const body = sequence.slice(3, -1);
  const parts = body.split(';');
  if (parts.length !== 3 || parts.some((p) => !/^\d+$/.test(p))) {
    return { key: Key.UNKNOWN };
  }
  const cb = Number(parts[0]);
  const x = Number(parts[1]);
  const y = Number(parts[2]);
  if (final === 'm') {
    return { key: Key.UNKNOWN };
  }
  if (cb === 64) {
    return { key: Key.WHEEL_UP, x, y };
  }
  if (cb === 65) {
    return { key: Key.WHEEL_DOWN, x, y };
  }
  if ((cb & 32) !== 0) {
    return { key: Key.MOTION, x, y };
  }
  if ((cb & 0x43) === 0) {
    return { key: Key.CLICK, x, y };
  }
  return { key: Key.UNKNOWN };
}
// parseKeyPress decodes one keypress event (Node keypress or raw chunk)
// into { key, rune }. The raw single char is always preserved in rune so
// text fields can type shortcut letters (q/j/k/h/space navigate outside
// input, type inside it). Unknown CSI sequences stay unknown so the loop
// never spins. SGR mouse reports decode to CLICK/WHEEL_UP/WHEEL_DOWN/MOTION.
function parseKeyPress(name, sequence) {
  if (typeof sequence === 'string' && sequence.startsWith('\x1b[<')) {
    const mouse = parseMouseSequence(sequence);
    if (mouse) {
      return { key: mouse.key, rune: '', x: mouse.x, y: mouse.y };
    }
    return { key: Key.UNKNOWN, rune: '' };
  }
  const rune = sequence && sequence.length === 1 ? sequence : '';
  if (sequence === '\r' || sequence === '\n') {
    return { key: Key.ENTER, rune: '' };
  }
  if (sequence === ' ') {
    return { key: Key.SPACE, rune: ' ' };
  }
  const arrow = arrowKey(name, sequence);
  if (arrow) {
    return { key: arrow, rune };
  }
  if (name === 'escape' || sequence === '\x1b') {
    return { key: Key.ESC, rune: '' };
  }
  if (name === 'q' || sequence === 'q' || sequence === 'Q') {
    return { key: Key.QUIT, rune };
  }
  if (rune) {
    return { key: Key.RUNE, rune };
  }
  return { key: Key.UNKNOWN, rune: '' };
}

export {
  Key,
  parseKeyPress,
  parseMouseSequence,
};
