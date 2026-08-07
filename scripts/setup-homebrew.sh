#!/bin/bash
set -e

echo "======================================"
echo "DeepScanBot Homebrew Setup Script"
echo "======================================"
echo ""

# Step 1: Fix homebrew-sync.yml
echo "📝 Step 1: Updating homebrew-sync.yml workflow..."
cat > .github/workflows/homebrew-sync.yml << 'EOF'
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
            echo "   Please ensure the cask file exists in the repository."
            exit 1
          fi
          echo "✅ Found cask file at homebrew-tap/Casks/deepscanbot.rb"

      - name: Create or update homebrew-tap branch
        run: |
          # Copy the cask file to temporary location
          mkdir -p /tmp/homebrew-cask
          cp homebrew-tap/Casks/deepscanbot.rb /tmp/homebrew-cask/deepscanbot.rb
          
          # Copy README if it exists
          if [ -f homebrew-tap/README.md ]; then
            cp homebrew-tap/README.md /tmp/homebrew-cask/README.md
            echo "✅ Found README.md in homebrew-tap/README.md"
          fi
          
          # Validate cask file exists and is non-empty
          if [ ! -s /tmp/homebrew-cask/deepscanbot.rb ]; then
            echo "❌ Error: Cask file is missing or empty."
            exit 1
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
          
          # Copy the cask file from temporary location
          cp /tmp/homebrew-cask/deepscanbot.rb Casks/deepscanbot.rb
          echo "✅ Copied cask file to homebrew-tap branch"
          
          # Copy README (optional, create default if not found)
          if [ -f /tmp/homebrew-cask/README.md ]; then
            cp /tmp/homebrew-cask/README.md README.md
            echo "✅ Copied README.md"
          else
            echo "📝 Creating default README.md for homebrew-tap"
            cat > README.md << 'READMEEOF'
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

## More Information

- [Documentation](https://mindfiredigital.github.io/DeepScanBot/)
- [GitHub Repository](https://github.com/mindfiredigital/DeepScanBot)
- [Report Issues](https://github.com/mindfiredigital/DeepScanBot/issues)
READMEEOF
          fi
          
          # Add and commit
          git add Casks/deepscanbot.rb README.md
          git diff --cached --quiet || git commit -m "chore(homebrew): sync cask to homebrew-tap branch"
          
          # Push to homebrew-tap branch
          git push origin homebrew-tap --force

      - name: Verify cask file
        run: |
          if [ -f Casks/deepscanbot.rb ]; then
            echo "✅ Cask file synced successfully to homebrew-tap branch"
            echo ""
            echo "Cask file contents:"
            cat Casks/deepscanbot.rb
          else
            echo "❌ Error: Cask file not found after sync"
            exit 1
          fi
EOF

echo "✅ Updated .github/workflows/homebrew-sync.yml"

# Step 2: Update .goreleaser.yml
echo ""
echo "📝 Step 2: Updating .goreleaser.yml..."
cat >> .goreleaser.yml << 'EOF'
    repository:
      owner: mindfiredigital
      name: DeepScanBot
      token: "{{ .Env.GITHUB_TOKEN }}"
EOF

echo "✅ Updated .goreleaser.yml"

# Step 3: Update homebrew-tap .gitignore
echo ""
echo "📝 Step 3: Updating homebrew-tap/.gitignore..."
cat > homebrew-tap/.gitignore << 'EOF'
# OS generated files
.DS_Store
Thumbs.db

# IDE files
.vscode/
.idea/
*.swp
*.swo
EOF

echo "✅ Updated homebrew-tap/.gitignore"

# Step 4: Update documentation
echo ""
echo "📝 Step 4: Updating documentation..."
sed -i 's/brew tap mindfiredigital\/tap/brew tap mindfiredigital\/DeepScanBot homebrew-tap/g' apps/docs/docs/installation.mdx

echo "✅ Updated apps/docs/docs/installation.mdx"

echo ""
echo "======================================"
echo "✅ All configuration files updated!"
echo "======================================"
echo ""
echo "Next steps:"
echo "1. Review the changes: git diff"
echo "2. Commit the changes: git add -A && git commit -m 'fix: setup homebrew automation'"
echo "3. Generate cask file: goreleaser release --snapshot --clean --skip=publish --skip=announce"
echo "4. Copy cask: cp dist/homebrew/homebrew-tap/Casks/deepscanbot.rb homebrew-tap/Casks/deepscanbot.rb"
echo "5. Commit cask: git add homebrew-tap/Casks/deepscanbot.rb && git commit -m 'feat: add homebrew cask file'"
echo "6. Create homebrew-tap branch (see README for instructions)"
echo ""
