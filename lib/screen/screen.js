// Folder facade: palette, frame and key model behind one import path.
// Mirrors internal/screen as a package.
export {
  RESET,
  BOLD,
  DIM,
  FG_WHITE,
  FG_GRAY,
  FG_RED,
  BG_BLACK,
  BG_DARK_RED,
  BG_WHITE,
  MOUSE_ENABLE,
  MOUSE_DISABLE,
  HIDE_CURSOR,
  SHOW_CURSOR,
  ALT_ENTER,
  ALT_LEAVE,
  HOME,
  CLEAR,
  ERASE_BELOW,
} from './palette.js';
export {
  INPUT_ROW,
  FIRST_ITEM_ROW,
  moveCursor,
  inputCursorCol,
  cursorBlock,
  viewport,
  hitRowToIndex,
  render,
  clampSelection,
  moveSteps,
  frameHeight,
  enterFrame,
  leaveFrame,
} from './frame.js';
export {
  Key,
  parseKeyPress,
  parseMouseSequence,
} from './keys.js';
