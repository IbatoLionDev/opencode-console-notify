// Domain: config TUI + default toast dictionaries. English fallback.
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { LANG_EN, LANG_ES, normalizeLang, t } from '../../../lib/i18n/dictionary.js';

test('dictionary: LANG_EN / LANG_ES', () => {
  assert.strictEqual(LANG_EN, 'en');
  assert.strictEqual(LANG_ES, 'es');
});

test('dictionary: normalizeLang maps es and falls back to en', () => {
  assert.strictEqual(normalizeLang('es'), 'es');
  assert.strictEqual(normalizeLang('en'), 'en');
  assert.strictEqual(normalizeLang('pt-BR'), 'en');
  assert.strictEqual(normalizeLang(undefined), 'en');
});

test('dictionary: t returns english by default', () => {
  assert.strictEqual(t('en', 'app.title'), 'OpenCode Notify Config');
  assert.strictEqual(t('en', 'alert.sessionIdle'), 'Task finished');
});

test('dictionary: t returns spanish when requested', () => {
  assert.strictEqual(t('es', 'app.title'), 'Configuración de OpenCode Notify');
  assert.strictEqual(t('es', 'alert.sessionIdle'), 'Tarea terminada');
});

test('dictionary: t falls back to english then to the key itself', () => {
  assert.strictEqual(t('es', 'nonexistent.key'), 'nonexistent.key');
  assert.strictEqual(t('en', 'nonexistent.key'), 'nonexistent.key');
});
