#!/usr/bin/env node

/**
 * scripts/verify-binary.js
 *
 * Verifies that the pre-built Go binaries in dist/ are valid executables.
 *
 * GoReleaser output layout:
 *   dist/deepscanbot_<os>_<arch>_v<version>/my-cli
 *   dist/deepscanbot_windows_amd64_v1/my-cli.exe
 *
 * Checks:
 *   - Directory exists for each platform
 *   - Binary file exists inside the directory
 *   - File is not empty
 *   - File has correct permissions
 *   - SHA256 checksum matches (if checksums.txt is present)
 *   - Binary format (ELF, Mach-O, PE) is valid
 *
 * Usage:
 *   node scripts/verify-binary.js
 */

"use strict";

const fs = require("fs");
const path = require("path");
const crypto = require("crypto");

const PROJECT_ROOT = path.resolve(__dirname, "..");
const DIST_DIR = path.join(PROJECT_ROOT, "dist");
const BINARY_NAME = "deepscanbot";
const PROJECT_NAME = "deepscanbot";

// Expected platform combinations
const EXPECTED_PLATFORMS = [
  { os: "linux", arch: "amd64" },
  { os: "linux", arch: "arm64" },
  { os: "darwin", arch: "amd64" },
  { os: "darwin", arch: "arm64" },
  { os: "windows", arch: "amd64" },
];

/**
 * Find the GoReleaser output directory for a given OS/arch.
 * Scans dist/ for directories named like: deepscanbot_<os>_<arch>_v*
 */
function findBinaryDirectory(os, arch) {
  if (!fs.existsSync(DIST_DIR)) return null;

  const entries = fs.readdirSync(DIST_DIR, { withFileTypes: true });
  for (const entry of entries) {
    if (!entry.isDirectory()) continue;
    const dirName = entry.name.toLowerCase();
    if (
      dirName.includes(PROJECT_NAME.toLowerCase()) &&
      dirName.includes(os.toLowerCase()) &&
      dirName.includes(arch.toLowerCase())
    ) {
      return path.join(DIST_DIR, entry.name);
    }
  }
  return null;
}

/**
 * Get the binary filename inside the GoReleaser output directory.
 */
function getBinaryFilename(os) {
  return os === "windows" ? `${BINARY_NAME}.exe` : BINARY_NAME;
}

/**
 * Load checksums from checksums.txt (GoReleaser format).
 * Format: <sha256>  <filename>
 * Note: GoReleaser uses flat filenames in checksums, not paths.
 */
function loadChecksums() {
  const checksumsPath = path.join(DIST_DIR, "checksums.txt");
  if (!fs.existsSync(checksumsPath)) {
    console.log("[verify-binary] No checksums.txt found. Skipping checksum verification.");
    return null;
  }

  const checksums = {};
  const content = fs.readFileSync(checksumsPath, "utf-8");
  for (const line of content.trim().split("\n")) {
    const [hash, filename] = line.trim().split(/\s+/);
    if (hash && filename) {
      checksums[filename] = hash;
    }
  }
  console.log(`[verify-binary] Loaded ${Object.keys(checksums).length} checksums from checksums.txt`);
  return checksums;
}

/**
 * Compute SHA256 hash of a file.
 */
function sha256(filePath) {
  return new Promise((resolve, reject) => {
    const hash = crypto.createHash("sha256");
    const stream = fs.createReadStream(filePath);
    stream.on("data", (data) => hash.update(data));
    stream.on("end", () => resolve(hash.digest("hex")));
    stream.on("error", reject);
  });
}

/**
 * Verify the binary for a single platform.
 */
async function verifyPlatform(platform) {
  const { os, arch } = platform;
  const issues = [];

  // 1. Find the directory
  const binaryDir = findBinaryDirectory(os, arch);
  if (!binaryDir) {
    return {
      name: `${os}/${arch}`,
      ok: false,
      issues: [`No matching directory found in ${DIST_DIR} for ${PROJECT_NAME}_${os}_${arch}_v*`],
    };
  }

  // 2. Locate the binary
  const binFilename = getBinaryFilename(os);
  const filePath = path.join(binaryDir, binFilename);

  if (!fs.existsSync(filePath)) {
    return {
      name: `${os}/${arch}`,
      ok: false,
      issues: [`Binary not found at ${filePath}`],
    };
  }

  // 3. Check file is not empty
  const stats = fs.statSync(filePath);
  if (stats.size === 0) {
    issues.push("File is empty");
  }

  // 4. Check file permissions (Unix)
  if (process.platform !== "win32") {
    const mode = stats.mode & 0o777;
    if (!(mode & 0o111)) {
      issues.push(`File is not executable (mode: ${mode.toString(8)})`);
    }
  }

  // 5. Check file signature (ELF, Mach-O, PE)
  const buffer = Buffer.alloc(4);
  const fd = fs.openSync(filePath, "r");
  fs.readSync(fd, buffer, 0, 4, 0);
  fs.closeSync(fd);

  const magic = buffer.toString("hex");
  const isELF = magic === "7f454c46";
  const isMachO = magic === "feedface" || magic === "feedfacf" || magic === "cffaedfe" || magic === "cefaedfe";
  const isPE = magic.startsWith("4d5a");

  if (!isELF && !isMachO && !isPE) {
    issues.push(`Unknown binary format (magic: ${magic})`);
  }

  const format = isELF ? "ELF" : isMachO ? "Mach-O" : isPE ? "PE" : "Unknown";

  // 6. Verify checksum
  const checksums = loadChecksums();
  if (checksums) {
    // GoReleaser checksums use flat filenames (e.g., "my-cli_linux_amd64_v1")
    const dirName = path.basename(binaryDir);
    const expectedHash = checksums[dirName];
    if (expectedHash) {
      const actualHash = await sha256(filePath);
      if (actualHash !== expectedHash) {
        issues.push(
          `SHA256 mismatch: expected ${expectedHash}, got ${actualHash}`
        );
      }
    } else {
      issues.push(`No checksum found for directory ${dirName} in checksums.txt`);
    }
  }

  return {
    name: `${os}/${arch}`,
    path: path.relative(PROJECT_ROOT, filePath),
    ok: issues.length === 0,
    issues,
    size: stats.size,
    format,
  };
}

/**
 * Main verification logic.
 */
async function main() {
  console.log("[verify-binary] DeepScanBot CLI Binary Verification");
  console.log("[verify-binary] ====================================");
  console.log(`[verify-binary] Directory: ${DIST_DIR}`);
  console.log("");

  if (!fs.existsSync(DIST_DIR)) {
    console.error(`[verify-binary] ERROR: dist/ directory not found at ${DIST_DIR}`);
    console.error("[verify-binary] Run GoReleaser first to generate binaries.");
    process.exit(1);
  }

  // List available directories
  const dirs = fs.readdirSync(DIST_DIR, { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name);
  console.log(`[verify-binary] Available directories: ${dirs.length > 0 ? dirs.join(", ") : "(none)"}`);
  console.log("");

  const results = await Promise.all(EXPECTED_PLATFORMS.map(verifyPlatform));

  let passed = 0;
  let failed = 0;

  for (const result of results) {
    if (result.ok) {
      console.log(
        `  ✓ ${result.name} (${(result.size / 1024).toFixed(1)} KB, ${result.format})`
      );
      if (result.path) {
        console.log(`      Path: ${result.path}`);
      }
      passed++;
    } else {
      console.log(`  ✗ ${result.name}`);
      for (const issue of result.issues) {
        console.log(`      - ${issue}`);
      }
      failed++;
    }
  }

  console.log("");
  console.log(`[verify-binary] Results: ${passed} passed, ${failed} failed`);

  if (failed > 0) {
    process.exit(1);
  }

  console.log("[verify-binary] All binaries verified successfully!");
}

main().catch((err) => {
  console.error(`[verify-binary] Fatal error: ${err.message}`);
  process.exit(1);
});