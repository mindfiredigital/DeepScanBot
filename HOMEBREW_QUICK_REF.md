# Quick Reference: Why Homebrew Isn't Working

## The Problem (Current State)

```
┌─────────────────────────────────────────────────────────┐
│  What You WANT:                                         │
│  ✅ Merge to main → Workflow runs → homebrew-tap branch │
│                                                         │
│  What's ACTUALLY Happening:                             │
│  ❌ Merge to main → Nothing happens                     │
│  ❌ No releases published → Workflow never runs          │
│  ❌ homebrew-tap branch doesn't exist                   │
│  ❌ Cask file doesn't exist                             │
└─────────────────────────────────────────────────────────┘
```

## Root Causes

### 1. Wrong Trigger ❌
```yaml
# CURRENT (WRONG):
on:
  release:
    types: [published]  # Only runs when GitHub Release is published

# NEEDED (CORRECT):
on:
  push:
    branches: [main]    # Runs when code is merged to main
```

### 2. Wrong Checkout ❌
```yaml
# CURRENT (WRONG):
ref: ${{ github.event.release.tag_name }}  # Checks out release tag

# NEEDED (CORRECT):
ref: main  # Checks out main branch
```

### 3. Missing Cask File ❌
```bash
# homebrew-tap/Casks/deepscanbot.rb DOESN'T EXIST YET
# GoReleaser needs to generate it first
```

### 4. Wrong Tap URL ❌
```bash
# CURRENT (WRONG - expects separate repo):
brew tap mindfiredigital/tap

# NEEDED (CORRECT - uses branch in same repo):
brew tap mindfiredigital/DeepScanBot homebrew-tap
```

## The 3-Step Fix

### Step 1: Update homebrew-sync.yml
```yaml
# Change the trigger
on:
  push:
    branches:
      - main
  workflow_dispatch:

# Change the checkout
- uses: actions/checkout@v4
  with:
    ref: main  # Not: ${{ github.event.release.tag_name }}
```

### Step 2: Generate the Cask File
```bash
goreleaser release --snapshot --clean --skip=publish
cp dist/homebrew-tap/Casks/deepscanbot.rb homebrew-tap/Casks/deepscanbot.rb
git add homebrew-tap/Casks/deepscanbot.rb
git commit -m "feat: add homebrew cask"
git push
```

### Step 3: Create the homebrew-tap Branch
```bash
git checkout --orphan homebrew-tap
git rm -rf .
mkdir -p Casks
cp ../homebrew-tap/Casks/deepscanbot.rb Casks/
# Create README.md
git add .
git commit -m "chore: initial homebrew-tap branch"
git push origin homebrew-tap
git checkout main
```

### Step 4: Fix Documentation
```bash
# In apps/docs/docs/installation.mdx, change:
brew tap mindfiredigital/tap

# To:
brew tap mindfiredigital/DeepScanBot homebrew-tap
```

## After the Fix: How It Works

```
┌──────────────┐
│ Merge to main│
└──────┬───────┘
       │
       ▼
┌──────────────────────┐
│ GitHub Actions       │
│ workflow triggers    │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ Checkout main branch                 │
│ Find: homebrew-tap/Casks/*.rb        │
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ Create/update homebrew-tap branch    │
│ Copy cask to: Casks/deepscanbot.rb   │
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ Push homebrew-tap branch             │
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ Users can now install:               │
│ brew tap mindfiredigital/DeepScanBot  │
│   homebrew-tap                       │
│ brew install deepscanbot             │
└──────────────────────────────────────┘
```

## Quick Test

After making changes and pushing to main:

1. Go to: https://github.com/mindfiredigital/DeepScanBot/actions
2. Look for "Sync Homebrew Tap" workflow
3. Click on it to see if it ran
4. Check the `homebrew-tap` branch was updated

## Common Mistakes

❌ **Don't:** Use `brew tap mindfiredigital/tap` (expects separate repo)
✅ **Do:** Use `brew tap mindfiredigital/DeepScanBot homebrew-tap` (uses branch)

❌ **Don't:** Expect workflow to run on release publish only
✅ **Do:** Configure it to run on push to main

❌ **Don't:** Checkout release tag in the workflow
✅ **Do:** Checkout main branch

❌ **Don't:** Forget to create the initial cask file
✅ **Do:** Generate it with GoReleaser first
