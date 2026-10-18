// Domain: custom alerts. Pure, no I/O.
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import {
  CUSTOM_TEXT_LIMIT,
  truncateCustomText,
  validateCustom,
  normalizeCustoms,
  nextCustomId,
  findCustom,
  addCustom,
  updateCustom,
  removeCustom,
  toggleCustom,
  customsForEvent,
} from '../../../lib/config/customs.js';

function settingsWith(customs) {
  return { version: 2, lang: 'en', alerts: {}, customAlerts: customs || [] };
}

test('customs: CUSTOM_TEXT_LIMIT is 160', () => {
  assert.strictEqual(CUSTOM_TEXT_LIMIT, 160);
});

test('customs: truncateCustomText truncates at 160 chars', () => {
  assert.strictEqual(truncateCustomText('a'.repeat(200)).length, 160);
  assert.strictEqual(truncateCustomText('short'), 'short');
  assert.strictEqual(truncateCustomText(null), '');
  assert.strictEqual(truncateCustomText(undefined), '');
});

test('customs: validateCustom throws unknown event and empty title', () => {
  assert.throws(() => validateCustom('no.such.event', 't'), /unknown event/);
  assert.throws(() => validateCustom('session.execution.succeeded', ''), /must not be empty/);
  assert.throws(() => validateCustom('session.execution.succeeded', '   '), /must not be empty/);
  assert.doesNotThrow(() => validateCustom('session.execution.succeeded', 'ok'));
});

test('customs: normalizeCustoms drops invalid entries and normalizes fields', () => {
  const out = normalizeCustoms([
    { id: 'custom-1', event: 'session.execution.succeeded', title: 'T', body: 'B', enabled: true },
    { event: 'bad.event', title: 'x' },             // unknown event -> dropped
    { id: 'custom-2', event: 'permission.asked' },    // empty title -> dropped
    null,                                              // not an object -> dropped
    { id: 'custom-3', event: 'form.created', title: 'Only title' }, // enabled defaults true
  ]);
  assert.deepStrictEqual(out.map((c) => c.id), ['custom-1', 'custom-3']);
  assert.strictEqual(out[1].enabled, true);
  assert.strictEqual(out[1].body, '');
});

test('customs: addCustom generates sequential ids and truncates title/body', () => {
  const s = settingsWith([]);
  const id1 = addCustom(s, 'session.execution.succeeded', 'T'.repeat(200), 'B'.repeat(200));
  assert.strictEqual(id1, 'custom-1');
  assert.strictEqual(s.customAlerts[0].title.length, 160);
  assert.strictEqual(s.customAlerts[0].body.length, 160);
  assert.strictEqual(s.customAlerts[0].enabled, true);
  const id2 = addCustom(s, 'permission.asked', 'second', '');
  assert.strictEqual(id2, 'custom-2');
});

test('customs: addCustom validates and leaves the list untouched on failure', () => {
  const s = settingsWith([]);
  assert.throws(() => addCustom(s, 'nope.event', 't'), /unknown event/);
  assert.strictEqual(s.customAlerts.length, 0);
});

test('customs: nextCustomId after manual entries', () => {
  const s = settingsWith([{ id: 'custom-5', event: 'session.execution.succeeded', title: 't', body: '', enabled: true }]);
  assert.strictEqual(nextCustomId(s), 'custom-6');
  assert.strictEqual(nextCustomId(settingsWith([])), 'custom-1');
});

test('customs: findCustom returns undefined when missing', () => {
  assert.strictEqual(findCustom(settingsWith([]), 'custom-9'), undefined);
  const s = settingsWith([{ id: 'custom-1', event: 'session.execution.succeeded', title: 't', body: '', enabled: true }]);
  assert.strictEqual(findCustom(s, 'custom-1').title, 't');
});

test('customs: updateCustom returns false for missing id, true and mutates otherwise', () => {
  const s = settingsWith([{ id: 'custom-1', event: 'session.execution.succeeded', title: 'old', body: 'b', enabled: true }]);
  assert.strictEqual(updateCustom(s, 'custom-9', 'permission.asked', 'x', 'y'), false);
  assert.strictEqual(updateCustom(s, 'custom-1', 'permission.asked', 'new title', 'new body'), true);
  const c = findCustom(s, 'custom-1');
  assert.strictEqual(c.event, 'permission.asked');
  assert.strictEqual(c.title, 'new title');
  assert.strictEqual(c.body, 'new body');
});

test('customs: removeCustom returns false when missing, true and removes when present', () => {
  const s = settingsWith([
    { id: 'custom-1', event: 'session.execution.succeeded', title: 'a', body: '', enabled: true },
    { id: 'custom-2', event: 'permission.asked', title: 'b', body: '', enabled: true },
  ]);
  assert.strictEqual(removeCustom(s, 'custom-9'), false);
  assert.strictEqual(s.customAlerts.length, 2);
  assert.strictEqual(removeCustom(s, 'custom-1'), true);
  assert.strictEqual(s.customAlerts.length, 1);
  assert.strictEqual(s.customAlerts[0].id, 'custom-2');
});

test('customs: toggleCustom flips enabled and returns false when missing', () => {
  const s = settingsWith([{ id: 'custom-1', event: 'session.execution.succeeded', title: 'a', body: '', enabled: true }]);
  assert.strictEqual(toggleCustom(s, 'custom-9'), false);
  assert.strictEqual(toggleCustom(s, 'custom-1'), true);
  assert.strictEqual(findCustom(s, 'custom-1').enabled, false);
  assert.strictEqual(toggleCustom(s, 'custom-1'), true);
  assert.strictEqual(findCustom(s, 'custom-1').enabled, true);
});

test('customs: customsForEvent filters enabled and matching event', () => {
  const s = settingsWith([
    { id: 'custom-1', event: 'session.execution.succeeded', title: 'a', body: '', enabled: true },
    { id: 'custom-2', event: 'session.execution.succeeded', title: 'b', body: '', enabled: false },
    { id: 'custom-3', event: 'permission.asked', title: 'c', body: '', enabled: true },
  ]);
  const hits = customsForEvent(s, 'session.execution.succeeded');
  assert.strictEqual(hits.length, 1);
  assert.strictEqual(hits[0].id, 'custom-1');
});
