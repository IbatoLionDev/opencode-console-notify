'use strict';

// Mirrors internal/doctor/doctor.go: the parity gate that verifies the SAME
// end state as install.ps1 (plugin file present + AUMID registered)
// regardless of which install path produced it.

const fs = require('node:fs');
const aumid = require('../aumid');
const config = require('../config');

// Healthy is true only when both halves of the end state hold.
function healthy(report) {
  return report.pluginOK && report.aumidOK;
}

// CheckHealth inspects the plugin file and the AUMID registration.
// A missing plugin file or key is reported, never thrown;
// errors are reserved for genuine failures (unreadable directory,
// registry access failure).
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

// Doctor prints the health report and returns exit code 0 when healthy,
// 1 when something is missing.
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

module.exports = {
  checkHealth,
  healthy,
  doctor,
};
