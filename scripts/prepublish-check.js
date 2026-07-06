#!/usr/bin/env node

/**
 * scripts/prepublish-check.js
 *
 * Pre-publish validation script.
 * Runs before `npm publish` to ensure the package is ready.
 *
 * GoReleaser output layout (inside dist/):
 *   dist/deepscanbot_<os>_<arch>_v<version>/my-cli
 *   dist/deepscanbot_windows_amd64_v1/my-cli.exe
 *   dist/checksums.txt
 *
 * Checks:
 *   1. package.json has a valid version
 *   2. dist/ directory exists and contains platform subdirectories with binaries
 *   3. postinstall.js exists and is valid
 *   4. README.md exists
 *   5. Checksums file exists
 *
 * Usage:
 *   node scripts/prepublish-check.js
 *
 * This is called automatically via the "prepublishOnly" npm script.
 */

"use strict";

const fs = require("fs");
const path = require("path");

const PROJECT_ROOT = path.resolve(__dirname, "..");
const DIST_DIR = path.join(PROJECT_ROOT, "dist");
const BINARY_NAME = "my-cli";
const PROJECT_NAME = "deepscanbot";

const REQUIRED_FILES = [
  "package.json",
  "postinstall.js",
  "README.md",
];

const EXPECTED_PLATFORMS = [
  { os: "linux", arch: "amd64" },
  { os: "linux", arch: "arm64" },
  { os: "darwin", arch: "amd64" },
  { os: "darwin", arch: "arm64" },
  { os: "windows", arch: "amd64" },
];

let errors = [];
let warnings = [];

/**
 * Find the GoReleaser output directory for a given OS/arch.
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

// 1. Check required files exist
console.log("[prepublish] Checking required files...");
for (const file of REQUIRED_FILES) {
  const filePath = path.join(PROJECT_ROOT, file);
  if (!fs.existsSync(filePath)) {
    errors.push(`Required file not found: ${file}`);
  } else {
    console.log(`  ✓ ${file}`);
  }
}

// 2. Check package.json
console.log("[prepublish] Checking package.json...");
try {
  const pkg = JSON.parse(fs.readFileSync(path.join(PROJECT_ROOT, "package.json"), "utf-8"));

  if (!pkg.version || pkg.version === "0.0.0") {
    errors.push("package.json version is not set (0.0.0). Update with a real version.");
  } else {
    console.log(`  ✓ Version: ${pkg.version}`);
  }

  if (!pkg.bin || Object.keys(pkg.bin).length === 0) {
    errors.push("package.json has no 'bin' entries.");
  } else {
    console.log(`  ✓ Bin entries: ${Object.keys(pkg.bin).join(", ")}`);
  }

  if (!pkg.files || pkg.files.length === 0) {
    warnings.push("package.json has no 'files' field. All files will be published.");
  }
} catch (err) {
  errors.push(`Invalid package.json: ${err.message}`);
}

// 3. Check dist/ directory
console.log("[prepublish] Checking dist/ directory...");
if (!fs.existsSync(DIST_DIR)) {
  errors.push("dist/ directory not found. Run GoReleaser first.");
} else {
  const allEntries = fs.readdirSync(DIST_DIR, { withFileTypes: true });
  const subdirs = allEntries.filter((e) => e.isDirectory()).map((e) => e.name);
  const files = allEntries.filter((e) => e.isFile()).map((e) => e.name);

  console.log(`  ✓ dist/ contains ${subdirs.length} subdirectories, ${files.length} files`);

  if (subdirs.length === 0) {
    errors.push("No directories found in dist/. Expected GoReleaser output directories.");
  } else {
    // Check each expected platform
    for (const platform of EXPECTED_PLATFORMS) {
      const binaryDir = findBinaryDirectory(platform.os, platform.arch);
      if (!binaryDir) {
        warnings.push(`Expected platform directory not found: ${PROJECT_NAME}_${platform.os}_${platform.arch}_v*`);
        continue;
      }

      const binFilename = platform.os === "windows" ? `${BINARY_NAME}.exe` : BINARY_NAME;
      const binPath = path.join(binaryDir, binFilename);

      if (fs.existsSync(binPath)) {
        const stats = fs.statSync(binPath);
        if (stats.size === 0) {
          errors.push(`Binary is empty: ${path.join(path.basename(binaryDir), binFilename)}`);
        } else {
          console.log(`  ✓ ${platform.os}/${platform.arch} (${(stats.size / 1024).toFixed(1)} KB)`);
        }
      } else {
        errors.push(`Binary not found in ${binaryDir}: expected ${binFilename}`);
      }
    }
  }

  // Check checksums
  if (fs.existsSync(path.join(DIST_DIR, "checksums.txt"))) {
    console.log("  ✓ checksums.txt");
  } else {
    warnings.push("checksums.txt not found in dist/");
  }
}

// 4. Check postinstall.js syntax
console.log("[prepublish] Checking postinstall.js...");
try {
  const content = fs.readFileSync(path.join(PROJECT_ROOT, "postinstall.js"), "utf-8");
  new Function(content);
  console.log("  ✓ postinstall.js syntax OK");
} catch (syntaxErr) {
  errors.push(`postinstall.js has syntax errors: ${syntaxErr.message}`);
}

// 5. Summary
console.log("");
console.log("[prepublish] =================================");
if (errors.length > 0) {
  console.log(`[prepublish] ❌ ${errors.length} error(s) found:`);
  for (const err of errors) {
    console.log(`  - ${err}`);
  }
}

if (warnings.length > 0) {
  console.log(`[prepublish] ⚠ ${warnings.length} warning(s):`);
  for (const warn of warnings) {
    console.log(`  - ${warn}`);
  }
}

if (errors.length === 0) {
  console.log("[prepublish] ✅ All checks passed. Ready to publish.");
} else {
  console.log("[prepublish] ❌ Fix errors before publishing.");
  process.exit(1);
}