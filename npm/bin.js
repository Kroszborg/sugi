#!/usr/bin/env node
'use strict';

const path = require('path');
const { spawnSync } = require('child_process');
const fs = require('fs');

const binName = process.platform === 'win32' ? 'sugi.exe' : 'sugi';
const binPath = path.join(__dirname, 'bin', binName);

if (!fs.existsSync(binPath)) {
  console.error('sugi: binary not found at ' + binPath);
  console.error('Try reinstalling: npm install -g sugi');
  process.exit(1);
}

const result = spawnSync(binPath, process.argv.slice(2), {
  stdio: 'inherit',
  env: process.env,
});

if (result.error) {
  console.error('sugi: failed to start:', result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 0);
