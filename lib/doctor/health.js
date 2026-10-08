// Domain + orchestration: health rules over the install end state
// (plugin file + identity + notifier backend) plus the doctor command.

import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import * as aumid from '../aumid/index.js';
import * as config from '../config/index.js';
import { maybePrintUpdateNotice } from './updateNotice.js';

function healthy(report) {
  return report.pluginOK && report.aumidOK && report.notifierOK;
}

function checkHealth(pluginsDir, reg = aumid.defaultRegistry) {
  const report = {
    pluginsDir,
    pluginPath: config.pluginPath(pluginsDir),
    pluginOK: false,
    pluginNote: '',
    aumidOK: false,
    aumidNote: '',
    notifierOK: false,
    notifierNote: '',
  };

  let info;
  try {
    info = fs.statSync(report.pluginPath);
  } catch (err) {
    if (err && err.code === 'ENOENT') {
      report.pluginOK = false;
      report.pluginNote = `MISSING (no file at ${report.pluginPath})`;
    } else {
      throw new Error(`stat plugin file: ${err.message}`);
    }
  }
  if (info) {
    if (info.isDirectory()) {
      report.pluginOK = false;
      report.pluginNote = `MISSING (${report.pluginPath} is a directory, not a file)`;
    } else if (info.size === 0) {
      report.pluginOK = false;
      report.pluginNote = `MISSING (${report.pluginPath} is empty)`;
    } else {
      report.pluginOK = true;
      report.pluginNote = `OK (${report.pluginPath})`;
    }
  }

  let status;
  try {
    status = reg.aumidStatus();
  } catch (err) {
    throw new Error(`check AUMID registration: ${err.message}`);
  }
  report.aumidOK = status.registered;
  if (status.registered) {
    report.aumidNote = `OK (${status.detail})`;
  } else {
    report.aumidNote = `MISSING (${status.detail})`;
  }

  const notifier = checkNotifier();
  report.notifierOK = notifier.ok;
  report.notifierNote = notifier.note;

  return report;
}

// checkNotifier verifies the toast backend binary for this platform:
// powershell.exe on Windows, notify-send on Linux. Anything else is an
// unsupported platform, reported rather than thrown.
function checkNotifier(platform = process.platform) {
  let binary = '';
  if (platform === 'win32') {
    binary = 'powershell.exe';
  } else if (platform === 'linux') {
    binary = 'notify-send';
  } else {
    return { ok: false, note: `MISSING (notifications not supported on ${platform})` };
  }
  const probe = spawnSync(binary, binary === 'notify-send' ? ['--version'] : ['-NoProfile', '-Command', 'exit 0'], { encoding: 'utf8', shell: false });
  if (probe.error || probe.status !== 0) {
    const hint = platform === 'linux' ? '; install libnotify-bin' : '';
    return { ok: false, note: `MISSING (${binary} not found on PATH${hint})` };
  }
  const where = probe.stdout ? ` (${probe.stdout.trim().split('\n')[0]})` : '';
  return { ok: true, note: `OK (${binary}${where})` };
}

async function doctor(pluginsDir, reg = aumid.defaultRegistry, out = process.stdout) {
  let report;
  try {
    report = checkHealth(pluginsDir, reg);
  } catch (err) {
    out.write(`Doctor failed: ${err.message}\n`);
    return 1;
  }

  out.write(`Plugin file: ${report.pluginNote}\n`);
  out.write(`${process.platform === 'win32' ? 'AUMID registration' : 'Identity'}: ${report.aumidNote}\n`);
  out.write(`Notifier: ${report.notifierNote}\n`);
  if (healthy(report)) {
    out.write(process.platform === 'win32'
      ? 'Healthy: plugin installed and AUMID registered.\n'
      : 'Healthy: plugin installed and notifier ready.\n');
    await maybePrintUpdateNotice(out);
    return 0;
  }
  out.write('Issues found: run "opencode-notify install" to fix.\n');
  await maybePrintUpdateNotice(out);
  return 1;
}

export {
  checkHealth,
  checkNotifier,
  healthy,
  doctor,
};
