# Release Flow

This document describes the complete release process for DeepScanBot, including version synchronization, building, testing, and publishing.

## Release Checklist

### Pre-Release

- [ ] All tests pass: `go test ./...`
- [ ] Linting passes: `golangci-lint run ./...`
- [ ] Code is formatted: `gofumpt -l -w .`
- [ ] Imports are organized: `gci write --skip-vendor -s standard -s default -s "prefix(github.com/mindfiredigital/DeepScanBot)" .`
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated (if applicable)

### Version Management

1. **Determine version** (follow Semantic Versioning):
   - MAJOR: Breaking changes
   - MINOR: New features, backward compatible
   - PATCH: Bug fixes, backward compatible

2. **Create Git tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

3. **Version synchronization** (automated by GitHub Actions):
   ```
   Git Tag (v1.0.0)
       ↓
   GoReleaser (v1.0.0)
       ↓
   package.json (1.0.0)
       ↓
   GitHub Release (v1.0.0)
       ↓
   npm package (1.0.0)
   ```

### Automated Release (GitHub Actions)

When a Git tag is pushed, the release workflow automatically:

1. **Build Phase**:
   - Checkout repository
   - Setup Go 1.22.4
   - Setup Node.js 20
   - Cache Go modules
   - Install GoReleaser

2. **Binary Build Phase**:
   - Run GoReleaser: `goreleaser release --clean`
   - Builds for all platforms:
     - Linux AMD64
     - Linux ARM64
     - macOS Intel (AMD64)
     - macOS Apple Silicon (ARM64)
     - Windows AMD64
   - Generates SHA256 checksums

3. **Package Preparation Phase**:
   - Copy binaries to npm package: `bash scripts/copy-to-npm.sh`
   - Update package.json version: `node scripts/sync-version.js`
   - Generate checksums: `bash scripts/generate-checksums.sh`
   - Verify binaries: `node scripts/verify-binary.js`

4. **Publish Phase**:
   - Authenticate with npm registry
   - Publish to npm: `npm publish --access public`
   - For prereleases: `npm publish --tag next --access public`

5. **GitHub Release Phase**:
   - Create GitHub Release with tag
   - Upload binary artifacts
   - Include installation instructions
   - Include changelog

### Manual Release (Emergency)

If GitHub Actions fails, you can release manually:

```bash
# 1. Build binaries
goreleaser release --clean

# 2. Copy to npm package
bash scripts/copy-to-npm.sh

# 3. Update version
VERSION=1.0.0 node scripts/sync-version.js

# 4. Generate checksums
bash scripts/generate-checksums.sh

# 5. Verify package
node scripts/prepublish-check.js

# 6. Publish to npm
npm publish

# 7. Create GitHub Release manually via web UI
```

### Snapshot Releases

For testing before a real release:

```bash
# Build snapshot
goreleaser release --snapshot --clean

# Copy binaries
bash scripts/copy-to-npm.sh

# Update version to snapshot
VERSION=0.0.0-snapshot node scripts/sync-version.js

# Dry run publish
npm publish --dry-run

# Or publish to test registry
NPM_REGISTRY=http://localhost:4873/ npm publish
```

## Post-Release

### Verification

After release, verify:

1. **npm package published**:
   ```bash
   npm view @mindfiredigital/deep-scan-bot version
   npm view @mindfiredigital/deep-scan-bot
   ```

2. **Installation works**:
   ```bash
   npm install -g @mindfiredigital/deep-scan-bot
   deepscanbot --version
   deepscanbot -h
   ```

3. **GitHub Release created**:
   - Check https://github.com/mindfiredigital/DeepScanBot/releases
   - Verify binaries are attached
   - Verify checksums are included

4. **Documentation deployed** (if applicable):
   - Check documentation site
   - Verify installation instructions are updated

### Announcement

After successful release:

1. Update project README with new version
2. Post to relevant channels (Twitter, Discord, etc.)
3. Update documentation site
4. Close milestone on GitHub

## Rollback

If a release has issues:

### npm Rollback

```bash
# npm doesn't support deletion, but you can unpublish within 72 hours
npm unpublish @mindfiredigital/deep-scan-bot@1.0.0

# Or deprecate the version
npm deprecate @mindfiredigital/deep-scan-bot@1.0.0 "Critical bug, use 1.0.1"
```

### GitHub Release Rollback

```bash# Delete the release (via web UI or API)
# Delete the tag
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0
```

### Hotfix Release

```bash
# Fix the issue
git commit -m "fix: critical bug in X"

# Create patch version
git tag -a v1.0.1 -m "Hotfix v1.0.1"
git push origin v1.0.1

# Automated release will proceed
```

## Version History

Track all releases in this section:

### v1.0.0 (Planned)
- Initial npm distribution
- Cross-platform binary support
- Automated release workflow
- Comprehensive documentation

## Troubleshooting

### GoReleaser Build Fails

**Check**:
- Go version matches `.goreleaser.yml` requirements
- All Go tests pass
- No compilation errors

**Solution**:
```bash
# Test build locally
go build -o deepscanbot ./apps/cli

# Check GoReleaser config
goreleaser check
```

### npm Publish Fails

**Check**:
- NPM_TOKEN is set in GitHub Secrets
- Version is not already published
- Package name is available

**Solution**:
```bash
# Check npm authentication
npm whoami

# Check package availability
npm view @mindfiredigital/deep-scan-bot

# Verify package.json
cat package.json | jq '.name, .version'
```

### Binary Verification Fails

**Check**:
- Binaries exist in dist/
- Binaries are executable
- Checksums match

**Solution**:
```bash
# List dist/ contents
ls -la dist/

# Verify binaries
node scripts/verify-binary.js

# Regenerate if needed
bash scripts/generate-checksums.sh
```

### postinstall.js Fails

**Check**:
- dist/ directory exists in package
- Binary exists for current platform
- File permissions are correct

**Solution**:
```bash
# Test postinstall manually
node postinstall.js

# Check dist/ structure
ls -la dist/

# Rebuild if needed
goreleaser build --snapshot --clean
bash scripts/copy-to-npm.sh
```

## References

- [Semantic Versioning](https://semver.org/)
- [GoReleaser Documentation](https://goreleaser.com/)
- [npm Publishing Guide](https://docs.npmjs.com/cli/v10/commands/npm-publish)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Git Tagging](https://git-scm.com/book/en/v2/Git-Basics-Tagging)