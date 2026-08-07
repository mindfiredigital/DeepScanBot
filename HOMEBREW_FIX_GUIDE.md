# Homebrew Setup - Why It's Not Working & How to Fix It

## Your Requirement
✅ Use a dedicated `homebrew-tap` **branch** in the SAME repository (not a separate repo)
✅ Automatically run when code is merged to `main`
✅ Everything in the DeepScanBot repo

## Current Problems (3 Critical Issues)

### ❌ Issue #1: Wrong Trigger in homebrew-sync.yml

**Current (lines 3-6):**
```yaml
on:
  release:
    types: [published]
  workflow_dispatch:
```

**Problem:** This workflow ONLY runs when a GitHub Release is **published**, NOT when code is merged to `main`.

**What happens:**
- You merge code to main → ❌ Workflow does NOT run

## The Complete Solution

You need to make THREE main changes:

### Change 1: Fix homebrew-sync.yml Trigger

**File:** `.github/workflows/homebrew-sync.yml`

**Change from:**
```yaml
on:
  release:
    types: [published]
  workflow_dispatch:
```

**Change to:**
```yaml
on:
  push:
    branches:
      - main
  workflow_dispatch:
```

This makes the workflow run automatically when you merge to `main`.

### Change 2: Fix the Checkout Reference

**File:** `.github/workflows/homebrew-sync.yml`

**Change from (line 20-26):**
```yaml
- name: Checkout repository
  uses: actions/checkout@v4

## Additional Required Changes

### Change 4: Remove "currently unavailable" Message

**File:** `README.md` (and any docs showing this message)

Remove or update the status message that says Homebrew is unavailable.

### Change 5: Generate the Cask File

You need to create the initial `homebrew-tap/Casks/deepscanbot.rb` file. You have two options:

**Option A: Generate it locally with GoReleaser**
```bash
# Install GoReleaser if not already installed
go install github.com/goreleaser/goreleaser/v2@latest

# Generate the cask file (snapshot mode, no publish)
goreleaser release --snapshot --clean --skip=publish --skip=announce

# Copy the generated cask to your repository
cp dist/homebrew-tap/Casks/deepscanbot.rb homebrew-tap/Casks/deepscanbot.rb

# Commit and push
git add homebrew-tap/Casks/deepscanbot.rb
git commit -m "feat: add initial homebrew cask file"
git push
```

**Option B: Let the workflow generate it**

Modify the workflow to run GoReleaser to generate the cask file before syncing.

## How It Will Work After Fixes

### The Flow:

1. **You make changes** to the code/config
2. **You merge to main** → Triggers `homebrew-sync.yml`

## Complete Fixed Files

### Fixed .github/workflows/homebrew-sync.yml

```yaml
name: Sync Homebrew Tap

on:
  push:
    branches:
      - main
  workflow_dispatch:

# Ensure multiple runs don't overlap
concurrency:
  group: homebrew-tap-sync
  cancel-in-progress: false

permissions:
  contents: write

jobs:
  sync-tap:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          token: ${{ secrets.GITHUB_TOKEN }}
          ref: main

      - name: Setup Git
        run: |
          git config user.name "goreleaserbot"
          git config user.email "goreleaser@mindfiredigital.com"

      - name: Check if cask file exists
        run: |
          if [ ! -f homebrew-tap/Casks/deepscanbot.rb ]; then
            echo "❌ Error: Cask file not found at homebrew-tap/Casks/deepscanbot.rb"
            echo "   Please ensure GoReleaser has generated the cask file."
            exit 1
          fi
          echo "✅ Found cask file"

      - name: Create or update homebrew-tap branch
        run: |
          # Copy the cask file to temporary location
          mkdir -p /tmp/homebrew-cask
          cp homebrew-tap/Casks/deepscanbot.rb /tmp/homebrew-cask/deepscanbot.rb
          
          # Copy README if it exists
          if [ -f homebrew-tap/README.md ]; then
            cp homebrew-tap/README.md /tmp/homebrew-cask/README.md
          fi

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
          
          # Create Casks directory
          mkdir -p Casks
          
          # Copy the cask file
          cp /tmp/homebrew-cask/deepscanbot.rb Casks/deepscanbot.rb
          echo "✅ Copied cask file to homebrew-tap branch"
          
          # Copy README (optional)
          if [ -f /tmp/homebrew-cask/README.md ]; then
            cp /tmp/homebrew-cask/README.md README.md
            echo "✅ Copied README.md"
          else
            echo "📝 Creating default README.md for homebrew-tap"
            cat > README.md << 'EOF'
# DeepScanBot Homebrew Tap

This is the official Homebrew tap for DeepScanBot.

## Installation

```bash
brew tap mindfiredigital/DeepScanBot homebrew-tap
brew install deepscanbot
```

## Usage

```bash
deepscanbot --help
deepscanbot doctor
```

## Testing the Setup

### Step 1: Generate Initial Cask File

```bash
# Install GoReleaser if not already installed
go install github.com/goreleaser/goreleaser/v2@latest

# Generate the cask file (snapshot mode, no publish)
goreleaser release --snapshot --clean --skip=publish --skip=announce

# Copy the generated cask to your repository
cp dist/homebrew-tap/Casks/deepscanbot.rb homebrew-tap/Casks/deepscanbot.rb

# Commit and push
git add homebrew-tap/Casks/deepscanbot.rb
git commit -m "feat: add initial homebrew cask file"
git push origin main
```

### Step 2: Create the homebrew-tap Branch Manually (First Time)

```bash
# Create the branch
git checkout --orphan homebrew-tap
git rm -rf .

# Create directory structure
mkdir -p Casks

# Copy the cask file
cp ../homebrew-tap/Casks/deepscanbot.rb Casks/deepscanbot.rb

# Create README
cat > README.md << 'EOF'
# DeepScanBot Homebrew Tap

This is the official Homebrew tap for DeepScanBot.

## Installation

```bash
brew tap mindfiredigital/DeepScanBot homebrew-tap
brew install deepscanbot
```

## Usage

```bash

## Summary of Changes Needed

| File | Change | Why |
|------|--------|-----|
| `.github/workflows/homebrew-sync.yml` | Change trigger from `release` to `push: branches: [main]` | Run automatically on merge to main |
| `.github/workflows/homebrew-sync.yml` | Change checkout ref from `release.tag_name` to `main` | Checkout main branch instead of release tag |
| `apps/docs/docs/installation.mdx` | Change `brew tap mindfiredigital/tap` to `brew tap mindfiredigital/DeepScanBot homebrew-tap` | Correct syntax for branch tap |
| `README.md` | Remove "currently unavailable" message | Homebrew will work after fixes |
| `homebrew-tap/Casks/deepscanbot.rb` | Create this file (generate with GoReleaser) | Cask file must exist for workflow to work |
| `homebrew-tap` branch | Create this branch manually first time | Workflow will maintain it after that |

## Why Your Current Setup Failed

1. ❌ **Wrong trigger**: Workflow only ran on release publish, not on main merge
2. ❌ **Wrong checkout**: Looked for cask in release tag, not main branch
3. ❌ **Missing cask file**: Never generated or committed the cask file
4. ❌ **Wrong tap URL**: Pointed to non-existent separate repository
5. ❌ **Never tested**: No releases published, so workflow never ran

## After Making These Changes

EVERY merge to main will automatically:
- ✅ Run the homebrew-sync workflow
- ✅ Update the homebrew-tap branch
- ✅ Make the latest version available via Homebrew

No manual intervention needed!

deepscanbot --help
deepscanbot doctor
```

## More Information

- [Documentation](https://mindfiredigital.github.io/DeepScanBot/)
- [GitHub Repository](https://github.com/mindfiredigital/DeepScanBot)
- [Report Issues](https://github.com/mindfiredigital/DeepScanBot/issues)
EOF

# Commit and push
git add Casks/deepscanbot.rb README.md
git commit -m "chore: initial homebrew-tap branch"
git push origin homebrew-tap

# Switch back to main
git checkout main
```

### Step 3: Test the Workflow

1. Make a small change to any file
2. Commit and push to main
3. Go to GitHub Actions tab
4. Verify the "Sync Homebrew Tap" workflow runs
5. Check that the `homebrew-tap` branch was updated

### Step 4: Test Homebrew Installation

```bash
# Untap if already tapped
brew untap mindfiredigital/DeepScanBot

# Tap the branch
brew tap mindfiredigital/DeepScanBot homebrew-tap

# Install
brew install deepscanbot

# Verify
deepscanbot --help
```


## More Information

- [Documentation](https://mindfiredigital.github.io/DeepScanBot/)
- [GitHub Repository](https://github.com/mindfiredigital/DeepScanBot)
- [Report Issues](https://github.com/mindfiredigital/DeepScanBot/issues)
EOF
          fi
          
          # Add and commit
          git add Casks/deepscanbot.rb README.md
          git diff --cached --quiet || git commit -m "chore(homebrew): sync cask to homebrew-tap branch"
          
          # Push to homebrew-tap branch
          git push origin homebrew-tap --force

      - name: Verify cask file
        run: |
          if [ -f Casks/deepscanbot.rb ]; then
            echo "✅ Cask file synced successfully"
            echo ""
            echo "Cask file contents:"
            cat Casks/deepscanbot.rb
          else
            echo "❌ Error: Cask file not found after sync"
            exit 1
          fi
```

### Fixed apps/docs/docs/installation.mdx (macOS section)

```markdown
<TabItem value="macos" label="macOS">

**Homebrew:**

```bash
brew tap mindfiredigital/DeepScanBot homebrew-tap
brew install deepscanbot
```

To upgrade to the latest version:

```bash
brew upgrade deepscanbot
```

**One-liner:**

```bash
curl -fsSL https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.sh | bash
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.sh | bash -s -- -b /usr/local/bin -v v1.0.0
```

The script detects your OS and architecture, downloads the binary, verifies the checksum, and installs it.

**Manual download:**

| Architecture  | Binary Name                     |
| ------------- | ------------------------------- |
| Intel x64     | `deepscanbot_darwin_amd64`      |
| Apple Silicon | `deepscanbot_darwin_arm64`      |

```bash
chmod +x deepscanbot_darwin_*
sudo mv deepscanbot_darwin_amd64 /usr/local/bin/deepscanbot
```

</TabItem>
```

3. **Workflow runs:**
   - Checks out `main` branch
   - Finds `homebrew-tap/Casks/deepscanbot.rb` (generated by GoReleaser)
   - Creates or updates the `homebrew-tap` branch
   - Copies the cask file to `Casks/deepscanbot.rb` in the `homebrew-tap` branch
   - Pushes the `homebrew-tap` branch
4. **Users install:**
   ```bash
   brew tap mindfiredigital/DeepScanBot homebrew-tap
   brew install deepscanbot
   ```

### Homebrew Tap URL Format

For a branch in the same repository:
```
https://github.com/mindfiredigital/DeepScanBot/tree/homebrew-tap
```

Homebrew will:
1. Clone the DeepScanBot repository
2. Checkout the `homebrew-tap` branch
3. Look for `Casks/deepscanbot.rb`
4. Install the cask

  with:
    fetch-depth: 0
    token: ${{ secrets.GITHUB_TOKEN }}
    ref: ${{ github.event.release.tag_name }}
```

**Change to:**
```yaml
- name: Checkout repository
  uses: actions/checkout@v4
  with:
    fetch-depth: 0
    token: ${{ secrets.GITHUB_TOKEN }}
    ref: main
```

This checks out the `main` branch instead of a release tag.

### Change 3: Update Documentation

**File:** `apps/docs/docs/installation.mdx`

**Change from (line 40):**
```bash
brew tap mindfiredigital/tap
```

**Change to:**
```bash
brew tap mindfiredigital/DeepScanBot homebrew-tap
```

This tells Homebrew to tap the `homebrew-tap` branch from the DeepScanBot repository.

- You publish a GitHub Release → ✅ Workflow runs (but this is not what you want)

### ❌ Issue #2: Wrong Checkout Reference

**Current (line 25):**
```yaml
ref: ${{ github.event.release.tag_name }}
```

**Problem:** The workflow checks out a **release tag**, not the `main` branch.

When triggered by a release, it looks for the cask file in that specific release tag, but:
- The cask file is generated by GoReleaser during the release
- With `skip_upload: true` in `.goreleaser.yml`, the cask might not be in the release
- It can't find `homebrew-tap/Casks/deepscanbot.rb` and fails

### ❌ Issue #3: Cask File Doesn't Exist Yet

```bash
$ ls -la homebrew-tap/Casks/
total 8
drwxr-xr-x 2 lenovo anuj 4096 Aug  5 12:00 .
drwxr-xr-x 3 lenovo anuj 4096 Aug  7 10:14 ..
```

The `deepscanbot.rb` file doesn't exist because:
1. GoReleaser generates it during the release process
2. `.goreleaser.yml` has `skip_upload: true` for homebrew_casks (line 67)
3. No releases have been published yet
4. The file is never committed to the repository

### ❌ Issue #4: Wrong Tap URL in Documentation

**Current (apps/docs/docs/installation.mdx, line 40):**
```bash
brew tap mindfiredigital/tap
```

**Problem:** This expects a separate repository `mindfiredigital/homebrew-tap`, which doesn't exist.

**Correct syntax for a branch in the same repo:**
```bash
brew tap mindfiredigital/DeepScanBot homebrew-tap
```
