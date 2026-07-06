#!/usr/bin/env node

/**
 * postinstall.js
 *
 * DeepScanBot CLI post-installation script.
 *
 * Detects the user's operating system and CPU architecture, then dynamically
 * scans the dist/ directory to find the correct pre-built Go binary produced
 * by GoReleaser.
 *
 * GoReleaser output layout:
 *   dist/deepscanbot_<os>_<arch>_v<goarm>/my-cli
 *   dist/deepscanbot_windows_amd64_v1/my-cli.exe
 *
 * The version suffix (_v1, _v8.0) varies and must not be hardcoded.
 *
 * Supported platforms:
 *   - macOS   amd64 (Intel)
 *   - macOS   arm64 (Apple Silicon)
 *   - Linux   amd64
 *   - Linux   arm64
 *   - Windows amd64
 */

"use strict";

const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const PKG_ROOT = path.resolve(__dirname);
const DIST_DIR = path.join(PKG_ROOT, "dist");
const BIN_DIR = path.join(PKG_ROOT, "bin");
const BINARY_NAME = "deepscanbot";
const PROJECT_NAME = "deepscanbot";

/**
 * Map Node.js process.platform to Go OS naming convention.
 */
function detectOS() {
  const platform = process.platform;
  const mapping = {
    darwin: "darwin",
    linux: "linux",
    win32: "windows",
  };

  if (!mapping[platform]) {
    throw new Error(
      `Unsupported operating system: "${platform}". ` +
      `DeepScanBot CLI supports macOS, Linux, and Windows.`
    );
  }

  return mapping[platform];
}

/**
 * Map Node.js process.arch to Go arch naming convention.
 */
function detectArch() {
  const arch = process.arch;
  const mapping = {
    x64: "amd64",
    arm64: "arm64",
  };

  if (!mapping[arch]) {
    throw new Error(
      `Unsupported CPU architecture: "${arch}". ` +
      `DeepScanBot CLI supports amd64 (x86_64) and arm64 (ARM64).`
    );
  }

  return mapping[arch];
}

/**
 * Validate that the OS/arch combination is one we support.
 */
function validatePlatform(os, arch) {
  const supported = [
    ["darwin", "amd64"],
    ["darwin", "arm64"],
    ["linux", "amd64"],
    ["linux", "arm64"],
    ["windows", "amd64"],
  ];

  const match = supported.some(([sos, sarch]) => sos === os && sarch === arch);
  if (!match) {
    throw new Error(
      `Unsupported platform combination: ${os}/${arch}. ` +
      `DeepScanBot CLI supports: ` +
      `macOS (Intel + Apple Silicon), Linux (amd64 + arm64), Windows (amd64).`
    );
  }
}

/**
 * Get the binary filename inside the GoReleaser output directory.
 */
function getBinaryFilename(os) {
  return os === "windows" ? `${BINARY_NAME}.exe` : BINARY_NAME;
}

/**
 * Verify the installed binary works correctly.
 */
function verifyBinary(binaryPath, os) {
  try {
    const cmd = os === "windows" ? `"${binaryPath}" --version` : `"${binaryPath}" --version`;
    const output = execSync(cmd, {
      encoding: "utf-8",
      timeout: 5000,
      stdio: ["pipe", "pipe", "pipe"]
    }).trim();

    console.log(`[deepscanbot] Binary verification: OK (version: ${output})`);
    return true;
  } catch (err) {
    // Try --help as fallback
    try {
      const cmd = os === "windows" ? `"${binaryPath}" --help` : `"${binaryPath}" --help`;
      execSync(cmd, {
        encoding: "utf-8",
        timeout: 5000,
        stdio: ["pipe", "pipe", "pipe"]
      });
      console.log(`[deepscanbot] Binary verification: OK (help command works)`);
      return true;
    } catch (helpErr) {
      console.warn(`[deepscanbot] Warning: Binary verification failed.`);
      console.warn(`[deepscanbot]   Version check: ${err.message}`);
      console.warn(`[deepscanbot]   Help check: ${helpErr.message}`);
      console.warn(`[deepscanbot]   The binary was installed but may not work correctly.`);
      console.warn(`[deepscanbot]   Try running: ${binaryPath} --help`);
      return false;
    }
  }
}

/**
 * Scan dist/ to find the directory that matches the detected OS and arch.
 *
 * GoReleaser creates directories named like:
 *   deepscanbot_linux_amd64_v1
 *   deepscanbot_linux_arm64_v8.0
 *   deepscanbot_darwin_amd64_v1
 *   deepscanbot_darwin_arm64_v8.0
 *   deepscanbot_windows_amd64_v1
 *
 * We find the directory by scanning all entries in dist/, filtering for
 * directories whose name contains both the project name, the OS, and the arch.
 * The version suffix (_v1, _v8.0) is ignored.
 *
 * @returns {string|null} Full path to the matching directory, or null.
 */
function findBinaryDirectory(os, arch) {
  if (!fs.existsSync(DIST_DIR)) {
    return null;
  }

  const entries = fs.readdirSync(DIST_DIR, { withFileTypes: true });

  // Build matching pattern parts. We scan case-insensitively for safety.
  const osPattern = os.toLowerCase();
  const archPattern = arch.toLowerCase();

  for (const entry of entries) {
    if (!entry.isDirectory()) continue;

    const dirName = entry.name.toLowerCase();

    // Directory name must contain the project name, OS, and arch
    if (
      dirName.includes(PROJECT_NAME.toLowerCase()) &&
      dirName.includes(osPattern) &&
      dirName.includes(archPattern)
    ) {
      return path.join(DIST_DIR, entry.name);
    }
  }

  return null;
}

/**
 * Main installation logic.
 */
function install() {
  try {
    // 1. Detect platform
    const os = detectOS();
    const arch = detectArch();
    validatePlatform(os, arch);

    console.log(`[deepscanbot] Detected platform: ${os}/${arch}`);

    // 2. List dist/ contents for debugging
    if (fs.existsSync(DIST_DIR)) {
      const entries = fs.readdirSync(DIST_DIR, { withFileTypes: true });
      const dirs = entries
        .filter((e) => e.isDirectory())
        .map((e) => e.name);
      console.log(`[deepscanbot] dist/ subdirectories: ${dirs.length > 0 ? dirs.join(", ") : "(none)"}`);
    }

    // 3. Find the matching GoReleaser output directory
    const binaryDir = findBinaryDirectory(os, arch);

    if (!binaryDir) {
      const entries = fs.existsSync(DIST_DIR)
        ? fs.readdirSync(DIST_DIR, { withFileTypes: true })
            .filter((e) => e.isDirectory())
            .map((e) => e.name)
        : [];

      throw new Error(
        `Could not find a GoReleaser output directory for ${os}/${arch} in ${DIST_DIR}.\n` +
        `  Detected OS:         ${os}\n` +
        `  Detected Arch:       ${arch}\n` +
        `  Directories found:   ${entries.length > 0 ? entries.join(", ") : "(none)"}\n` +
        `  Expected pattern:    ${PROJECT_NAME}_${os}_${arch}_v*/\n` +
        `\n` +
        `Possible causes:\n` +
        `  - The package was not built for your platform.\n` +
        `  - The dist/ directory is empty or missing.\n` +
        `  - Run "goreleaser build --snapshot --clean" first.`
      );
    }

    console.log(`[deepscanbot] Found binary directory: ${path.relative(PKG_ROOT, binaryDir)}`);

    // 4. Locate the binary inside the directory
    const binFilename = getBinaryFilename(os);
    const srcPath = path.join(binaryDir, binFilename);

    if (!fs.existsSync(srcPath)) {
      throw new Error(
        `Binary not found inside the expected directory.\n` +
        `  Expected path: ${srcPath}\n` +
        `  Detected OS:   ${os}\n` +
        `  Detected Arch: ${arch}\n` +
        `\n` +
        `Contents of ${binaryDir}:\n` +
        `  ${fs.readdirSync(binaryDir).join(", ")}`
      );
    }

    // 5. Ensure bin/ directory exists
    if (!fs.existsSync(BIN_DIR)) {
      fs.mkdirSync(BIN_DIR, { recursive: true });
    }

    // 6. Copy binary to bin/
    const targetName = os === "windows" ? `${BINARY_NAME}.exe` : BINARY_NAME;
    const targetPath = path.join(BIN_DIR, targetName);

    console.log(`[deepscanbot] Installing binary: ${srcPath} -> ${targetPath}`);
    fs.copyFileSync(srcPath, targetPath);

    // 7. Apply executable permissions on Unix
    if (os !== "windows") {
      fs.chmodSync(targetPath, 0o755);
      console.log(`[deepscanbot] Applied executable permissions to ${targetPath}`);
    }

    // 8. Verify the binary runs
    const verified = verifyBinary(targetPath, os);
    if (!verified) {
      console.warn(`[deepscanbot] Installation completed with warnings.`);
    }

    console.log(`[deepscanbot] Installation complete!`);
    console.log(`[deepscanbot] Run "deepscanbot -h" to get started.`);
  } catch (err) {
    console.error(`[deepscanbot] Installation failed: ${err.message}`);
    process.exit(1);
  }
}

install();