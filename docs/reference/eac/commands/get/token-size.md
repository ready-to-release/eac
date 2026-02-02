# Get token-size

<!-- book:cmd get token-size -->

## How It Works

Estimates token counts using the **characters/4 heuristic**:

- Reads file content and counts characters
- Divides by 4 to approximate token count
- Works well for code files (actual tokenization varies by model)

This is a fast, zero-dependency estimation - not exact tokenization.

## Common Use Cases

### Find Large Files Before Claude Context

Claude has a ~25,000 token input limit. Find files that exceed it:

```bash
# Find Go files exceeding 25,000 tokens
eac get token-size "go/**/*.go" --threshold 25000

# Find files over 20,000 tokens
eac get token-size "**/*.go" --threshold 20000
```

### Pre-commit Check

Add to CI or pre-commit hooks to catch oversized files:

```bash
# Fail if any source file exceeds limit (exit code 1)
eac get token-size "src/**/*.ts" --threshold 25000
```

### Get All Token Counts

Without `--threshold`, shows all files with their token counts:

```bash
# Show token counts for all Go files
eac get token-size "go/**/*.go"

# JSON output for scripting
eac get token-size "go/core/**/*.go" --as-json
```

## Output Fields (JSON)

| Field | Description |
|-------|-------------|
| `file_path` | Relative path to file |
| `tokens` | Estimated token count (chars/4) |
| `characters` | Character count |
| `bytes` | File size in bytes |
| `lines` | Line count |
| `method` | Estimation method used (`char/4`) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No files exceed threshold (or no threshold set) |
| 1 | One or more files exceed threshold |

## See Also

- [get files](./files.md) - List repository files
- [show files](../show/files.md) - Formatted file listing
