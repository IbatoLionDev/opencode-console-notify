// Key model for the fullscreen loop: arrows and j/k navigate, Enter
// selects, Space toggles, digits and letters double as shortcuts.
// Byte decoding lives in keys_bytes.js (a mirror of Go ParseKey); this
// file keeps the string-level adapters and re-exports the model.
import {
  Key,
  parseKeyBytes,
  parseMouseBytes,
} from './keys_bytes.js';

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

// parseMouseSequence decodes one SGR mouse report string
// (ESC [ < Cb ; Cx ; Cy M/m): CLICK/WHEEL_UP/WHEEL_DOWN/MOTION with
// x/y, UNKNOWN for releases and unhandled buttons, undefined when
// the string is not a mouse report at all.
function parseMouseSequence(sequence) {
  if (typeof sequence !== 'string' || !sequence.startsWith('\x1b[<')) {
    return undefined;
  }
  const bytes = [...Buffer.from(sequence, 'utf8')];
  const parsed = parseMouseBytes(bytes);
  if (parsed.size === 0) {
    return { key: Key.UNKNOWN };
  }
  return { key: parsed.key, x: parsed.x, y: parsed.y };
}

// parseKeyPress decodes one keypress event (Node keypress or raw chunk)
// into { key, rune }. It encodes the sequence back to bytes so every
// path shares the byte parser; a bare event name only covers arrows.
function parseKeyPress(name, sequence) {
  if (typeof sequence === 'string' && sequence.startsWith('\x1b[<')) {
    const mouse = parseMouseSequence(sequence);
    if (mouse) {
      return { key: mouse.key, rune: '', x: mouse.x, y: mouse.y };
    }
    return { key: Key.UNKNOWN, rune: '' };
  }
  if (typeof sequence === 'string' && sequence.length > 0) {
    const parsed = parseKeyBytes([...Buffer.from(sequence, 'utf8')]);
    return { key: parsed.key, rune: parsed.rune, x: parsed.x, y: parsed.y };
  }
  const arrow = arrowKey(name, '');
  if (arrow) {
    return { key: arrow, rune: '' };
  }
  return { key: Key.UNKNOWN, rune: '' };
}

export {
  Key,
  parseKeyPress,
  parseMouseSequence,
};
