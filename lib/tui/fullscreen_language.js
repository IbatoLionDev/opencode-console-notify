// Fullscreen language view: English/Spanish picker rendered as a frame.
import { t } from '../i18n/dictionary.js';
import {
  frameHeight,
  hitRowToIndex,
  Key,
} from '../screen/screen.js';
import { persistSettings, renderFrame } from './fullscreen_views.js';

function langIndex(lang) {
  return lang === 'es' ? 1 : 0;
}

function currentMark(active, lang) {
  return active ? t(lang, 'lang.current') : '';
}

function setLanguage(ctx, lang) {
  ctx.lang = lang;
  ctx.settings.lang = lang;
  if (!persistSettings(ctx)) {
    return 'abort';
  }
  return 'language';
}

function languageMouse(ctx, pressed, items) {
  if (pressed.key === Key.WHEEL_UP) {
    return setLanguage(ctx, 'en');
  }
  if (pressed.key === Key.WHEEL_DOWN) {
    return setLanguage(ctx, 'es');
  }
  if (pressed.key !== Key.CLICK) {
    return null;
  }
  const idx = hitRowToIndex(items, langIndex(ctx.lang), frameHeight(ctx.rows), pressed.y);
  if (idx === 0) {
    return setLanguage(ctx, 'en');
  }
  if (idx === 1) {
    return setLanguage(ctx, 'es');
  }
  return 'language';
}

async function runLanguageView(ctx, ask) {
  const items = [
    { label: 'English', state: currentMark(ctx.lang === 'en', ctx.lang) },
    { label: 'Español', state: currentMark(ctx.lang === 'es', ctx.lang) },
  ];
  renderFrame(ctx, t(ctx.lang, 'lang.title'), items, langIndex(ctx.lang), t(ctx.lang, 'footer.back'));
  const pressed = await ask();
  const mouseNext = languageMouse(ctx, pressed, items);
  if (mouseNext) {
    return mouseNext;
  }
  if (pressed.key === Key.RUNE) {
    if (pressed.rune === '1' || pressed.rune === 'e' || pressed.rune === 'E') {
      return setLanguage(ctx, 'en');
    }
    if (pressed.rune === '2') {
      return setLanguage(ctx, 'es');
    }
    return 'language';
  }
  if (pressed.key === Key.ENTER || pressed.key === Key.SPACE) {
    return setLanguage(ctx, ctx.lang === 'es' ? 'en' : 'es');
  }
  if (pressed.key === Key.QUIT || pressed.key === Key.ESC) {
    return 'menu';
  }
  return 'language';
}

export {
  langIndex,
  runLanguageView,
};
