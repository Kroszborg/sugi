#!/usr/bin/env node
// npm binary wrapper for sugi
// Downloads the correct pre-built binary from GitHub Releases on postinstall.
// Falls back to `go install` if no release binary exists yet.

'use strict';

const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { spawnSync } = require('child_process');

const VERSION = require('./package.json').version;
const REPO = 'Kroszborg/sugi';
const BIN_DIR = path.join(__dirname, 'bin');

function platformInfo() {
  const platform = process.platform;
  const arch = process.arch;

  const osMap = { darwin: 'macOS', linux: 'Linux', win32: 'Windows' };
  const archMap = { x64: 'x86_64', arm64: 'arm64' };

  const osName = osMap[platform];
  const cpu = archMap[arch];

  if (!osName || !cpu) {
    throw new Error(`Unsupported platform: ${platform}/${arch}`);
  }

  const ext = platform === 'win32' ? '.zip' : '.tar.gz';
  const binName = platform === 'win32' ? 'sugi.exe' : 'sugi';

  return { osName, cpu, ext, binName };
}

// Download url to dest, following up to 10 redirects.
// Only opens the WriteStream once we have a confirmed 200 response.
function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    let redirects = 0;

    const get = (u) => {
      if (redirects > 10) return reject(new Error('Too many redirects'));
      https.get(u, { headers: { 'User-Agent': 'sugi-npm-installer' } }, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302 ||
            res.statusCode === 307 || res.statusCode === 308) {
          redirects++;
          res.resume(); // drain so socket can be reused
          return get(res.headers.location);
        }
        if (res.statusCode !== 200) {
          res.resume();
          return reject(new Error(`HTTP ${res.statusCode} downloading ${u}`));
        }
        // Only create the file once we have a real 200 response
        const file = fs.createWriteStream(dest);
        res.pipe(file);
        file.on('finish', () => file.close(resolve));
        file.on('error', (e) => { fs.unlink(dest, () => {}); reject(e); });
        res.on('error', (e) => { file.close(); fs.unlink(dest, () => {}); reject(e); });
      }).on('error', reject);
    };

    get(url);
  });
}

function tryGoBuild() {
  const goCheck = spawnSync('go', ['version'], { stdio: 'pipe' });
  if (goCheck.status !== 0) return false;

  if (!fs.existsSync(BIN_DIR)) fs.mkdirSync(BIN_DIR, { recursive: true });

  const binName = process.platform === 'win32' ? 'sugi.exe' : 'sugi';
  const binDest = path.join(BIN_DIR, binName);

  console.log('sugi: building from source via `go install`...');
  const result = spawnSync(
    'go',
    ['install', `github.com/${REPO}/cmd/sugi@v${VERSION}`],
    { stdio: 'inherit', env: process.env }
  );

  if (result.status !== 0) return false;

  // Locate the installed binary in GOBIN or GOPATH/bin
  const goEnv = spawnSync('go', ['env', 'GOBIN', 'GOPATH'], { stdio: 'pipe' });
  if (goEnv.status !== 0) return false;

  const [gobin, gopath] = goEnv.stdout.toString().trim().split('\n');
  const candidates = [
    gobin && path.join(gobin.trim(), binName),
    gopath && path.join(gopath.trim(), 'bin', binName),
  ].filter(Boolean);

  for (const src of candidates) {
    if (fs.existsSync(src)) {
      fs.copyFileSync(src, binDest);
      fs.chmodSync(binDest, 0o755);
      console.log(`sugi: installed to ${binDest}`);
      return true;
    }
  }

  return false;
}

function extractArchive(archivePath, ext, binName) {
  if (!fs.existsSync(BIN_DIR)) fs.mkdirSync(BIN_DIR, { recursive: true });

  if (ext === '.tar.gz') {
    // Use array args to avoid shell injection
    const result = spawnSync('tar', ['-xzf', archivePath, '-C', BIN_DIR, binName], { stdio: 'inherit' });
    if (result.status !== 0) throw new Error('tar extraction failed');
  } else {
    // Windows: use PowerShell Expand-Archive
    const ps = `Expand-Archive -Path '${archivePath}' -DestinationPath '${BIN_DIR}' -Force`;
    const result = spawnSync('powershell', ['-NoProfile', '-Command', ps], { stdio: 'inherit' });
    if (result.status !== 0) throw new Error('Expand-Archive failed');
  }

  const binDest = path.join(BIN_DIR, binName);
  if (!fs.existsSync(binDest)) throw new Error(`Binary not found after extraction: ${binDest}`);
  fs.chmodSync(binDest, 0o755);
  return binDest;
}

async function main() {
  const { osName, cpu, ext, binName } = platformInfo();

  const assetName = `sugi_${VERSION}_${osName}_${cpu}${ext}`;
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${assetName}`;
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'sugi-'));
  const archivePath = path.join(tmpDir, assetName);

  console.log(`sugi: downloading ${assetName}...`);

  let downloadOk = false;
  try {
    await downloadFile(url, archivePath);
    downloadOk = true;
  } catch (e) {
    console.error(`sugi: download failed - ${e.message}`);
  }

  if (!downloadOk) {
    console.log('sugi: falling back to go install...');
    const ok = tryGoBuild();
    fs.rmSync(tmpDir, { recursive: true, force: true });
    if (!ok) {
      console.error('sugi: install failed. Install manually:');
      console.error(`  go install github.com/${REPO}/cmd/sugi@latest`);
      console.error(`  or download from: https://github.com/${REPO}/releases`);
      process.exit(1);
    }
    return;
  }

  try {
    const binDest = extractArchive(archivePath, ext, binName);
    console.log(`sugi: installed to ${binDest}`);
  } catch (e) {
    console.error(`sugi: extraction failed - ${e.message}`);
    process.exit(1);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

main().catch((e) => {
  console.error(e.message);
  process.exit(1);
});
