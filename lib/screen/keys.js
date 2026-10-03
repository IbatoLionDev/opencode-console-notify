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
  const arrow = arrowKey(name, sequence);
  if (arrow) {
    return { key: arrow, rune: '' };
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
  Key,
  parseKeyPress,
};
