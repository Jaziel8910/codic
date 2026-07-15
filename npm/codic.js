#!/usr/bin/env node
'use strict';

const https = require('https');
const fs = require('fs');
const os = require('os');
const path = require('path');
const { execFileSync } = require('child_process');

const VERSION = 'v0.1.0';
const REPO = 'Jaziel8910/codic';

function assetFor(platform, arch) {
  if (platform === 'win32' && arch === 'x64') return 'codic-windows.exe';
  if (platform === 'linux' && arch === 'x64') return 'codic-linux-amd64';
  if (platform === 'darwin' && arch === 'x64') return 'codic-darwin-amd64';
  if (platform === 'darwin' && arch === 'arm64') return 'codic-darwin-arm64';
  return null;
}

function download(url, dest, cb) {
  https
    .get(url, (res) => {
      const code = res.statusCode;
      if (code === 301 || code === 302 || code === 307) {
        const loc = res.headers.location;
        if (!loc) return cb(new Error('redirect with no location header'));
        res.resume();
        return download(loc, dest, cb);
      }
      if (code !== 200) {
        res.resume();
        return cb(new Error('download failed with HTTP ' + code));
      }
      const out = fs.createWriteStream(dest);
      res.pipe(out);
      out.on('finish', () => out.close(() => cb(null)));
      out.on('error', cb);
    })
    .on('error', cb);
}

function run(binPath) {
  try {
    const res = execFileSync(binPath, process.argv.slice(2), { stdio: 'inherit' });
    process.exit(res && res.status ? res.status : 0);
  } catch (e) {
    if (typeof e.status === 'number') process.exit(e.status);
    if (e.code === 'ENOENT') {
      console.error('codic: binary not found at ' + binPath);
      process.exit(1);
    }
    process.exit(1);
  }
}

function main() {
  const platform = process.platform;
  const arch = process.arch;
  const asset = assetFor(platform, arch);

  if (!asset) {
    console.error(
      'codic: no prebuilt binary for ' + platform + '/' + arch + '.\n' +
      'Build from source: https://github.com/' + REPO
    );
    process.exit(1);
  }

  const binDir = path.join(os.homedir(), '.codic', 'bin');
  fs.mkdirSync(binDir, { recursive: true });
  const binPath = path.join(binDir, asset);

  if (fs.existsSync(binPath)) {
    return run(binPath);
  }

  const url =
    'https://github.com/' + REPO + '/releases/download/' + VERSION + '/' + asset;
  process.stderr.write('codic: downloading ' + asset + ' ...\n');

  download(url, binPath, (err) => {
    if (err) {
      console.error('codic: download failed: ' + err.message);
      process.exit(1);
    }
    if (platform !== 'win32') fs.chmodSync(binPath, 0o755);
    run(binPath);
  });
}

main();
