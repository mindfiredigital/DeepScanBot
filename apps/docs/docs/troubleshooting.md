---
sidebar_position: 6
---

# Troubleshooting

This page provides solutions for common issues encountered when using DeepScanBot.

## Installation Issues

### Problem: `npm install -g @mindfiredigital/deepscanbot` fails

**Solutions:**

- Ensure you have Node.js 18+ installed: `node --version`
- Check npm permissions: `npm config get prefix`
- Try with sudo (Unix): `sudo npm install -g @mindfiredigital/deepscanbot`
- Clear npm cache: `npm cache clean --force`

### Problem: "Unsupported platform" error during installation

**Solutions:**

- Verify your OS and architecture: `node -e "console.log(process.platform, process.arch)"`
- Supported platforms: macOS (x64, arm64), Linux (x64, arm64), Windows (x64)
- Ensure you're using a 64-bit version of Node.js

### Problem: "Command not found" after installation

**Solutions:**

- Check npm global bin directory is in your PATH: `npm bin -g`
- Add to PATH: `export PATH=$(npm bin -g):$PATH`
- On Windows, restart your terminal after installation

## Crawling Issues

### Problem: No results or empty output

**Solutions:**

- Verify the URL is accessible: `curl -I https://example.com`
- Increase timeout: `deepscanbot scan <url> timeout=10`
- Disable TLS verification for testing: `deepscanbot scan <url> insecure=true`
- Check if robots.txt is blocking: `deepscanbot scan <url> ignore-robots=true`

### Problem: Too many requests or being blocked

**Solutions:**

- Reduce concurrency: `deepscanbot scan <url> concurrency=2`
- Add crawl delay: `deepscanbot scan <url> delay=1s`
- Use a proxy: `deepscanbot scan <url> proxy=http://proxy.example.com:8080`

### Problem: Timeouts

**Solutions:**

- Increase timeout value: `deepscanbot scan <url> timeout=30`
- For very slow sites, use: `deepscanbot scan <url> timeout=60`
- Reduce concurrency to avoid overwhelming the server: `deepscanbot scan <url> concurrency=2`

### Problem: Website blocking

**Solutions:**

- Respect robots.txt (default behavior)
- Add delays between requests: `deepscanbot scan <url> delay=1s`
- Reduce concurrency: `deepscanbot scan <url> concurrency=2`
- Use a proxy if needed: `deepscanbot scan <url> proxy=http://proxy.example.com:8080`

## Getting Help

If you're still experiencing issues:

1. Check the [Usage Guide](/docs/guide/usage) for detailed examples
2. Run `deepscanbot doctor` to verify your installation
3. Enable verbose output: `deepscanbot scan <url> --verbose`
4. Check the [GitHub Issues](https://github.com/mindfiredigital/DeepScanBot/issues) for similar problems
