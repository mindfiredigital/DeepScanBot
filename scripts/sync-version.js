#!/usr/bin/env node

/**
 * scripts/sync-version.js
 *
 * Synchronizes the package.json version with a Git tag or environment variable.
 *
 * Usage:
 *   VERSION=1.2.3 node scripts/sync-version.js
 *
 * The version is read from:
 *   1. The VERSION environment variable (set by CI)
 *   2. The latest Git tag matching v*
 *   3. Falls back to the current package.json version
 */

"use strict";

const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const PKG_PATH = path.resolve(__dirname, "..", "package.json");

function getVersionFromEnv() {
  return process.env.VERSION || null;
}

function getVersionFromGit() {
  try {
    const tag = execSync(
      'git describe --tags --abbrev=0 --match "v*" 2>/dev/null || echo ""',
      { encoding: "utf-8", timeout: 5000 }
    ).trim();

    if (tag) {
      return tag.replace(/^v/, "");
    }
  } catch (_) {
    // Not a git repository or no tags
  }
  return null;
}

function syncVersion() {
  // Read current package.json
  const pkg = JSON.parse(fs.readFileSync(PKG_PATH, "utf-8"));
  const currentVersion = pkg.version;

  // Determine new version
  const newVersion = getVersionFromEnv() || getVersionFromGit() || currentVersion;

  if (newVersion === currentVersion) {
    console.log(`[sync-version] Version unchanged: ${currentVersion}`);
    return;
  }

  // Update package.json
  pkg.version = newVersion;
  fs.writeFileSync(PKG_PATH, JSON.stringify(pkg, null, 2) + "\n", "utf-8");

  console.log(`[sync-version] Version updated: ${currentVersion} -> ${newVersion}`);
}

syncVersion();