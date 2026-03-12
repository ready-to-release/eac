# Get cli-release-notes

<!-- book:cmd get cli-release-notes -->

Generates release notes markdown for CLI binary releases. Includes installation instructions with platform-specific download links, binary sizes, UPX-compressed variants, and supply chain security verification commands.

## Usage

```bash
eac get cli-release-notes --version <version> --tag <tag> --commit <sha> [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--version` | string | | Version string (required) |
| `--tag` | string | | Git tag name (required) |
| `--commit` | string | | Git commit SHA (required) |
| `--module` | string | `clie` | Module name |
| `--binary-prefix` | string | `clie` | Binary name prefix |
| `--repo` | string | | GitHub repository (owner/repo) |
| `--run-id` | string | | GitHub Actions run ID |

## Output

Outputs raw markdown to stdout. Binary sizes are read from `out/build/{module}/`.

Platforms included: Linux AMD64/ARM64, macOS Intel/Apple Silicon, Windows AMD64, plus UPX-compressed variants for Linux and Windows.

## Examples

```bash
eac get cli-release-notes --version 1.0.0 --tag clie/1.0.0 --commit abc123

eac get cli-release-notes --version 2.0.0 --tag clie/2.0.0 --commit abc123 \
  --repo org/repo --run-id 12345678
```

## See Also

- [get](get.md)
- [release](../release/index.md)
