// Palette: black background, soft-white body, dark-red selection with
// bright-white text, dim-red borders, light-gray footer. Red is reserved
// for selection and borders so the screen never looks saturated.
// Mirrors internal/screen palette consts.
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
};
