---
sidebar_position: 2
---

# CLI Flag Naming Conventions

This document defines the standard flag naming conventions for DeepScanBot CLI to ensure consistency across all commands and future development.

## Overview

All DeepScanBot CLI flags follow consistent naming conventions to provide a predictable and intuitive user experience. This document serves as a reference for contributors and maintainers.

## Flag Naming Rules

### 1. Format

- **Long flags only**: All flags use the `--flag-name` format (kebab-case)
- **No short flags**: Short flags (e.g., `-v`) are not used to avoid conflicts and improve readability
- **Consistent casing**: All flags use lowercase with hyphens (kebab-case)

### 2. Naming Patterns by Type

#### Boolean Flags

Boolean flags should use descriptive adjectives or verbs without a `--enable-` or `--disable-` prefix:

**Good:**
- `--json` (output in JSON format)
- `--dry-run` (preview without executing)
- `--force` (force overwrite)
- `--yes` (auto-confirm)
- `--no-input` (disable interactive prompts)
- `--unique` (deduplicate URLs)
- `--insecure` (skip TLS validation)
- `--sitemap` (discover from sitemap)
- `--resume` (resume previous crawl)
- `--ignore-robots` (ignore robots.txt)
- `--cross-domain` (follow cross-domain links)
- `--show-source` (include source URLs)
- `--disable-redirects` (disable redirects)

**Avoid:**
- `--enable-json` (redundant)
- `--do-resume` (unnecessary verb)

#### String Flags

String flags should use descriptive nouns:

- `--output` (output file base name)
- `--proxy` (HTTP proxy URL)
- `--content-types` (content types to accept)

#### Integer Flags

Integer flags should use descriptive nouns:

- `--depth` (crawl depth)
- `--timeout` (timeout in seconds)
- `--concurrency` (concurrent requests)
- `--host-concurrency` (concurrent requests per host)
- `--retries` (number of retries)
- `--size` (maximum page size in bytes)

#### Duration Flags

Duration flags should use descriptive nouns with clear units in the description:

- `--delay` (delay between requests)
- `--retry-backoff` (backoff duration for retries)

### 3. Flag Precedence

When adding new flags, follow this precedence order:

1. **Global flags** (available to all commands):
   - `--json`
   - `--no-input`
   - `--dry-run`

2. **Command-specific flags** (only for relevant commands):
   - Scan command: `--depth`, `--timeout`, `--concurrency`, etc.
   - Version command: `--json` (inherited from global)
   - Doctor command: `--json` (inherited from global)

### 4. Default Values

- Always provide sensible defaults for all flags
- Document defaults in the flag description
- Use common values that work for most use cases

**Example:**
```go
scanCmd.Flags().Int("depth", 2, "Maximum crawl depth (default: 2)")
scanCmd.Flags().Int("concurrency", 8, "Maximum concurrent requests (default: 8)")
```

### 5. Flag Descriptions

Flag descriptions should:
- Be concise (under 80 characters when possible)
- Start with a verb or adjective (for booleans) or noun (for values)
- Include units or constraints in parentheses
- Mention the default value if not obvious

**Good:**
- `"Maximum crawl depth (default: 2)"`
- `"HTTP proxy URL (e.g., http://127.0.0.1:8080)"`
- `"Delay between requests to the same host (e.g., 500ms, 1s)"`

**Avoid:**
- `"This flag sets the depth"` (too verbose)
- `"depth"` (too terse)

### 6. Backward Compatibility

When modifying existing flags:

1. **Never rename** existing flags without providing aliases
2. **Never remove** flags without deprecation warnings
3. **Add new flags** with clear documentation
4. **Support both formats** when transitioning (flags and key=value)

**Example of backward-compatible change:**
```go
// Old key=value format still works
deepscanbot scan https://example.com depth=3

// New flag format also works
deepscanbot scan https://example.com --depth=3

// Both can be mixed (flags take precedence)
deepscanbot scan https://example.com depth=2 --depth=3  # Uses --depth=3
```

## Standard Flag Reference

### Global Flags (All Commands)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output results in JSON format |
| `--no-input` | bool | `false` | Disable all interactive prompts; fail if required input is missing |
| `--dry-run` | bool | `false` | Preview actions that would be performed without making changes |
| `--quiet` | bool | `false` | Suppress non-essential output (only show warnings and errors) |
| `--verbose` | bool | `false` | Display additional informational messages |
| `--debug` | bool | `false` | Display detailed debugging information |
| `--help` / `-h` | bool | `false` | Show help for any command |

### Scan Command Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--depth` | int | `2` | Maximum crawl depth |
| `--timeout` | int | `2` | Request timeout in seconds |
| `--concurrency` | int | `8` | Maximum concurrent requests |
| `--host-concurrency` | int | `2` | Maximum concurrent requests per host |
| `--output` | string | `crawler_results` | Output file base name (without extension) |
| `--proxy` | string | `""` | HTTP proxy URL (e.g., http://127.0.0.1:8080) |
| `--delay` | duration | `0` | Delay between requests to the same host (e.g., 500ms, 1s) |
| `--retries` | int | `0` | Number of retries for failed requests |
| `--retry-backoff` | duration | `1s` | Initial backoff duration for retries (e.g., 1s, 500ms) |
| `--content-types` | string | `text/html` | Content types to accept (comma-separated) |
| `--size` | int | `-1` | Maximum page size in bytes; -1 for unlimited |
| `--ignore-robots` | bool | `false` | Ignore robots.txt rules |
| `--cross-domain` | bool | `false` | Follow links to different domains |
| `--sitemap` | bool | `false` | Discover URLs from sitemap.xml |
| `--resume` | bool | `false` | Resume from previous crawl results |
| `--insecure` | bool | `false` | Skip TLS certificate validation |
| `--unique` | bool | `false` | Only process unique URLs (deduplicate) |
| `--show-source` | bool | `false` | Include source URL in output for discovered links |
| `--disable-redirects` | bool | `false` | Disable following HTTP redirects |
| `--force` | bool | `false` | Overwrite existing output file without prompting |
| `--yes` | bool | `false` | Auto-confirm all destructive operations (e.g. overwriting files) |

## Implementation Guidelines

### Adding a New Flag

1. **Choose the right type**: bool, string, int, or duration
2. **Follow naming conventions**: Use kebab-case, descriptive names
3. **Set a sensible default**: Choose a value that works for most users
4. **Write a clear description**: Include units, examples, and default value
5. **Add to flag registration**: Register in the `init()` function
6. **Update documentation**: Add to usage guide and this document
7. **Add tests**: Verify the flag works correctly

**Example:**
```go
// In init()
scanCmd.Flags().Int("max-errors", 10, "Maximum errors before stopping crawl (default: 10)")

// In command Run function
maxErrors := cmd.Flags().GetInt("max-errors")
```

### Modifying an Existing Flag

1. **Check for dependencies**: Ensure no breaking changes
2. **Update default if needed**: Document the change in release notes
3. **Maintain backward compatibility**: Support old behavior if possible
4. **Update tests**: Ensure existing tests still pass
5. **Update documentation**: Reflect the change in all docs

### Deprecating a Flag

1. **Add deprecation notice**: Update flag description
2. **Maintain functionality**: Keep the flag working
3. **Suggest alternative**: Point users to the new flag
4. **Set removal timeline**: Document when it will be removed
5. **Update tests**: Add deprecation warning tests

**Example:**
```go
scanCmd.Flags().String("old-flag", "", "DEPRECATED: Use --new-flag instead (will be removed in v2.0)")
```

## Testing Requirements

All flags must have tests that verify:

1. **Default values**: Flag works without being specified
2. **Explicit values**: Flag works when explicitly set
3. **Invalid values**: Flag handles invalid input gracefully
4. **Help output**: Flag appears in `--help` output
5. **JSON output**: Flag appears in `--help --json` output
6. **Backward compatibility**: Old key=value format still works (if applicable)

## Code Review Checklist

When reviewing PRs that add or modify flags:

- [ ] Flag name follows kebab-case convention
- [ ] Flag type is appropriate (bool, string, int, duration)
- [ ] Default value is sensible and documented
- [ ] Description is clear and concise
- [ ] Flag is added to the correct command (global vs. command-specific)
- [ ] Documentation is updated (usage guide, this document)
- [ ] Tests are added or updated
- [ ] Backward compatibility is maintained
- [ ] No duplicate flags exist
- [ ] Flag appears in help output correctly

## Migration Guide for Contributors

### Converting Key=Value to Flags

If you're adding a new option that was previously only available via key=value:

1. **Add the flag** with the same name as the key=value option
2. **Support both formats** in the parser
3. **Flags take precedence** over key=value options
4. **Document both formats** in examples
5. **Update tests** to cover both formats

**Example:**
```go
// Add flag
scanCmd.Flags().Bool("unique", false, "Only process unique URLs (deduplicate)")

// In parseKeyValue(), the "unique" key is already supported
case "unique":
    opts.Unique = val == "true"

// In mergeOptions(), flag takes precedence
if !cmd.Flags().Lookup("unique").Changed {
    opts.Unique = kvOpts.Unique
}
```

## Questions?

If you have questions about flag naming or need to propose a new convention, please open an issue or discussion on GitHub.