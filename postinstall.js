#!/usr/bin/env node

/**
 * postinstall.js
 * DeepScanBot CLI post-installation script.
 */

"use strict";

const fs = require("fs");
const path = require("path");

const PKG_ROOT = path.resolve(__dirname);
const DIST_DIR = path.join(PKG_ROOT, "dist");
const BIN_DIR = path.join(PKG_ROOT, "bin");
const BINARY_NAME = "deepscanbot";
const PROJECT_NAME = "deepscanbot";

/**
 * Detect current operating system.
 */
function detectOS() {
  const mapping = { darwin: "darwin", linux: "linux", win32: "windows" };
  const os = mapping[process.platform];
  if (!os) {
    throw new Error(`Unsupported operating system: ${process.platform}. Only macOS, Linux, and Windows are supported.`);
  }
  return os;
}

/**
 * Detect current CPU architecture.
 */
function detectArch() {
  const mapping = { x64: "amd64", arm64: "arm64" };
  const arch = mapping[process.arch];
  if (!arch) {
    throw new Error(`Unsupported architecture: ${process.arch}. Only amd64 and arm64 are supported.`);
  }
  return arch;
}

/**
 * Scan dist/ exactly once to locate the GoReleaser output directory dynamically.
 */
function findBinaryDirectory(os, arch) {
  if (!fs.existsSync(DIST_DIR)) {
    console.log("[deepscanbot] dist directory not found.");
    console.log("[deepscanbot] Skipping binary installation.");
    process.exit(0);
  }

  const entries = fs.readdirSync(DIST_DIR, { withFileTypes: true });
  const projLower = PROJECT_NAME.toLowerCase();
  const osLower = os.toLowerCase();
  const archLower = arch.toLowerCase();

  for (const entry of entries) {
    if (entry.isDirectory()) {
      const nameLower = entry.name.toLowerCase();
      if (nameLower.includes(projLower) && nameLower.includes(osLower) && nameLower.includes(archLower)) {
        return path.join(DIST_DIR, entry.name);
      }
    }
  }
  return null;
}

/**
 * Determine the correct file name based on the platform.
 */
function getBinaryFilename(os) {
  return os === "windows" ? `${BINARY_NAME}.exe` : BINARY_NAME;
}

/**
 * Isolate binary copying and permission application.
 */
function copyBinary(src, dest, os) {
  if (!fs.existsSync(BIN_DIR)) {
    fs.mkdirSync(BIN_DIR, { recursive: true });
  }

  fs.copyFileSync(src, dest);

  if (os !== "windows") {
    fs.chmodSync(dest, 0o755);
    console.log(`[deepscanbot] Chmod applied: Executable permissions set (chmod 755)`);
  }
}

/**
 * Main execution orchestration.
 */
function install() {
  try {
    const os = detectOS();
    const arch = detectArch();
    console.log(`[deepscanbot] Detected platform: ${os}/${arch}`);

    const binaryDir = findBinaryDirectory(os, arch);
    if (!binaryDir) {
      throw new Error(`Unsupported platform combination: Build artifacts for ${os}/${arch} not found.`);
    }

    const filename = getBinaryFilename(os);
    const srcPath = path.join(binaryDir, filename);

    if (!fs.existsSync(srcPath)) {
      throw new Error(`Missing binary: Expected executable target was not found at ${srcPath}`);
    }

    console.log(`[deepscanbot] Selected binary: ${srcPath}`);

    const destPath = path.join(BIN_DIR, filename);
    copyBinary(srcPath, destPath, os);

    console.log(`[deepscanbot] Installation complete.`);
  } catch (err) {
    console.error(`[deepscanbot] Installation failed: ${err.message}`);
    process.exit(1);
  }
}

install();