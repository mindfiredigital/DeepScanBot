---
sidebar_position: 4
---

# Idempotency and Retry Safety

DeepScanBot is designed to be safe for retries and automation. This document explains which operations are idempotent, how retries are handled, and what to expect when running commands multiple times.

## What is Idempotency?

An operation is **idempotent** if executing it multiple times produces the same result as executing it once. This is critical for:

- **CI/CD pipelines** that automatically retry failed jobs
- **Automation tools** that may retry on transient failures
- **AI agents** that need predictable behavior
- **Scripts** that implement retry logic

## Idempotent Operations

### Scan Command

The `scan` command is **mostly idempotent** with the following characteristics:

#### Idempotent Aspects

1. **Crawling Behavior**: Running the same scan multiple times produces the same crawl results
   - Same URLs are discovered
   - Same content is fetched
   - Same results are generated

2. **File Writing**: Output files are written atomically
   - Uses atomic writes (temp file + rename)
   - Prevents corruption if interrupted
   - Safe to retry after failure

3. **Duplicate URL Detection**: The crawler automatically deduplicates URLs
   - Each URL is only crawled once per session
   - Prevents duplicate work on retry

4. **Resume Mode**: The `--resume` flag enables safe retries
   - Loads existing results from output file
   - Skips already-crawled URLs
   - Continues from where it left off

#### Non-Idempotent Aspects

1. **Output File Overwriting**: By default, running the same scan overwrites the output file
   - Use `--resume` to preserve previous results
   - Use `--force` to explicitly allow overwrites
   - Use `--dry-run` to preview without making changes

2. **Timestamps**: Each scan generates new timestamps
   - `started_at` and `finished_at` reflect the current run
   - This is expected and correct behavior

3. **External Side Effects**: The scan may trigger external changes
   - Web servers may log each request
   - Rate limiting may be triggered
   - These are external to DeepScanBot and cannot be controlled

### Version Command

The `version` command is **fully idempotent**:
- Always returns the same version information
- No side effects
- Safe to run multiple times

### Doctor Command

The `doctor` command is **fully idempotent**:
- Performs read-only checks
- No side effects
- Safe to run multiple times

## Retry Safety

### Automatic Retries

DeepScanBot includes built-in retry logic for transient failures:

```bash
# Retry failed requests up to 3 times with exponential backoff
deepscanbot scan https://example.com --retries=3 --retry-backoff=2s
```

**Retry Behavior:**
- Failed requests are retried up to N times (configurable via `--retries`)
- Backoff duration doubles with each retry (configurable via `--retry-backoff`)
- Only transient errors are retried (network errors, 5xx status codes)
- Permanent errors (404, invalid URLs) are not retried

### Safe Retry Patterns

#### Pattern 1: Simple Retry

```bash
#!/bin/bash
# Retry scan up to 3 times on failure

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

#### Pattern 2: Retry with Resume

```bash
#!/bin/bash
# Retry with resume mode to avoid re-crawling

deepscanbot scan https://example.com --output=results --resume --retries=3
```

This pattern:
- Uses `--resume` to load previous results
- Only crawls URLs that haven't been crawled yet
- Safe to run multiple times without duplicate work

#### Pattern 3: CI/CD with Timeout

```yaml
# GitHub Actions example
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

#### Pattern 4: Idempotent Batch Processing

```bash
#!/bin/bash
# Process multiple URLs idempotently

INPUT_FILE="urls.txt"
OUTPUT_DIR="results"
PROCESSED_FILE=".processed_urls"

mkdir -p "$OUTPUT_DIR"

while IFS= read -r url; do
    # Create a unique output file for each URL
    url_hash=$(echo "$url" | md5sum | cut -d' ' -f1)
    output_file="$OUTPUT_DIR/$url_hash.json"
    
    # Only scan if not already processed
    if [ ! -f "$output_file" ]; then
        echo "Scanning: $url"
        deepscanbot scan "$url" \
            --json \
            --output="$OUTPUT_DIR/$url_hash" \
            --timeout=30s \
            --retries=2 || {
            echo "Failed to scan: $url"
            continue
        }
    else
        echo "Already scanned: $url"
    fi
done < "$INPUT_FILE"
```

## Handling Interrupted Scans

### Scenario: Scan Interrupted Mid-Crawl

If a scan is interrupted (Ctrl+C, timeout, system crash):

1. **Without `--resume`**: Running the same command again will start from scratch
   - All URLs will be re-crawled
   - Output file will be overwritten
   - Safe but inefficient

2. **With `--resume`**: Running the same command again will continue from where it left off
   - Loads existing results from output file
   - Only crawls URLs that haven't been crawled yet
   - More efficient for large crawls

### Example: Recovering from Interruption

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
# Error: Output file "results.txt" already exists. Refusing to overwrite without confirmation.

# Second run - succeeds with explicit overwrite
deepscanbot scan https://example.com --output=results --force

# Second run - succeeds with auto-confirm
deepscanbot scan https://example.com --output=results --yes
```

### Duplicate URL Prevention

The crawler automatically prevents duplicate URL processing:

```go
// In-memory deduplication during a single crawl session
if !c.pageStorage.MarkVisitedIfNew(targetURL) {
    c.storeSkipped(targetURL, source, depth, "duplicate")
    return
}
```

This ensures:
- Each URL is only crawled once per session
- No duplicate network requests
- No duplicate entries in output

## Atomic File Operations

All file writes use atomic operations to prevent corruption:

### How It Works

1. Write to a temporary file in the same directory
2. Sync to disk
3. Atomically rename temp file to target file

This ensures:
- File is either completely written or not modified
- No partial/corrupted files if interrupted
- Safe to retry after failure

### Example

```go
// WriteJSONReportToFile uses atomic writes
func WriteJSONReportToFile(filename string, report CrawlReport) error {
    jsonData, err := json.MarshalIndent(report, "", "  ")
    if err != nil {
        return err
    }
    
    // Atomic write prevents corruption
    return WriteFileAtomic(filename, jsonData, 0o644)
}
```

## Known Limitations

### 1. External State Changes

DeepScanBot cannot control external side effects:
- Web servers may log each request
- Rate limiting may be triggered
- Cached content may change between runs

**Mitigation**: Use `--resume` to minimize repeated requests.

### 2. Non-Deterministic Crawling

Concurrent crawling may produce results in different orders:
- URLs may be discovered in different orders
- Results may be written in different orders
- This does not affect correctness

**Mitigation**: Sort results if order matters for your use case.

### 3. Time-Based Content

Web content may change between runs:
- Pages may be updated
- Links may be added/removed
- This is expected behavior for a web crawler

**Mitigation**: Use `--resume` with consistent output files to track changes over time.

## Testing Idempotency

### Test 1: Repeated Execution

```bash
# Run scan twice with same parameters
deepscanbot scan https://example.com --depth=2 --output=test1 --force
deepscanbot scan https://example.com --depth=2 --output=test1 --force

# Results should be identical (except timestamps)
diff <(jq 'del(.started_at, .finished_at)' test1.json) \
     <(jq 'del(.started_at, .finished_at)' test1.json)
```

### Test 2: Resume Mode

```bash
# Start a scan
deepscanbot scan https://example.com --depth=3 --output=test2

# Interrupt it (Ctrl+C)

# Resume the scan
deepscanbot scan https://example.com --depth=3 --output=test2 --resume

# Should complete without re-crawling
```

### Test 3: Retry with Transient Failure

```bash
# Start a server that fails the first 2 requests
# Then succeeds on the 3rd request

# Scan with retries
deepscanbot scan http://localhost:8080 --retries=3 --retry-backoff=1s

# Should succeed after retries
```

## Best Practices

### For CI/CD

```yaml
# Always use these flags in CI/CD
- --no-input: Fail instead of prompting
- --force: Allow overwriting output
- --timeout: Set reasonable timeout
- --retries: Enable automatic retries
- --json: Machine-readable output
```

### For Automation

```bash
#!/bin/bash
set -euo pipefail

# Use resume mode for repeated runs
deepscanbot scan "$URL" \
    --output="results/$(echo $URL | md5sum | cut -d' ' -f1)" \
    --resume \
    --timeout=30s \
    --retries=2 \
    --no-input \
    --force
```

### For Scripts

```bash
#!/bin/bash
# Check if output already exists and is recent

OUTPUT="results.json"
MAX_AGE_HOURS=24

if [ -f "$OUTPUT" ]; then
    AGE_HOURS=$(( $(date +%s) - $(stat -c %Y "$OUTPUT") ))
    if [ $AGE_HOURS -lt $((MAX_AGE_HOURS * 3600)) ]; then
        echo "Output is recent, skipping scan"
        exit 0
    fi
fi

deepscanbot scan "$URL" --output=results --force
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