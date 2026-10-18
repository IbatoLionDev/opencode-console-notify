// Domain: toast payloads and notifier argument builders. Pure, no I/O.
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import {
  buildToastScript,
  escapeToastXml,
  buildNotifySendArgs,
  NOTIFY_SEND_APP,
} from '../../../lib/notifier/toastScript.js';
import { AUMID } from '../../../lib/aumid/constants.js';

test('toastScript: NOTIFY_SEND_APP', () => {
  assert.strictEqual(NOTIFY_SEND_APP, 'OpenCode');
});

test('toastScript: escapeToastXml escapes the five XML entities', () => {
  assert.strictEqual(escapeToastXml('a&b<c>d"e\'f'), 'a&amp;b&lt;c&gt;d&quot;e&apos;f');
  assert.strictEqual(escapeToastXml('plain text'), 'plain text');
  assert.strictEqual(escapeToastXml(''), '');
});

test('toastScript: buildToastScript contains escaped title/line, AUMID and tags', () => {
  const script = buildToastScript('Title & <Tag>', 'Line "quote"');
  assert.ok(script.includes('&amp;'), 'title ampersand escaped');
  assert.ok(script.includes('&lt;Tag&gt;'), 'title tags escaped');
  assert.ok(script.includes('&quot;quote&quot;'), 'line quotes escaped');
  assert.ok(script.includes(AUMID), 'AUMID referenced');
  assert.ok(script.includes("'opencode-test'"), 'tag set');
  assert.ok(script.includes("'opencode'"), 'group set');
  assert.ok(script.includes('Show($t)'), 'toast shown');
});

test('toastScript: buildNotifySendArgs exact argv, never through a shell', () => {
  const args = buildNotifySendArgs('OpenCode', 'hello world');
  assert.deepStrictEqual(args, [
    '--app-name=OpenCode',
    '--urgency=normal',
    '--expire-time=8000',
    'OpenCode',
    'hello world',
  ]);
});
