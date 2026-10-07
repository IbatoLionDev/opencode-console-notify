// Byte-level key model: decodes raw stdin bytes, mirroring Go ParseKey
// (internal/screen/keys.go + keys_mouse.go). The fullscreen loop reads
// raw data chunks through createAsk (fullscreen.js) so SGR mouse reports
// arrive whole; Node keypress events split them and break hover/click.
// Single 'l' stays a plain rune here too: it is the language shortcut,
// while Go maps it to Right (see keys.js).
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

// parseKeyBytes decodes one keypress from the head of raw stdin bytes.
// Returns { key, rune, x, y, size }: size is the bytes consumed and 0
// means more bytes may still complete the sequence.
function parseKeyBytes(bytes) {
  if (bytes.length === 0) {
    return { key: Key.UNKNOWN, rune: '', size: 0 };
  }
  const b0 = bytes[0];
  if (b0 === 0x0d || b0 === 0x0a) {
    return { key: Key.ENTER, rune: '', size: 1 };
  }
  if (b0 === 0x20) {
    return { key: Key.SPACE, rune: ' ', size: 1 };
  }
  if (b0 === 0x1b) {
    return parseEscapeBytes(bytes);
  }
  const [rune, size] = decodeRuneBytes(bytes);
  switch (rune) {
    case 'q':
    case 'Q':
      return { key: Key.QUIT, rune, size };
    case 'k':
    case 'K':
      return { key: Key.UP, rune, size };
    case 'j':
    case 'J':
      return { key: Key.DOWN, rune, size };
    case 'h':
    case 'H':
      return { key: Key.LEFT, rune, size };
    default:
      return { key: Key.RUNE, rune, size };
  }
}

function parseEscapeBytes(bytes) {
  if (bytes.length === 1) {
    return { key: Key.ESC, rune: '', size: 1 };
  }
  if (bytes.length === 2) {
    return { key: Key.UNKNOWN, rune: '', size: 0 };
  }
  // SGR mouse: ESC [ < Cb ; Cx ; Cy M/m.
  if (bytes[1] === 0x5b && bytes.length > 3 && bytes[2] === 0x3c) {
    return parseMouseBytes(bytes);
  }
  if (bytes[1] !== 0x5b && bytes[1] !== 0x4f) {
    return { key: Key.ESC, rune: '', size: 1 };
  }
  switch (bytes[2]) {
    case 0x41:
      return { key: Key.UP, rune: '', size: 3 };
    case 0x42:
      return { key: Key.DOWN, rune: '', size: 3 };
    case 0x43:
      return { key: Key.RIGHT, rune: '', size: 3 };
    case 0x44:
      return { key: Key.LEFT, rune: '', size: 3 };
    default:
      return { key: Key.UNKNOWN, rune: '', size: 3 };
  }
}

// parseMouseBytes decodes one SGR mouse report: wheel is Cb 64/65,
// motion sets bit 5, left click is Cb 0 (+modifier bits). Release
// (m) stays unknown so one press never double-fires.
function parseMouseBytes(bytes) {
  const found = mouseEndBytes(bytes);
  if (found.end < 0) {
    return { key: Key.UNKNOWN, size: 0 };
  }
  if (found.malformed) {
    return { key: Key.UNKNOWN, size: 3 };
  }
  const body = String.fromCharCode(...bytes.slice(3, found.end));
  const parts = body.split(';');
  if (parts.length !== 3 || parts.some((p) => !/^\d+$/.test(p))) {
    return { key: Key.UNKNOWN, size: found.end + 1 };
  }
  const cb = Number(parts[0]);
  const x = Number(parts[1]);
  const y = Number(parts[2]);
  const size = found.end + 1;
  if (bytes[found.end] === 0x6d) {
    return { key: Key.UNKNOWN, size };
  }
  if (cb === 64) {
    return { key: Key.WHEEL_UP, x, y, size };
  }
  if (cb === 65) {
    return { key: Key.WHEEL_DOWN, x, y, size };
  }
  if ((cb & 32) !== 0) {
    return { key: Key.MOTION, x, y, size };
  }
  if ((cb & 0x43) === 0) {
    return { key: Key.CLICK, x, y, size };
  }
  return { key: Key.UNKNOWN, size };
}

// mouseEndBytes scans the SGR body for its M/m terminator: end is the
// terminator index (-1 while more bytes may complete it), malformed
// when a byte outside digits and ';' can never belong to a report.
function mouseEndBytes(bytes) {
  for (let i = 3; i < bytes.length; i++) {
    const c = bytes[i];
    if (c === 0x4d || c === 0x6d) {
      return { end: i, malformed: false };
    }
    if (!((c >= 0x30 && c <= 0x39) || c === 0x3b)) {
      return { end: i, malformed: true };
    }
  }
  return { end: -1, malformed: false };
}

// decodeRuneBytes decodes one UTF-8 rune like Go decodeRune: invalid
// bytes become U+FFFD of size 1 so the loop always advances.
function decodeRuneBytes(bytes) {
  const c = bytes[0];
  if (c < 0x80) {
    return [String.fromCharCode(c), 1];
  }
  let size = 0;
  let min = 0;
  let mask = 0;
  if ((c >> 5) === 0x06) {
    size = 2;
    min = 0x80;
    mask = 0x1f;
  } else if ((c >> 4) === 0x0e) {
    size = 3;
    min = 0x800;
    mask = 0x0f;
  } else if ((c >> 3) === 0x1e) {
    size = 4;
    min = 0x10000;
    mask = 0x07;
  } else {
    return ['�', 1];
  }
  if (bytes.length < size) {
    return ['�', 1];
  }
  let r = c & mask;
  for (let i = 1; i < size; i++) {
    if ((bytes[i] >> 6) !== 0x02) {
      return ['�', 1];
    }
    r = (r << 6) | (bytes[i] & 0x3f);
  }
  if (r < min || (r >= 0xd800 && r <= 0xdfff)) {
    return ['�', 1];
  }
  return [String.fromCodePoint(r), size];
}

export {
  Key,
  parseKeyBytes,
  parseMouseBytes,
};
