// Domain: the upgrade fast-path decision (4.0.2 stale-plugin self-heal).
// Pure: the decision helper takes resolved inputs only — the registry
// version, the local version, and whether the installed plugin bytes
// match the packaged copy — so no network, npm or filesystem sits in the
// unit test.
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { upgradeDecision, STALE_PLUGIN_REASON } from '../../../lib/upgrade/upgrader.js';

test('upgrade decision: equal bytes with latest <= local keep the Already-up-to-date fast path', () => {
  assert.deepStrictEqual(upgradeDecision('4.0.1', '4.0.1', true), { action: 'already' });
  assert.deepStrictEqual(upgradeDecision('4.0.0', '4.0.1', true), { action: 'already' });
  assert.deepStrictEqual(upgradeDecision('3.2.1', '4.0.1', true), { action: 'already' });
});

test('upgrade decision: differing bytes with latest <= local choose the reinstall repair', () => {
  assert.deepStrictEqual(
    upgradeDecision('4.0.1', '4.0.1', false),
    { action: 'repair', reason: STALE_PLUGIN_REASON },
  );
  assert.deepStrictEqual(
    upgradeDecision('3.2.0', '4.0.1', false),
    { action: 'repair', reason: STALE_PLUGIN_REASON },
  );
});

test('upgrade decision: a newer release upgrades regardless of plugin bytes', () => {
  assert.deepStrictEqual(upgradeDecision('4.0.2', '4.0.1', true), { action: 'upgrade' });
  assert.deepStrictEqual(upgradeDecision('4.0.2', '4.0.1', false), { action: 'upgrade' });
});

test('upgrade decision: the stale-repair reason stays a single self-explanatory line', () => {
  assert.strictEqual(
    STALE_PLUGIN_REASON,
    'Already up to date, but the plugin file differs from this release; reinstalling it.\n',
  );
  assert.ok(!STALE_PLUGIN_REASON.slice(0, -1).includes('\n'), 'reason must be one line');
});
