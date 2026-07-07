#!/usr/bin/env node

/**
 * scripts/prepublish-check.js
 *
 * Pre-publish validation script.
 * Runs before `npm publish` to ensure the package is ready.
 *
 * Checks:
 * 1. package.json has a valid name, version, bin entry, and files array pattern matching.
 * 2. Required root files (README.md, LICENSE.md, postinstall.js) exist.
 * 3. dist/ directory exists and contains platform subdirectories with binaries.
 * 4. Checksums file exists (warning only, as snapshot builds omit it).
 * 5. Binaries are not empty, have executable permissions (Unix), and can be executed natively (if applicable).
 * 6. postinstall.js exists and is syntactically valid via native node syntax validation.
 *
 * Usage:
 * node scripts/prepublish-check.js
 */

"use strict";

const fs = require("fs");
const path = require("path");
const os = require("os");
const childProcess = require("child_process");

const PROJECT_ROOT = path.resolve(__dirname, "..");
const DIST_DIR = path.join(PROJECT_ROOT, "dist");

const PROJECT_NAME = "deepscanbot";
const BINARY_NAME = "deepscanbot";
const EXPECTED_PACKAGE_NAME = "@mindfiredigital/deepscanbot";

const REQUIRED_FILES = [
  "package.json",
  "postinstall.js",
  "README.md",
  "LICENSE.md",
];

const EXPECTED_PLATFORMS = [
  { os: "linux", arch: "amd64", example: "dist/deepscanbot_linux_amd64_v1/" },
  { os: "linux", arch: "arm64", example: "dist/deepscanbot_linux_arm64_v8.0/" },
  { os: "darwin", arch: "amd64", example: "dist/deepscanbot_darwin_amd64_v1/" },
  { os: "darwin", arch: "arm64", example: "dist/deepscanbot_darwin_arm64_v8.0/" },
  { os: "windows", arch: "amd64", example: "dist/deepscanbot_windows_amd64_v1/" },
];

let errors = [];
let warnings = [];
let checkedBinariesCount = 0;
let validatedPlatforms = [];
let parsedPackageName = "Unknown";
let parsedPackageVersion = "Unknown";

// Semantic Versioning Regex
const SEMVER_REGEX = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-zA-Z0-9-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-zA-Z0-9-]*))*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$/;

/**
 * Find the GoReleaser output directory for a given OS/arch.
 */
function findBinaryDirectory(platformOs, platformArch) {
  if (!fs.existsSync(DIST_DIR)) return null;
  const entries = fs.readdirSync(DIST_DIR, { withFileTypes: true });
  for (const entry of entries) {
    if (!entry.isDirectory()) continue;
    const dirName = entry.name.toLowerCase();
    if (
      dirName.includes(PROJECT_NAME.toLowerCase()) &&
      dirName.includes(platformOs.toLowerCase()) &&
      dirName.includes(platformArch.toLowerCase())
    ) {
      return path.join(DIST_DIR, entry.name);
    }
  }
  return null;
}

console.log("[prepublish] Starting pre-publish validations...\n");

// 1. Check required root files
console.log("[prepublish] 1. Checking required root files...");
for (const file of REQUIRED_FILES) {
  const filePath = path.join(PROJECT_ROOT, file);
  if (!fs.existsSync(filePath)) {
    errors.push(`Required file not found: ${file}`);
  } else {
    console.log(`  ✓ ${file} exists`);
  }
}

// 2. Check package.json strict validations
console.log("\n[prepublish] 2. Checking package.json requirements...");
try {
  const pkgContent = fs.readFileSync(path.join(PROJECT_ROOT, "package.json"), "utf-8");
  const pkg = JSON.parse(pkgContent);

  parsedPackageName = pkg.name || "Missing";
  parsedPackageVersion = pkg.version || "Missing";

  // Name validation
  if (pkg.name !== EXPECTED_PACKAGE_NAME) {
    errors.push(`Invalid package name: Expected "${EXPECTED_PACKAGE_NAME}", found "${pkg.name}"`);
  } else {
    console.log(`  ✓ Package name: ${pkg.name}`);
  }

  // Version validation
  if (!pkg.version || pkg.version === "0.0.0" || !SEMVER_REGEX.test(pkg.version)) {
    errors.push(`Invalid package version: "${pkg.version}". Must be a valid, non-zero semantic version.`);
  } else {
    console.log(`  ✓ Valid SemVer version: ${pkg.version}`);
  }

  // Bin validation
  if (!pkg.bin || pkg.bin[BINARY_NAME] !== "./bin/deepscanbot") {
    errors.push(`Invalid or missing "bin" entry. Expected: "bin": { "deepscanbot": "./bin/deepscanbot" }`);
  } else {
    console.log(`  ✓ Bin entry correctly mapped: ${pkg.bin[BINARY_NAME]}`);
  }

  // Files validation (flexible matching patterns like dist/**/* or bin/**)
  const requiredFilesArray = ["dist", "bin", "postinstall.js", "README.md", "LICENSE.md"];
  if (!pkg.files || !Array.isArray(pkg.files)) {
    errors.push("package.json is missing the 'files' array.");
  } else {
    const missingFiles = requiredFilesArray.filter(req => {
      return !pkg.files.some(f => f === req || f.startsWith(`${req}/`) || f.startsWith(`${req}\\`));
    });

    if (missingFiles.length > 0) {
      errors.push(`package.json 'files' array is missing required entries or glob patterns for: ${missingFiles.join(", ")}`);
    } else {
      console.log(`  ✓ "files" array contains all required paths or wildcards`);
    }
  }

} catch (err) {
  errors.push(`Failed to read or parse package.json: ${err.message}`);
}

// 3. Check dist/ directory and binaries
console.log("\n[prepublish] 3. Checking dist/ directory and GoReleaser output...");
if (!fs.existsSync(DIST_DIR)) {
  errors.push("dist/ directory not found. Run GoReleaser first.");
} else {
  // Check checksums (Demoted to Warning since snapshot builds omit this file)
  const checksumsPath = path.join(DIST_DIR, "checksums.txt");
  if (fs.existsSync(checksumsPath)) {
    console.log("  ✓ checksums.txt found");
  } else {
    warnings.push("checksums.txt is missing from dist/ directory (Expected for local snapshot builds, but required for production release versions).");
  }

  const currentOsStr = os.platform() === "win32" ? "windows" : os.platform();
  const currentArchStr = os.arch() === "x64" ? "amd64" : (os.arch() === "arm64" ? "arm64" : os.arch());

  // Validate each expected platform
  for (const platform of EXPECTED_PLATFORMS) {
    const binaryDir = findBinaryDirectory(platform.os, platform.arch);

    if (!binaryDir) {
      errors.push(`Missing build directory for ${platform.os}/${platform.arch}. Expected GoReleaser output similar to: ${platform.example}`);
      continue;
    }

    const binFilename = platform.os === "windows" ? `${BINARY_NAME}.exe` : BINARY_NAME;
    const binPath = path.join(binaryDir, binFilename);

    if (fs.existsSync(binPath)) {
      const stats = fs.statSync(binPath);

      // Empty binary check
      if (stats.size === 0) {
        errors.push(`Binary is empty: ${binPath}`);
        continue;
      }

      console.log(`  ✓ ${platform.os}/${platform.arch} binary found (${(stats.size / 1024 / 1024).toFixed(2)} MB)`);
      checkedBinariesCount++;
      validatedPlatforms.push(`${platform.os}/${platform.arch}`);

      // Executable bit check (Unix)
      if (platform.os !== "windows") {
        try {
          fs.accessSync(binPath, fs.constants.X_OK);
          console.log(`    ✓ Executable permissions set (chmod +x)`);
        } catch (e) {
          errors.push(`Binary is not executable (missing execution permissions): ${binPath}`);
        }
      }

      // Execution check (Native platform only)
      if (platform.os === currentOsStr && platform.arch === currentArchStr) {
        try {
          childProcess.execSync(`"${binPath}" --help`, { stdio: "ignore" });
          console.log(`    ✓ Execution test successful (ran --help natively)`);
        } catch (execErr) {
          errors.push(`Failed to execute native binary ${binPath}. Error: ${execErr.message}`);
        }
      } else {
        console.log(`    - Skipping execution test (cross-platform mismatch)`);
      }

    } else {
      errors.push(`Binary not found in ${binaryDir}: expected ${binFilename}`);
    }
  }
}

// 4. Check postinstall.js syntax using runtime execution check flags
console.log("\n[prepublish] 4. Checking postinstall.js syntax...");
try {
  const postinstallPath = path.join(PROJECT_ROOT, "postinstall.js");
  const result = childProcess.spawnSync(process.execPath, ["--check", postinstallPath], { encoding: "utf-8" });

  if (result.status !== 0) {
    errors.push(`postinstall.js has syntax errors:\n${result.stderr || result.stdout}`);
  } else {
    console.log("  ✓ postinstall.js syntax OK");
  }
} catch (syntaxErr) {
  errors.push(`Failed to run syntax compilation test on postinstall.js: ${syntaxErr.message}`);
}

// 5. Final Summary
console.log("\n=======================================================");
console.log("[prepublish] FINAL SUMMARY");
console.log("=======================================================");
console.log(`Package Name       : ${parsedPackageName}`);
console.log(`Package Version    : ${parsedPackageVersion}`);
console.log(`Binaries Checked   : ${checkedBinariesCount}`);
console.log(`Valid Platforms    : ${validatedPlatforms.length > 0 ? validatedPlatforms.join(", ") : "None"}`);
console.log("-------------------------------------------------------");

if (warnings.length > 0) {
  console.log(`⚠ WARNINGS (${warnings.length}):`);
  for (const warn of warnings) {
    console.log(`  - ${warn}`);
  }
  console.log("");
}

if (errors.length > 0) {
  console.log(`❌ ERRORS (${errors.length}):`);
  for (const err of errors) {
    console.log(`  - ${err}`);
  }
  console.log("=======================================================");
  console.log("❌ Pre-publish validations failed. Fix errors before publishing.");
  process.exit(1);
} else {
  console.log("=======================================================");
  console.log("✅ All checks passed. Ready to publish.");
  process.exit(0);
}