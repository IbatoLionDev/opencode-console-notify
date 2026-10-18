// Adapter: test-toast delivery. encodePowerShell is pure; sending lives in
// spawnSync and is not exercised here (platform-specific).
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { encodePowerShell } from '../../../lib/notifier/sender.js';

test('sender: encodePowerShell is UTF-16LE base64 and round-trips', () => {
  const encoded = encodePowerShell('$x = 1');
  const buf = Buffer.from(encoded, 'base64');
  assert.strictEqual(buf.toString('utf16le'), '$x = 1');
  assert.strictEqual(buf.length, '$x = 1'.length * 2);
});

test('sender: encodePowerShell handles non-ASCII', () => {
  const encoded = encodePowerShell('é');
  assert.strictEqual(Buffer.from(encoded, 'base64').toString('utf16le'), 'é');
});
