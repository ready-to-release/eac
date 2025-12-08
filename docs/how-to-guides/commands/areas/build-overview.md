<!-- EDITOR
# Editor: how-to-guides/commands/areas/build-overview.md

## Soul

Type-aware build system that dispatches to correct tooling (Go, MkDocs, Docker) based on module contracts, supporting multi-platform compilation with compression options.

## Sections

1. What is the Build System?
2. When to Use Build Commands
3. Common Use Cases
4. Key Concepts
   - Module Type Dispatch
   - Build Output Structure
   - Platform Targets
   - Compression Options
5. Workflow Overview
   - Local Development
   - CI/CD Pipeline
   - Release Build
6. Module Type Details
   - Go Modules
   - MkDocs Sites
   - Containers
7. Integration Points
   - With Testing
   - With Validation
   - With Release
8. Best Practices
9. Troubleshooting
10. Next Steps
11. Related Areas
-->

# Build System

The build system in EAC provides type-aware compilation and artifact generation for all module types in your monorepo.

## What is the Build System?

EAC's build system enables you to:

- **Build any module type** with a single command
- **Dispatch to correct tooling** based on module contracts
- **Generate optimized artifacts** with compression options
- **Inject version information** into binaries
- **Build for multiple platforms** (Windows, Linux, macOS)

The system automatically detects module types and invokes the appropriate build tooling.

## When to Use Build Commands

Use build commands when you need:

| Scenario               | Command                  |
| ---------------------- | ------------------------ |
| Build a single module  | `build <module>`         |
| Build multiple modules | `build <m1> <m2>`        |
| Build all modules      | `build`                  |
| Build with compression | `build --compressed`     |
| Build with version     | `build --version v1.0.0` |

### Common Use Cases

- **Local development** - Build before testing
- **CI/CD pipelines** - Generate release artifacts
- **Cross-platform builds** - Target multiple OS/architectures
- **Optimized releases** - Strip debug info, apply UPX compression

## Key Concepts

### Module Type Dispatch

The build system dispatches to different builders based on module type:

| Module Type      | Builder          | Output              |
| ---------------- | ---------------- | ------------------- |
| `go-cli`         | Go compiler      | Platform binaries   |
| `go-commands`    | Go compiler      | Library package     |
| `go-library`     | Go compiler      | Library package     |
| `go-mcp`         | Go compiler      | MCP server binary   |
| `mkdocs-site`    | MkDocs           | Static HTML site    |
| `containers`     | Docker           | Container image     |
| `specifications` | Gherkin parser   | Validated specs     |
| `contracts`      | Schema validator | Validated contracts |

### Build Output Structure

```text
out/
└── build/
    ├── r2r-cli/
    │   ├── build.log
    │   ├── r2r-linux-amd64
    │   ├── r2r-linux-arm64
    │   ├── r2r-windows-amd64.exe
    │   ├── r2r-darwin-amd64
    │   └── r2r-darwin-arm64
    ├── docs/
    │   ├── build.log
    │   └── site/
    └── orchestrator.log
```

### Platform Targets

For `go-cli` modules, builds target multiple platforms:

| Platform | Architecture | Binary                     |
| -------- | ------------ | -------------------------- |
| Linux    | amd64        | `<name>-linux-amd64`       |
| Linux    | arm64        | `<name>-linux-arm64`       |
| Windows  | amd64        | `<name>-windows-amd64.exe` |
| macOS    | amd64        | `<name>-darwin-amd64`      |
| macOS    | arm64        | `<name>-darwin-arm64`      |

### Compression Options

| Option             | Effect               | Size Reduction |
| ------------------ | -------------------- | -------------- |
| `--compressed`     | Strip debug symbols  | ~30%           |
| `--compressed-upx` | Strip + UPX compress | ~60-70%        |

## Workflow Overview

### Local Development

```bash
# Build single module (verbose output)
r2r eac build eac-commands

# Build and test
r2r eac build src-auth && r2r eac test src-auth
```

### CI/CD Pipeline

```bash
# Build all modules
r2r eac build

# Build with version injection
r2r eac build --version $GIT_TAG r2r-cli

# Build optimized release
r2r eac build --compressed-upx r2r-cli
```

### Release Build

```bash
# Full release build workflow
r2r eac build --version v1.2.0 --compressed-upx r2r-cli

# Verify outputs
ls out/build/r2r-cli/
```

## Module Type Details

### Go Modules

Go modules (`go-cli`, `go-commands`, `go-library`, `go-mcp`) use:

- `go mod tidy` (optional, controlled by flags)
- `go generate` (if generate directives exist)
- `go build` with appropriate flags

```bash
# With tidy first (default for local)
r2r eac build --tidy-first eac-commands

# Skip tidy (default for CI)
r2r eac build --no-tidy eac-commands
```

### MkDocs Sites

MkDocs sites build static documentation:

```bash
r2r eac build docs

# Output: out/build/docs/site/
```

### Containers

Container modules build Docker images:

```bash
r2r eac build ext-eac

# Tags and pushes to configured registry
```

## Integration Points

### With Testing

```bash
# Build then test
r2r eac build eac-core
r2r eac test eac-core
```

### With Validation

```bash
# Full CI workflow
r2r eac build && r2r eac test && r2r eac validate
```

### With Release

```bash
# Pre-release build
r2r eac build --version $(r2r eac release get-version) --compressed r2r-cli
```

## Best Practices

### Do's

- **Build before testing** - Ensure compilation succeeds
- **Use compression for releases** - Smaller artifacts
- **Inject versions** - Traceability in binaries
- **Check build logs** - `out/build/<module>/build.log`

### Don'ts

- **Don't skip tidy in releases** - Ensures clean dependencies
- **Don't ignore build failures** - Fix before proceeding
- **Don't commit build artifacts** - Keep `out/` in .gitignore

## Troubleshooting

| Problem             | Solution                          |
| ------------------- | --------------------------------- |
| Module not found    | Check moniker in module contract  |
| Unknown module type | Verify type in modules.yml        |
| Go not found        | Install Go >= 1.21                |
| UPX not found       | Install UPX or use `--compressed` |
| Permission denied   | Check file permissions            |

## Next Steps

- [Build Configuration](build-configuration.md) - Configure build options
- [Build Commands](build-commands.md) - Full command reference

## Related Areas

- [Test](test-overview.md) - Run tests after building
- [Validate](validate-overview.md) - Validate before building
- [Release](release-overview.md) - Release built artifacts
