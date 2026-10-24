// Domain: plugin-file health rules, including the 4.0.2 content check
// (installed bytes compared against the packaged copy). Hermetic: a fake
// AUMID registry and an injected notifier probe keep the tests
// platform-independent, and scratch state lives outside the repo.
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pluginPath } from '../../../lib/config/index.js';
import { pluginSourcePath } from '../../../lib/installer/source.js';
import { checkHealth, healthy, doctor } from '../../../lib/doctor/health.js';

// fakeRegistry stands in for the AUMID registry: registered by default,
// with whatever detail the status line would carry.
function fakeRegistry(registered = true) {
  return { aumidStatus: () => ({ registered, detail: 'test registry' }) };
}

// okNotifier stands in for the platform notifier probe.
function okNotifier() {
  return () => ({ ok: true, note: 'OK (test notifier)' });
}

// capturedOut collects everything a command writes for assertions.
function capturedOut() {
  const lines = [];
  return { lines, write(s) { lines.push(s); return true; } };
}

// tempPluginsDir creates a scratch plugins directory outside the repo,
// pins optional pre-made content at the plugin path (null: a directory
// instead of a file; Buffer: file contents; undefined: no file), and
// removes the directory in teardown.
function tempPluginsDir(t, contents) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'ocn-health-'));
  t.after(() => fs.rmSync(dir, { recursive: true, force: true }));
  const dest = pluginPath(dir);
  if (contents === null) {
    fs.mkdirSync(dest, { recursive: true });
  } else if (contents !== undefined) {
    fs.writeFileSync(dest, contents);
  }
  return { dir, dest };
}

// staleStub is a V1-shaped plugin export the current loader rejects;
// it is non-empty and differs from the packaged copy.
const staleStub = Buffer.from(
  'export default async ({ client, $, directory }) => { /* stale bytes */ }\n',
  'utf8',
);

test('doctor health: EMPTY-style fixtures differ from the packaged copy', (t) => {
  // Self-check guarding the stale fixture used by later tests: the
  // packaged copy exists and is never equal to the stub above.
  const source = pluginSourcePath();
  assert.ok(fs.existsSync(source), 'packaged plugin copy must exist');
  assert.ok(!staleStub.equals(fs.readFileSync(source)), 'stub must differ from packaged copy');
});

test('doctor health: installed bytes matching the packaged copy keep OK with pluginOK true', (t) => {
  const { dir, dest } = tempPluginsDir(t, fs.readFileSync(pluginSourcePath()));
  const report = checkHealth(dir, fakeRegistry(), okNotifier());
  assert.strictEqual(report.pluginOK, true);
  assert.strictEqual(report.pluginNote, `OK (${dest})`);
  assert.strictEqual(healthy(report), true);
});

test('doctor health: differing bytes report STALE with pluginOK false so healthy() fails', (t) => {
  const { dir, dest } = tempPluginsDir(t, staleStub);
  const report = checkHealth(dir, fakeRegistry(), okNotifier());
  assert.strictEqual(report.pluginOK, false);
  assert.strictEqual(healthy(report), false);
  assert.strictEqual(
    report.pluginNote,
    `STALE (${dest} differs from this release; run "opencode-console-notify install" to rewrite it)`,
  );
});

test('doctor health: MISSING keeps its wording and wins over the content check', (t) => {
  const noFile = tempPluginsDir(t, undefined);
  const directory = tempPluginsDir(t, null);
  const empty = tempPluginsDir(t, Buffer.alloc(0));

  const noFileReport = checkHealth(noFile.dir, fakeRegistry(), okNotifier());
  assert.strictEqual(noFileReport.pluginOK, false);
  assert.strictEqual(noFileReport.pluginNote, `MISSING (no file at ${noFile.dest})`);

  const directoryReport = checkHealth(directory.dir, fakeRegistry(), okNotifier());
  assert.strictEqual(directoryReport.pluginOK, false);
  assert.strictEqual(directoryReport.pluginNote, `MISSING (${directory.dest} is a directory, not a file)`);

  const emptyReport = checkHealth(empty.dir, fakeRegistry(), okNotifier());
  assert.strictEqual(emptyReport.pluginOK, false);
  assert.strictEqual(emptyReport.pluginNote, `MISSING (${empty.dest} is empty)`);
  // No empty or absent state may ever surface the new STALE wording.
  for (const note of [noFileReport.pluginNote, directoryReport.pluginNote, emptyReport.pluginNote]) {
    assert.ok(!note.includes('differs from this release'), `unexpected content note: ${note}`);
  }
});

test('doctor command: stale bytes print the STALE line, the Issues found line, and exit 1', async (t) => {
  const { dir, dest } = tempPluginsDir(t, staleStub);
  const out = capturedOut();
  const code = await doctor(dir, fakeRegistry(), out);
  assert.strictEqual(code, 1);
  const output = out.lines.join('');
  assert.ok(
    output.includes(
      `Plugin file: STALE (${dest} differs from this release; run "opencode-console-notify install" to rewrite it)\n`,
    ),
    `expected STALE line, got: ${JSON.stringify(output)}`,
  );
  assert.ok(
    output.includes('Issues found: run "opencode-notify install" to fix.\n'),
    `expected Issues found line, got: ${JSON.stringify(output)}`,
  );
  assert.ok(!output.includes('Healthy:'), `unexpected healthy line: ${JSON.stringify(output)}`);
});
