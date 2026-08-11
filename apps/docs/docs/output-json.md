---
sidebar_position: 4
---

# Output & JSON

DeepScanBot supports multiple output formats for different use cases.

## Output Formats

### Text Output (Default)

The default text output provides a human-readable report:

```
https://example.com [status=200] [result=passed]
https://example.com/about [status=200] [result=passed]
https://example.com/contact [result=discovered]
https://example.com/admin [result=skipped] [skipped=disallowed by robots.txt]
```

**File extension:** `.txt`

### JSON Output

JSON output is ideal for automation, scripting, and integration with other tools.

```bash
deepscanbot scan https://example.com --json
```

**File extension:** `.json`

> **Note:** The output filename is specified without extension. DeepScanBot automatically appends the appropriate extension based on the output format:
> - Text format (default): `crawler_results.txt`
> - JSON format (`--json` or `json=true`): `crawler_results.json`

## JSON Output Structure

All JSON responses follow this structure:

```json
{
  "status": "success | error",
  "data": { ... },
  "error": {
    "message": "Error description",
    "code": "error_code"
  },
  "meta": {
    "timestamp": "ISO8601 timestamp",
    "command": "command_name",
    "duration_ms": 1234
  }
}
```

## Piping and Automation

```bash
# Parse with jq
deepscanbot scan https://example.com --json | jq '.data.summary'

# Extract specific fields
deepscanbot --version --json | jq -r '.data.version'

# Check command success
if deepscanbot scan https://example.com --json | jq -e '.status == "success"'; then
  echo "Scan completed successfully"
fi
```

> **Note:** `json=true` is a scan option that enables JSON output for the scan command. The `--json` flag is a global CLI option that works with any command (such as `scan`, `version`, `doctor`, etc.) to produce JSON-formatted output.
