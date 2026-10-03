// Line-mode renders: pure string builders for the config views.
// Interaction (loop) lives in tui.js. Mirrors internal/tui/tui_renders.go.
import { t } from '../i18n/dictionary.js';

const NPM_PAGE = 'https://www.npmjs.com/package/opencode-console-notify';
const REPO_URL = 'https://github.com/IbatoLionDev/opencode-console-notify';
const RELEASES_URL = 'https://github.com/IbatoLionDev/opencode-console-notify/releases';

function alertItems(settings, lang) {
  return [
    { key: '1', event: 'session.idle', name: t(lang, 'alert.sessionIdle'), trigger: t(lang, 'trigger.sessionIdle'), enabled: settings.alerts.sessionIdle !== false, field: 'sessionIdle' },
    { key: '2', event: 'session.error', name: t(lang, 'alert.sessionError'), trigger: t(lang, 'trigger.sessionError'), enabled: settings.alerts.sessionError !== false, field: 'sessionError' },
    { key: '3', event: 'permission.asked', name: t(lang, 'alert.permissionAsked'), trigger: t(lang, 'trigger.permissionAsked'), enabled: settings.alerts.permissionAsked !== false, field: 'permissionAsked' },
    { key: '4', event: 'question.asked', name: t(lang, 'alert.questionAsked'), trigger: t(lang, 'trigger.questionAsked'), enabled: settings.alerts.questionAsked !== false, field: 'questionAsked' },
  ];
}

function onOff(lang, on) {
  return t(lang, on ? 'state.on' : 'state.off');
}

function renderMenu(lang) {
  return [
    t(lang, 'app.title'),
    `1. ${t(lang, 'menu.language')} (L)`,
    `2. ${t(lang, 'menu.update')} (U)`,
    `3. ${t(lang, 'menu.info')} (I)`,
    `4. ${t(lang, 'menu.alerts')} (A)`,
    `5. ${t(lang, 'menu.exit')} (Q)`,
    t(lang, 'footer.menu'),
    '',
  ].join('\n');
}

function renderLanguage(lang) {
  const markEn = lang === 'en' ? ` [${t(lang, 'lang.current')}]` : '';
  const markEs = lang === 'es' ? ` [${t(lang, 'lang.current')}]` : '';
  return [
    t(lang, 'lang.title'),
    `1. English${markEn}`,
    `2. Español${markEs}`,
    `[1/2] ${t(lang, 'lang.current')}: ${lang} | ${t(lang, 'footer.back')}`,
    '',
  ].join('\n');
}

function renderAlerts(settings, lang) {
  const lines = [t(lang, 'alerts.title')];
  for (const a of alertItems(settings, lang)) {
    lines.push(`${a.key}. ${a.name} [${onOff(lang, a.enabled)}]`);
    lines.push(`   ${a.trigger}`);
  }
  lines.push(t(lang, 'footer.toggle'));
  lines.push('');
  return lines.join('\n');
}

function renderInfo(lang) {
  return [
    t(lang, 'menu.info'),
    'install, uninstall, doctor, test, upgrade, version, config',
    `npm: ${NPM_PAGE}`,
    `repo: ${REPO_URL}`,
    `releases: ${RELEASES_URL}`,
    t(lang, 'footer.back'),
    '',
  ].join('\n');
}

export {
  NPM_PAGE,
  REPO_URL,
  RELEASES_URL,
  alertItems,
  renderMenu,
  renderLanguage,
  renderAlerts,
  renderInfo,
};
