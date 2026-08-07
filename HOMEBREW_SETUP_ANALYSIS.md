# Homebrew Setup Analysis & Issues

## Current State

### ✅ What's Working:
1. **homebrew-sync.yml workflow exists** at `.github/workflows/homebrew-sync.yml`
2. **GoReleaser is configured** to generate cask files in `.goreleaser.yml`
3. **homebrew-tap directory exists** with README.md and Casks/ folder structure
4. **Documentation mentions Homebrew** in `apps/docs/docs/installation.mdx`

### ❌ What's NOT Working:

## Root Causes

### 1. **Missing Separate homebrew-tap Repository**
**Problem:** The docs say `brew tap mindfiredigital/tap` which requires a **separate GitHub repository** called `mindfiredigital/homebrew-tap`. This repository **does not exist yet**.

**Current Implementation (WRONG):**
- The `homebrew-sync.yml` workflow creates a `homebrew-tap` **branch** in the same DeepScanBot repository
- This won't work with `brew tap mindfiredigital/tap`

**What's Needed:**
- Create a new repository: `https://github.com/mindfiredigital/homebrew-tap`
- Update the workflow to push to this separate repository

### 2. **homebrew-tap Branch Doesn't Exist**
```bash
$ git ls-remote --heads origin homebrew-tap
# (no output - branch doesn't exist)
```

The workflow is supposed to create this branch, but it has **never run** because:

### 3. **homebrew-sync.yml Workflow Has Never Run**
**Why it hasn't run:**
- The workflow triggers on: `release: types: [published]`
- This means it only runs when a GitHub Release is **published** (not just created)
- **No releases have been published yet** from this repository
- The workflow also supports manual trigger via `workflow_dispatch`, but it hasn't been triggered manually either

### 4. **Casks Directory is Empty**
```bash
$ ls -la homebrew-tap/Casks/
total 8
drwxr-xr-x 2 lenovo anuj 4096 Aug  5 12:00 .
drwxr-xr-x 3 lenovo anuj 4096 Aug  7 10:14 ..
```

The `homebrew-tap/Casks/deepscanbot.rb` file doesn't exist because:
- GoReleaser generates it during the release process
- But releases haven't happened yet
- Also, `.goreleaser.yml` has `skip_upload: true` for homebrew_casks

## The Documentation Message

The docs say:
```
Status: Homebrew installation is currently unavailable until the 
mindfiredigital/homebrew-tap repository is released.
```

This is **accurate** because:
1. The `mindfiredigital/homebrew-tap` repository doesn't exist
2. Even if it did, no releases have been published to trigger the workflow
3. The workflow hasn't run, so no cask has been synced

## What Needs to Happen

### Step 1: Create the homebrew-tap Repository
```bash
# Create new repository on GitHub: mindfiredigital/homebrew-tap
# Make it public
```

### Step 2: Update homebrew-sync.yml Workflow
The workflow needs to:
1. Checkout the release tag (already does this)
2. Extract the cask file from `homebrew-tap/Casks/deepscanbot.rb` in the release tag
3. Push to the **separate** `mindfiredigital/homebrew-tap` repository, not create a branch in DeepScanBot

### Step 3: Create a GitHub Release
- Tag: `v1.0.0` (or whatever version)
- Publish the release
- This triggers the release.yml workflow
- GoReleaser generates binaries and the cask file
- The homebrew-sync.yml workflow should then sync to the tap repository

### Step 4: Update .goreleaser.yml
Remove or set `skip_upload: false` for homebrew_casks if you want GoReleaser to handle the upload directly.

## Why the Workflow Won't Work As-Is

Looking at lines 60-115 of `homebrew-sync.yml`:

```yaml
# Check if homebrew-tap branch exists
if git ls-remote --heads origin homebrew-tap | grep -q homebrew-tap; then
  echo "Branch homebrew-tap exists, checking it out..."
  git checkout homebrew-tap
  git pull origin homebrew-tap
else
  echo "Branch homebrew-tap does not exist, creating orphan branch..."
  git checkout --orphan homebrew-tap
  git rm -rf .
fi

# ... copies files ...

# Push to homebrew-tap branch
git push origin homebrew-tap --force
```

This pushes to a **branch in the same repository**, but Homebrew expects a **separate repository** for taps.

## Quick Fix Checklist

- [ ] Create `mindfiredigital/homebrew-tap` repository on GitHub
- [ ] Update `.github/workflows/homebrew-sync.yml` to push to the separate repository
- [ ] Update `apps/docs/docs/installation.mdx` to remove "currently unavailable" message
- [ ] Create and publish a GitHub release (e.g., v1.0.0)
- [ ] Verify the workflow runs and syncs the cask
- [ ] Test: `brew tap mindfiredigital/tap && brew install deepscanbot`

## Alternative Approach (Simpler)

Instead of the complex homebrew-sync.yml workflow, you could:

1. Use GoReleaser's built-in homebrew tap support with a separate tap repository
2. Configure GoReleaser to push directly to `mindfiredigital/homebrew-tap`
3. This eliminates the need for the homebrew-sync.yml workflow entirely

See: https://goreleaser.com/customization/homebrew/
