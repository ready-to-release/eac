# Get files-by-module

<!-- book:cmd get files-by-module -->

Returns files owned by a specific module from the `FILES_BY_MODULE` JSON data (produced by `get changed-modules-ci`). Parses the JSON mapping and extracts files for the requested module.

## Usage

```bash
eac get files-by-module <module> [--count] [--json <json>] [--format shell]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `module` | Yes | Module name to get files for |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--count` | bool | Output only the file count |
| `--json` | string | Explicit FILES_BY_MODULE JSON (defaults to `FILES_BY_MODULE` env var) |
| `--format` | string | `shell` outputs `COUNT` and `FILES` variables; `list` outputs one file per line (default) |

## Output Formats

**Default (list):** One file path per line.

**Count:** Just the number of files.

**Shell:**
```
COUNT="5"
FILES="file1.go file2.go file3.go file4.go file5.go"
```

## Examples

```bash
# List files for a module
eac get files-by-module docs

# Just the count
eac get files-by-module docs --count

# From explicit JSON
eac get files-by-module docs --json '{"docs":["docs/index.md","docs/guide.md"]}'

# Shell format for eval
eval $(eac get files-by-module docs --format shell)
echo "$COUNT files changed"
```

## See Also

- [get](get.md)
- [get files](files.md)
- [show files](../show/files.md)
