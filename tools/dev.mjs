// npm run dev [-- go] -- <args>: run the repo copy like a dev run,
// against a stable sandbox that never touches the real plugins directory
// (%HOME%/.config/opencode/plugins) or any global install.
//
//   npm run dev -- config                    # JS CLI, sandboxed
//   npm run dev -- config --lang es          # args pass through
//   npm run dev:go -- config --list-alerts   # Go CLI via go run
//
// The sandbox lives in .dev-sandbox/ (gitignored) so settings persist
// between dev runs while staying isolated from the real install.
// Pass --plugins-dir explicitly to use another directory instead.
import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = path.dirname(path.dirname(fileURLToPath(import.meta.url)));

const raw = process.argv.slice(2);
let runtime = 'node';
let args = raw;
if (raw[0] === 'go') {
  runtime = 'go';
  args = raw.slice(1);
}

const sandboxPlugins = path.join(repoRoot, '.dev-sandbox', 'plugins');
fs.mkdirSync(sandboxPlugins, { recursive: true });

// Bare `npm run dev` opens the config TUI: it is the daily driver.
const withDefault = args.length === 0 ? ['config'] : args;
const hasPluginsDir = withDefault.some((a, i) => (
  a === '--plugins-dir'
  || a === '-plugins-dir'
  || a.startsWith('--plugins-dir=')
  || a.startsWith('-plugins-dir=')
  || (i > 0 && (withDefault[i - 1] === '--plugins-dir' || withDefault[i - 1] === '-plugins-dir'))
));
const finalArgs = hasPluginsDir ? withDefault : [...withDefault, '--plugins-dir', sandboxPlugins];

console.log(`[dev] runtime=${runtime} pluginsDir=${hasPluginsDir ? '(yours)' : sandboxPlugins}`);

let result;
if (runtime === 'go') {
  result = spawnSync('go', ['run', './cmd/opencode-notify', ...finalArgs], { stdio: 'inherit', cwd: repoRoot });
} else {
  result = spawnSync(process.execPath, ['bin/opencode-console-notify.js', ...finalArgs], { stdio: 'inherit', cwd: repoRoot });
}

if (result.error) {
  console.error(`[dev] failed to start: ${result.error.message}`);
  process.exit(1);
}
process.exit(result.status ?? 1);
