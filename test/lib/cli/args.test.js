// Domain: CLI argument parsing. Pure, no I/O.
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { parseArgs } from '../../../lib/cli/args.js';

test('args: install command parses with empty pluginsDir and rest', () => {
  const r = parseArgs(['install']);
  assert.strictEqual(r.command, 'install');
  assert.strictEqual(r.pluginsDir, '');
  assert.deepStrictEqual(r.rest, []);
});

test('args: command aliases --help/-h --version/-V', () => {
  assert.strictEqual(parseArgs(['--help']).command, 'help');
  assert.strictEqual(parseArgs(['-h']).command, 'help');
  assert.strictEqual(parseArgs(['--version']).command, 'version');
  assert.strictEqual(parseArgs(['-V']).command, 'version');
});

test('args: every command is accepted', () => {
  for (const cmd of ['install', 'uninstall', 'doctor', 'test', 'upgrade', 'version', 'help', 'config']) {
    assert.strictEqual(parseArgs([cmd]).command, cmd, cmd);
  }
});

test('args: config keeps the rest args untouched', () => {
  const r = parseArgs(['config', '--lang', 'es', '--toggle', 'sessionIdle=off']);
  assert.strictEqual(r.command, 'config');
  assert.deepStrictEqual(r.rest, ['--lang', 'es', '--toggle', 'sessionIdle=off']);
});

test('args: --plugins-dir and -plugins-dir, long and equals forms (siempre con comando)', () => {
  assert.strictEqual(parseArgs(['install', '--plugins-dir', 'D:\\plugins']).pluginsDir, 'D:\\plugins');
  assert.strictEqual(parseArgs(['install', '-plugins-dir', 'D:\\plugins']).pluginsDir, 'D:\\plugins');
  assert.strictEqual(parseArgs(['install', '--plugins-dir=D:\\plugins']).pluginsDir, 'D:\\plugins');
  assert.strictEqual(parseArgs(['doctor', '--plugins-dir', 'D:\\plugins']).pluginsDir, 'D:\\plugins');
  // sin comando no hay parseo válido: lanza la señal de uso
  assert.throws(() => parseArgs(['--plugins-dir', 'D:\\plugins']));
});

test('args: rest solo viaja en config; los demás comandos la traen vacía', () => {
  assert.deepStrictEqual(parseArgs(['install']).rest, []);
  assert.deepStrictEqual(parseArgs(['config', '--lang', 'es']).rest, ['--lang', 'es']);
});

test('args: plugins-dir consumed before the command does not become the command', () => {
  const r = parseArgs(['--plugins-dir', 'D:\\plugins', 'doctor']);
  assert.strictEqual(r.command, 'doctor');
  assert.strictEqual(r.pluginsDir, 'D:\\plugins');
});

test('args: empty args throws the usage signal', () => {
  assert.throws(() => parseArgs([]));
});

test('args: unknown flag throws', () => {
  assert.throws(() => parseArgs(['--nope']), /unknown flag/);
});

test('args: unknown command throws', () => {
  assert.throws(() => parseArgs(['frobnicate']), /unknown command/);
});

test('args: flag that needs a value throws', () => {
  assert.throws(() => parseArgs(['--plugins-dir']), /needs a value/);
});
