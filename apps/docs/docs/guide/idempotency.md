---
sidebar_position: 4
---

# Idempotency and Retry Safety

DeepScanBot is designed to be safe for retries and automation. This document explains which operations are idempotent, how retries are handled, and what to expect when running commands multiple times.

## What is Idempotency?

An operation is idempotent if executing it multiple times produces the same result as executing it once. This matters for:

- CI/CD pipelines that automatically retry failed jobs
- Automation tools that may retry on transient failures
- AI agents that need predictable behavior
- Scripts that implement retry logic

## Idempotent Operations

### Scan Command

The scan command is mostly idempotent:

**Idempotent aspects:**

- Crawling behavior - running the same scan multiple times produces the same crawl results
- File writing - output files are written atomically using temp file plus rename, preventing corruption if interrupted
- Duplicate URL detection - the crawler automatically deduplicates URLs, each URL is only crawled once per session
- Resume mode - the `--resume` flag enables safe retries by loading existing results and skipping already-crawled URLs

**Non-idempotent aspects:**

- Output file overwriting - by default, running the same scan overwrites the output file. Use `--resume` to preserve previous results or `--force` to explicitly allow overwrites
- Timestamps - each scan generates new timestamps, which is expected behavior
- External side effects - web servers may log each request and rate limiting may be triggered, these are external to DeepScanBot

### Version Command

The version command is fully idempotent - it always returns the same version information with no side effects.

### Doctor Command

The doctor command is fully idempotent - it performs read-only checks with no side effects.

## Retry Safety

### Automatic Retries

DeepScanBot includes built-in retry logic for transient failures:

```bash
deepscanbot scan https://example.com --retries=3 --retry-backoff=2s
```

Retry behavior:
- Failed requests are retried up to N times (configurable via `--retries`)
- Backoff duration doubles with each retry (configurable via `--retry-backoff`)
- Only transient errors are retried (network errors, 5xx status codes)
- Permanent errors (404, invalid URLs) are not retried

### Safe Retry Patterns

**Simple retry:**

```bash
for i in {1..3}; do
    if deepscanbot scan https://example.com --output=results; then
        echo "Scan succeeded"
        exit 0
    fi
    echo "Attempt $i failed, retrying..."
    sleep 5
done
echo "Scan failed after 3 attempts"
exit 1
```

**Retry with resume:**

```bash
deepscanbot scan https://example.com --output=results --resume --retries=3
```

This pattern uses `--resume` to load previous results and only crawls URLs that haven't been crawled yet.

**CI/CD with timeout:**

```yaml
- name: Run DeepScanBot
  run: |
    deepscanbot scan https://example.com \
      --json \
      --output=results \
      --timeout=5m \
      --retries=3 \
      --retry-backoff=2s \
      --no-input \
      --force
  timeout-minutes: 10
```

## Handling Interrupted Scans

If a scan is interrupted (Ctrl+C, timeout, system crash):

**Without `--resume`**: Running the same command again will start from scratch. All URLs will be re-crawled and the output file will be overwritten. Safe but inefficient.

**With `--resume`**: Running the same command again will continue from where it left off. It loads existing results from the output file and only crawls URLs that haven't been crawled yet. More efficient for large crawls.

Example:

```bash
# Start a long crawl
deepscanbot scan https://example.com --depth=5 --output=results --timeout=10m

# If interrupted, resume with:
deepscanbot scan https://example.com --depth=5 --output=results --resume --timeout=10m
```

## Preventing Duplicate Resources

### Output File Conflicts

By default, DeepScanBot prevents accidental overwrites:

```bash
# First run - creates output file
deepscanbot scan https://example.com --output=results

# Second run - fails with error (in non-interactive mode)
deepscanbot scan https://example.com --output=results --no-input

# Second run - succeeds with explicit overwrite
deepscanbot scan https://example.com --output=results --force
```

### Duplicate URL Prevention

The crawler automatically prevents duplicate URL processing. Each URL is only crawled once per session, preventing duplicate network requests and duplicate entries in output.

## Atomic File Operations

All file writes use atomic operations to prevent corruption. The process writes to a temporary file in the same directory, syncs to disk, then atomically renames the temp file to the target file. This ensures the file is either completely written or not modified, with no partial or corrupted files if interrupted.

## Known Limitations

### External State Changes

DeepScanBot cannot control external side effects. Web servers may log each request, rate limiting may be triggered, and cached content may change between runs. Use `--resume` to minimize repeated requests.

### Non-Deterministic Crawling

Concurrent crawling may produce results in different orders. URLs may be discovered in different orders and results may be written in different orders. This does not affect correctness. Sort results if order matters for your use case.

### Time-Based Content

Web content may change between runs. Pages may be updated and links may be added or removed. This is expected behavior for a web crawler. Use `--resume` with consistent output files to track changes over time.

## Best Practices

### For CI/CD

Always use these flags in CI/CD:
- `--no-input` - fail instead of prompting
- `--force` - allow overwriting output
- `--timeout` - set reasonable timeout
- `--retries` - enable automatic retries
- `--json` - machine-readable output

### For Automation

Use resume mode for repeated runs:

```bash
deepscanbot scan "$URL" \
    --output="results/$(echo $URL | md5sum | cut -d' ' -f1)" \
    --resume \
    --timeout=30s \
    --retries=2 \
    --no-input \
    --force
```

## Summary

| Operation | Idempotent? | Safe to Retry? | Notes |
|-----------|-------------|----------------|-------|
| `scan` (without resume) | Partial | Yes | Overwrites output, but safe |
| `scan` (with `--resume`) | Yes | Yes | Preserves previous results |
| `scan` (with `--force`) | Yes | Yes | Explicitly allows overwrites |
| `version` | Yes | Yes | No side effects |
| `doctor` | Yes | Yes | Read-only checks |

## Questions?

If you have questions about idempotency or need help with retry logic, please open an issue or discussion on GitHub.