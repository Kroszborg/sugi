#!/usr/bin/env node
// npm binary wrapper for sugi
// Downloads the correct pre-built binary from GitHub Releases on postinstall.
// Falls back to `go build` if no release exists yet (local / dev install).

'use strict';

const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { execSync, spawnSync } = require('child_process');

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

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    const get = (u) => {
      https.get(u, { headers: { 'User-Agent': 'sugi-npm-installer' } }, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          file.close();
          const newFile = fs.createWriteStream(dest);
          return get(res.headers.location);
        }
        if (res.statusCode !== 200) {
          file.close();
          return reject(new Error(`HTTP ${res.statusCode} for ${u}`));
        }
        res.pipe(file);
        file.on('finish', () => { file.close(); resolve(); });
      }).on('error', (e) => { file.close(); reject(e); });
    };
    get(url);
  });
}

function tryGoBuild() {
  // Check if go is available
  const goCheck = spawnSync('go', ['version'], { stdio: 'pipe' });
  if (goCheck.status !== 0) {
    return false;
  }

  if (!fs.existsSync(BIN_DIR)) fs.mkdirSync(BIN_DIR, { recursive: true });

  const binName = process.platform === 'win32' ? 'sugi.exe' : 'sugi';
  const binDest = path.join(BIN_DIR, binName);

  // If we're inside an npm-installed package, try go install as a fallback
  console.log('sugi: building from source via `go install`…');
  const result = spawnSync(
    'go',
    ['install', `github.com/${REPO}/cmd/sugi@latest`],
    { stdio: 'inherit', env: process.env }
  );

  if (result.status === 0) {
    // Find the installed binary in GOPATH/bin or GOBIN
    const goEnv = spawnSync('go', ['env', 'GOBIN', 'GOPATH'], { stdio: 'pipe' });
    if (goEnv.status === 0) {
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
    }
  }

  return false;
}

async function main() {
  const { osName, cpu, ext, binName } = platformInfo();

  const assetName = `sugi_${VERSION}_${osName}_${cpu}${ext}`;
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${assetName}`;
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'sugi-'));
  const archivePath = path.join(tmpDir, assetName);

  console.log(`sugi: downloading ${assetName}…`);

  let downloadOk = false;
  try {
    await downloadFile(url, archivePath);
    downloadOk = true;
  } catch (e) {
    console.error(`sugi: download failed — ${e.message}`);
  }

  if (!downloadOk) {
    console.log(`sugi: falling back to go install…`);
    const ok = tryGoBuild();
    if (!ok) {
      console.error(`sugi: install failed. Install manually:`);
      console.error(`  go install github.com/${REPO}/cmd/sugi@latest`);
      console.error(`  or download from: https://github.com/${REPO}/releases`);
      process.exit(1);
    }
    fs.rmSync(tmpDir, { recursive: true, force: true });
    return;
  }

  if (!fs.existsSync(BIN_DIR)) fs.mkdirSync(BIN_DIR, { recursive: true });
  const binDest = path.join(BIN_DIR, binName);

  try {
    if (ext === '.tar.gz') {
      execSync(`tar -xzf "${archivePath}" -C "${BIN_DIR}" sugi`, { stdio: 'inherit' });
    } else {
      execSync(
        `powershell -Command "Expand-Archive -Path '${archivePath}' -DestinationPath '${BIN_DIR}' -Force"`,
        { stdio: 'inherit' }
      );
    }
    fs.chmodSync(binDest, 0o755);
    console.log(`sugi: installed to ${binDest}`);
  } catch (e) {
    console.error(`sugi: extraction failed — ${e.message}`);
    process.exit(1);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

main().catch((e) => {
  console.error(e.message);
  process.exit(1);
});
