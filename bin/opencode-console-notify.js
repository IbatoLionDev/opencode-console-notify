#!/usr/bin/env node

// Entry point: all logic lives in lib/cli.
import { run } from '../lib/cli/index.js';

process.exitCode = run(process.argv.slice(2));
