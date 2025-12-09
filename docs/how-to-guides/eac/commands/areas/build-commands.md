# Build Commands

{{ page_breadcrumb() }}

Command reference for EAC's build system.

## Quick Reference

| Command | Description                          |
| ------- | ------------------------------------ |
| `build` | Build one or more modules by moniker |

---

## build

Build one or more modules by moniker.

### Synopsis

```bash
r2r eac build [module1] [module2] ... [options]
```

### Description

Builds modules based on their type defined in module contracts. Automatically dispatches to the correct build tooling (Go, MkDocs, Docker, etc.).

### Arguments

| Argument    | Required | Description                                |
| ----------- | -------- | ------------------------------------------ |
| `module...` | No       | Module monikers to build (defaults to all) |

### Flags

| Flag               | Short | Type   | Default     | Description                           |
| ------------------ | ----- | ------ | ----------- | ------------------------------------- |
| `--tidy-first`     |       | bool   | local: true | Run `go mod tidy` before building     |
| `--no-tidy`        |       | bool   | CI: true    | Skip `go mod tidy`                    |
| `--skip-deps`      |       | bool   | false       | Skip build dependency verification    |
| `--compressed`     |       | bool   | false       | Strip debug info for smaller binaries |
| `--compressed-upx` |       | bool   | false       | Apply UPX compression (requires UPX)  |
| `--version`        | `-v`  | string | -           | Inject version string into binary     |

### Examples

```bash
# Build all modules
r2r eac build

# Build single module (verbose output)
r2r eac build eac-commands

# Build multiple modules (parallel)
r2r eac build r2r-cli eac-core

# Build with go mod tidy first
r2r eac build --tidy-first r2r-cli

# Build with stripped debug info
r2r eac build --compressed r2r-cli

# Build with UPX compression
r2r eac build --compressed-upx r2r-cli

# Build with version injected
r2r eac build --version v1.2.3 r2r-cli

# Build without checking dependencies
r2r eac build --skip-deps eac-core

# Combine flags for release build
r2r eac build --version v1.2.0 --compressed-upx r2r-cli
```

### Output (Single Module)

```text
Building module: eac-commands (type: go-commands)
Module root: go/eac/commands
Output directory: C:\projects\eac\out\build\eac-commands
Build log: C:\projects\eac\out\build\eac-commands\build.log
Tidy mode: enabled (default for local builds)

=== go-commands: eac-commands ===
🔄 Running go mod tidy...
✅ go mod tidy completed
🔄 Running go generate...
✅ go generate completed

ℹ️  This module uses 'go run .' and is never compiled to a binary
ℹ️  Auto-built during testing (no explicit build needed)
```

### Output (CLI Module)

```text
Building module: r2r-cli (type: go-cli)
Module root: scripts/r2r-cli
Output directory: C:\projects\eac\out\build\r2r-cli

=== go-cli: r2r-cli ===
🔄 Running go mod tidy...
✅ go mod tidy completed
🔄 Building for linux/amd64...
✅ Built: r2r-linux-amd64
🔄 Building for linux/arm64...
✅ Built: r2r-linux-arm64
🔄 Building for windows/amd64...
✅ Built: r2r-windows-amd64.exe
🔄 Building for darwin/amd64...
✅ Built: r2r-darwin-amd64
🔄 Building for darwin/arm64...
✅ Built: r2r-darwin-arm64

✅ Build complete: 5 binaries generated
```

### Output (Multiple Modules)

```text
Building 3 modules in parallel...

  ✓ eac-core (2.3s)
  ✓ eac-commands (3.1s)
  ✓ r2r-cli (8.5s)

✅ All 3 modules built successfully
Total time: 8.5s
```

### Supported Module Types

| Type             | Build Action                      |
| ---------------- | --------------------------------- |
| `go-cli`         | Multi-platform binary compilation |
| `go-commands`    | Go package build (no binary)      |
| `go-library`     | Go package build (no binary)      |
| `go-mcp`         | MCP server binary compilation     |
| `mkdocs-site`    | `mkdocs build` → static site      |
| `mkdocs-book`    | PDF/EPUB generation               |
| `containers`     | Docker image build                |
| `specifications` | Gherkin syntax validation         |
| `contracts`      | Schema validation                 |
| `markdown`       | Markdown validation               |

### Build Output Location

```text
out/build/
├── <module>/
│   ├── build.log           # Full build output
│   └── <artifacts>         # Module-specific outputs
└── orchestrator.log        # Multi-module build summary
```

### Exit Codes

| Code | Description                  |
| ---- | ---------------------------- |
| 0    | Build completed successfully |
| 1    | Build failed                 |
| 2    | Module not found             |
| 3    | Invalid module type          |

---

## Common Workflows

### Local Development

```bash
# Build and test single module
r2r eac build src-auth
r2r eac test src-auth

# Build with validation
r2r eac build eac-commands && r2r eac validate
```

### CI/CD Pipeline

```bash
# Build all modules (skip tidy in CI)
r2r eac build --no-tidy

# Build with JUnit test output
r2r eac build && r2r eac test --as-junit
```

### Release Build

```bash
# Full release workflow
VERSION=$(git describe --tags --always)

# Build optimized binaries
r2r eac build --version $VERSION --compressed-upx r2r-cli

# Verify outputs
ls -la out/build/r2r-cli/
sha256sum out/build/r2r-cli/*
```

### Dependency Order Build

```bash
# Build in correct dependency order
r2r eac get execution order r2r-cli | while read module; do
  r2r eac build $module
done
```

### Selective Build

```bash
# Build only Go modules
r2r eac get modules --type go-* | while read module; do
  r2r eac build $module
done

# Build changed modules only
r2r eac get changed-modules | while read module; do
  r2r eac build $module
done
```

---

## Integration Patterns

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

CHANGED=$(r2r eac get changed-modules)
for module in $CHANGED; do
  echo "Building $module..."
  r2r eac build $module || exit 1
done
```

### GitHub Actions

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

      - name: Upload binaries
        uses: actions/upload-artifact@v4
        with:
          name: binaries
          path: out/build/
```

### Makefile

```makefile
.PHONY: build build-release clean

build:
  r2r eac build

build-release:
  r2r eac build --version $(VERSION) --compressed-upx r2r-cli

clean:
  rm -rf out/build/
```

---

## Compression Comparison

| Option             | ldflags       | Binary Size | Startup         |
| ------------------ | ------------- | ----------- | --------------- |
| None               | -             | 100%        | Fast            |
| `--compressed`     | `-s -w`       | ~70%        | Fast            |
| `--compressed-upx` | `-s -w` + UPX | ~30-40%     | Slightly slower |

### When to Use Each

- **Development**: No compression (fast builds)
- **CI artifacts**: `--compressed` (balanced)
- **Releases**: `--compressed-upx` (smallest size)

---

## Troubleshooting

| Problem             | Solution                             |
| ------------------- | ------------------------------------ |
| Module not found    | Check moniker in `modules.yml`       |
| Unknown module type | Verify `type` field in contract      |
| Build fails         | Check `out/build/<module>/build.log` |
| Go not found        | Install Go >= 1.21                   |
| UPX not found       | Install UPX or use `--compressed`    |
| Permission denied   | Check file permissions               |
| Cross-compile fails | Ensure `CGO_ENABLED=0`               |
| Version not showing | Use `var version = "dev"` in main.go |

---

## Related Documentation

- [Build Overview](build-overview.md) - Build system concepts
- [Build Configuration](build-configuration.md) - Configuration options
- [Test Commands](test-commands.md) - Run tests after building

{{ diataxis_footer() }}
