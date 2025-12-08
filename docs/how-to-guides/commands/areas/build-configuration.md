<!-- EDITOR
# Editor: how-to-guides/commands/areas/build-configuration.md

## Soul

Configuration reference for build system covering module types, build flags, version injection, compression settings, output directories, and CI/CD integration patterns.

## Sections

1. Module Contract Configuration
   - Module Type Declaration
   - Supported Module Types
2. Build Flags
   - Go Build Flags
   - Version Injection
   - Compression Settings
3. Output Configuration
   - Build Output Directory
   - Platform-Specific Outputs
   - MkDocs Output
4. Environment Variables
5. Module Type Configurations
   - go-cli Configuration
   - mkdocs-site Configuration
   - containers Configuration
6. CI/CD Configuration
   - GitHub Actions Build
   - Release Build Configuration
7. Build Hooks
   - Pre-build Hook
   - Post-build Hook
8. Parallel Build Settings
   - Default Behavior
   - Controlling Parallelism
9. Troubleshooting
10. Related Documentation
-->

# Build Configuration

This guide covers configuration options for EAC's build system, including module types, build flags, and output settings.

## Module Contract Configuration

### Module Type Declaration

```yaml
# .r2r/eac/modules.yml
modules:
  - moniker: r2r-cli
    type: go-cli
    description: R2R command-line interface
    files:
      root: scripts/r2r-cli
      patterns:
        - "**/*.go"
        - "go.mod"
        - "go.sum"
```

### Supported Module Types

| Type | Description | Build Output |
|------|-------------|--------------|
| `go-cli` | Go CLI application | Multi-platform binaries |
| `go-commands` | Go command package | Library (no binary) |
| `go-library` | Go library | Library (no binary) |
| `go-mcp` | Go MCP server | Server binary |
| `mkdocs-site` | MkDocs documentation | Static HTML site |
| `mkdocs-book` | MkDocs book (PDF/EPUB) | PDF and EPUB files |
| `containers` | Docker container | Container image |
| `specifications` | Gherkin specs | Validated specs |
| `contracts` | Configuration contracts | Validated contracts |
| `markdown` | Markdown files | Validated markdown |

## Build Flags

### Go Build Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--tidy-first` | Run `go mod tidy` before build | Local: true, CI: false |
| `--no-tidy` | Skip `go mod tidy` | false |
| `--skip-deps` | Skip dependency verification | false |
| `--compressed` | Strip debug symbols (`-ldflags="-s -w"`) | false |
| `--compressed-upx` | Strip + UPX compression | false |
| `--version <string>` | Inject version via ldflags | none |

### Version Injection

Version strings are injected using Go ldflags:

```bash
# Injects: -ldflags="-X main.version=v1.2.0"
r2r eac build --version v1.2.0 r2r-cli
```

Access in Go code:

```go
package main

var version = "dev" // Replaced at build time

func main() {
    fmt.Printf("Version: %s\n", version)
}
```

### Compression Settings

**Standard compression (`--compressed`):**

```bash
r2r eac build --compressed r2r-cli
# Uses: -ldflags="-s -w"
# Effect: Strips debug symbols
# Size reduction: ~30%
```

**UPX compression (`--compressed-upx`):**

```bash
r2r eac build --compressed-upx r2r-cli
# Uses: -ldflags="-s -w" + UPX
# Effect: Strips + executable compression
# Size reduction: ~60-70%
# Requires: UPX installed
```

## Output Configuration

### Build Output Directory

```yaml
# Default output structure
out/
└── build/
    ├── <module>/
    │   ├── build.log          # Build output log
    │   └── <artifacts>        # Module-specific outputs
    └── orchestrator.log       # Multi-module build log
```

### Platform-Specific Outputs

For `go-cli` modules:

```text
out/build/r2r-cli/
├── build.log
├── r2r-linux-amd64
├── r2r-linux-arm64
├── r2r-windows-amd64.exe
├── r2r-darwin-amd64
└── r2r-darwin-arm64
```

### MkDocs Output

```text
out/build/docs/
├── build.log
└── site/
    ├── index.html
    ├── assets/
    └── ...
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GOOS` | Target operating system | Current OS |
| `GOARCH` | Target architecture | Current arch |
| `CGO_ENABLED` | Enable CGO | 0 (disabled) |
| `R2R_BUILD_DIR` | Build output directory | `out/build/` |

## Module Type Configurations

### go-cli Configuration

```yaml
modules:
  - moniker: r2r-cli
    type: go-cli
    files:
      root: scripts/r2r-cli
    build:
      platforms:
        - linux/amd64
        - linux/arm64
        - windows/amd64
        - darwin/amd64
        - darwin/arm64
```

### mkdocs-site Configuration

```yaml
modules:
  - moniker: docs
    type: mkdocs-site
    files:
      root: docs
      patterns:
        - "**/*.md"
        - "mkdocs.yml"
```

### containers Configuration

```yaml
modules:
  - moniker: ext-eac
    type: containers
    files:
      root: containers/ext-eac
      patterns:
        - "Dockerfile"
        - "**/*"
    container:
      registry: ghcr.io
      repository: ready-to-release/ext-eac
```

## CI/CD Configuration

### GitHub Actions Build

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Build all modules
        run: r2r eac build --no-tidy

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: binaries
          path: out/build/
```

### Release Build Configuration

```yaml
name: Release

on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5

      - name: Install UPX
        run: sudo apt-get install -y upx

      - name: Build release
        run: |
          r2r eac build \
            --version ${{ github.ref_name }} \
            --compressed-upx \
            r2r-cli

      - name: Upload release
        uses: softprops/action-gh-release@v1
        with:
          files: out/build/r2r-cli/*
```

## Build Hooks

### Pre-build Hook

```yaml
# Module with pre-build hook
modules:
  - moniker: eac-commands
    type: go-commands
    hooks:
      pre_build:
        - "go generate ./..."
```

### Post-build Hook

```yaml
modules:
  - moniker: r2r-cli
    type: go-cli
    hooks:
      post_build:
        - "sha256sum out/build/r2r-cli/* > checksums.txt"
```

## Parallel Build Settings

### Default Behavior

- Single module: Verbose, sequential output
- Multiple modules: Parallel execution, summarized output

### Controlling Parallelism

```bash
# Build specific modules in parallel
r2r eac build eac-commands eac-core r2r-cli

# Build in dependency order
r2r eac get execution order r2r-cli | xargs r2r eac build
```

## Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| Build hangs | Large module | Check `build.log` for progress |
| Missing binary | Wrong module type | Verify `go-cli` type in contract |
| UPX fails | UPX not installed | Install UPX or use `--compressed` |
| Version not injected | Wrong variable name | Use `var version = "dev"` |
| Cross-compile fails | CGO dependency | Set `CGO_ENABLED=0` |

## Related Documentation

- [Build Overview](build-overview.md) - Build system concepts
- [Build Commands](build-commands.md) - Command reference
- [Test Configuration](test-configuration.md) - Test after build
