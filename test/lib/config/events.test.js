// Domain: OpenCode event catalog. Pure, no I/O.
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { knownEvents, isKnownEvent, eventDescription, isNoisyEvent } from '../../../lib/config/events.js';

test('events: knownEvents returns copies of the catalog', () => {
  const all = knownEvents();
  assert.ok(all.length > 0);
  assert.ok(all.some((e) => e.type === 'session.execution.succeeded'));
  // copies: mutating one entry must not affect the catalog
  all[0].type = 'mutated';
  assert.notStrictEqual(knownEvents()[0].type, 'mutated');
});

test('events: isKnownEvent true for catalog entries, false otherwise', () => {
  assert.strictEqual(isKnownEvent('session.execution.succeeded'), true);
  assert.strictEqual(isKnownEvent('permission.asked'), true);
  assert.strictEqual(isKnownEvent('form.created'), true);
  assert.strictEqual(isKnownEvent('no.such.event'), false);
  assert.strictEqual(isKnownEvent(''), false);
});

test('events: eventDescription returns es/en and falls back to the type', () => {
  assert.strictEqual(eventDescription('session.execution.succeeded', 'en'), 'run finished, waiting for input');
  assert.strictEqual(eventDescription('session.execution.succeeded', 'es'), 'ejecución terminada, esperando tu entrada');
  assert.strictEqual(eventDescription('no.such.event', 'en'), 'no.such.event');
  assert.strictEqual(eventDescription('no.such.event', 'es'), 'no.such.event');
});

test('events: isNoisyEvent marks noisy families only', () => {
  assert.strictEqual(isNoisyEvent('session.usage.updated'), true);
  assert.strictEqual(isNoisyEvent('session.step.streamed'), true);
  assert.strictEqual(isNoisyEvent('session.text.delta'), true);
  assert.strictEqual(isNoisyEvent('session.execution.succeeded'), false);
  assert.strictEqual(isNoisyEvent('permission.asked'), false);
  assert.strictEqual(isNoisyEvent('no.such.event'), false);
});
