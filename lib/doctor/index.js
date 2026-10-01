import fs from 'node:fs';
import * as aumid from '../aumid/index.js';
import * as config from '../config/index.js';

function healthy(report) {
  return report.pluginOK && report.aumidOK;
}

function checkHealth(pluginsDir, reg = aumid.defaultRegistry) {
  const report = {
    pluginsDir,
    pluginPath: config.pluginPath(pluginsDir),
    pluginOK: false,
    pluginNote: '',
    aumidOK: false,
    aumidNote: '',
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

  return report;
}

function doctor(pluginsDir, reg = aumid.defaultRegistry, out = process.stdout) {
  let report;
  try {
    report = checkHealth(pluginsDir, reg);
  } catch (err) {
    out.write(`Doctor failed: ${err.message}\n`);
    return 1;
  }

  out.write(`Plugin file: ${report.pluginNote}\n`);
  out.write(`AUMID registration: ${report.aumidNote}\n`);
  if (healthy(report)) {
    out.write('Healthy: plugin installed and AUMID registered.\n');
    return 0;
  }
  out.write('Issues found: run "opencode-notify install" to fix.\n');
  return 1;
}

export {
  checkHealth,
  healthy,
  doctor,
};
