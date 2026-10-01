#!/usr/bin/env node
'use strict';

// Thin entry point: all logic lives in lib/cli (mirrors cmd/ + internal/cli).
// Keeps the bin small so npx and global installs behave identically.
const { run } = require('../lib/cli');

process.exitCode = run(process.argv.slice(2));
