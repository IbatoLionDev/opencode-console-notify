// Domain: AUMID identity constants. Pure, no I/O.
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import {
  AUMID,
  AUMID_DISPLAY_NAME,
  AUMID_KEY_PATH,
  AUMID_PS_PATH,
} from '../../../lib/aumid/constants.js';

test('aumid constants: identity values', () => {
  assert.strictEqual(AUMID, 'OpenCode.Notifier');
  assert.strictEqual(AUMID_DISPLAY_NAME, 'OpenCode');
});

test('aumid constants: key paths are Windows registry paths', () => {
  assert.ok(typeof AUMID_KEY_PATH === 'string' && AUMID_KEY_PATH.length > 0);
  assert.ok(typeof AUMID_PS_PATH === 'string' && AUMID_PS_PATH.startsWith('HKCU:\\'));
  assert.ok(AUMID_KEY_PATH.includes('AppUserModelId'));
  assert.ok(AUMID_KEY_PATH.includes(AUMID));
  assert.ok(AUMID_PS_PATH.includes('AppUserModelId'));
  assert.ok(AUMID_PS_PATH.includes(AUMID));
});
