#!/usr/bin/env node
'use strict';

// Entry point: all logic lives in lib/cli.
const { run } = require('../lib/cli');

process.exitCode = run(process.argv.slice(2));
