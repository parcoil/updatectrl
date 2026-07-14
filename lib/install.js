"use strict";

const { execSync } = require("child_process");
const fs = require("fs");
const https = require("https");
const http = require("http");
const path = require("path");
const zlib = require("zlib");
const { createGunzip } = zlib;

const PACKAGE_NAME = "updatectrl";
const REPO = "parcoil/updatectrl";
const VERSION = require("../package.json").version;

const PLATFORM_MAP = {
  win32: "windows",
  darwin: "darwin",
  linux: "linux",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

function getPlatformInfo() {
  const platform = process.platform;
  const arch = process.arch;

  const osName = PLATFORM_MAP[platform];
  const archName = ARCH_MAP[arch];

  if (!osName) {
    throw new Error(`Unsupported platform: ${platform}`);
  }
  if (!archName) {
    throw new Error(`Unsupported architecture: ${arch}`);
  }

  return { os: osName, arch: archName };
}

function getBinaryName(os, arch) {
  const base = `updatectrl-${os}-${arch}`;
  return os === "windows" ? `${base}.exe` : base;
}

function getCacheDir() {
  const cacheDir = path.join(__dirname, "..", "lib", "bin");
  if (!fs.existsSync(cacheDir)) {
    fs.mkdirSync(cacheDir, { recursive: true });
  }
  return cacheDir;
}

function getBinaryPath(os, arch) {
  const name = getBinaryName(os, arch);
  return path.join(getCacheDir(), name);
}

function getLocalVersion(binaryPath) {
  try {
    const output = execSync(`"${binaryPath}" --version`, {
      encoding: "utf-8",
      timeout: 5000,
    }).trim();
    // Output is like "updatectrl version 0.1.0"
    const match = output.match(/version\s+([\d.]+)/);
    return match ? match[1] : null;
  } catch {
    return null;
  }
}

function download(url) {
  return new Promise((resolve, reject) => {
    const mod = url.startsWith("https") ? https : http;
    mod
      .get(url, { headers: { "User-Agent": `${PACKAGE_NAME}-installer` } }, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return download(res.headers.location).then(resolve, reject);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} downloading ${url}`));
        }
        const chunks = [];
        res.on("data", (chunk) => chunks.push(chunk));
        res.on("end", () => resolve(Buffer.concat(chunks)));
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

function decompressGzip(buffer) {
  return new Promise((resolve, reject) => {
    zlib.gunzip(buffer, (err, result) => {
      if (err) reject(err);
      else resolve(result);
    });
  });
}

async function downloadBinary() {
  const { os, arch } = getPlatformInfo();
  const binaryName = getBinaryName(os, arch);
  const binaryPath = getBinaryPath(os, arch);

  // Skip download if binary already exists (local dev or already installed)
  if (fs.existsSync(binaryPath)) {
    return;
  }

  console.log(`Downloading ${PACKAGE_NAME} v${VERSION} for ${os}/${arch}...`);

  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryName}`;
  let data;
  try {
    data = await download(url);
  } catch (err) {
    throw new Error(
      `Failed to download ${binaryName} from ${url}\n` +
        `  You may need to build the binary manually or check if v${VERSION} exists on GitHub releases.\n` +
        `  Error: ${err.message}`
    );
  }

  // Decompress if gzipped (Linux and macOS binaries are distributed as .tar.gz)
  if (os !== "windows") {
    try {
      data = await decompressGzip(data);
    } catch {
      // Not gzipped, use raw data
    }
  }

  fs.writeFileSync(binaryPath, data);

  if (os !== "windows") {
    fs.chmodSync(binaryPath, 0o755);
  }

  console.log(`${PACKAGE_NAME} v${VERSION} installed successfully.`);
}

async function cleanup() {
  const { os, arch } = getPlatformInfo();
  const binaryPath = getBinaryPath(os, arch);
  try {
    if (fs.existsSync(binaryPath)) {
      fs.unlinkSync(binaryPath);
    }
  } catch {
    // Ignore cleanup errors
  }
}

module.exports = { downloadBinary, cleanup };

if (require.main === module) {
  downloadBinary().catch((err) => {
    console.error(`\n${PACKAGE_NAME} installation failed:\n  ${err.message}\n`);
    process.exit(1);
  });
}
