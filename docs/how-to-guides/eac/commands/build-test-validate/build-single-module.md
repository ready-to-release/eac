# Build Single Module

## What You'll Accomplish

Compile a module and generate its artifacts, automatically handling dependencies in the correct order.

## Prerequisites

### Required Setup

- Module contract exists (module.yml)
- Dependencies are buildable
- Required build tools installed (Go, etc.)

## Steps

### 1. Build the Module

```bash
r2r eac build src-auth
```

**What happens**:

- Resolves dependencies
- Builds dependencies first
- Compiles src-auth
- Generates artifacts

### 2. Verify Build Succeeded

```bash
r2r eac show artifacts src-auth
```

**What happens**: Shows build artifacts with paths and sizes

### 3. Validate Artifacts Exist

```bash
r2r eac validate artifacts src-auth
```

**What happens**: Checks all required artifacts are present

## Build Options

```bash
# Build with version info
r2r eac build src-auth --version v1.2.0

# Build with compression
r2r eac build src-auth --compressed

# Build for specific OS
r2r eac build src-auth --os linux
```

## Example Scenario

You've made changes to src-auth and want to build it for testing:

```bash
# Build the module
r2r eac build src-auth

# Output:
# Building dependencies...
# ✓ eac-core (cached)
# Building src-auth...
# ✓ Compiled successfully
# ✓ Artifacts: out/bin/src-auth

# Verify artifacts
r2r eac show artifacts src-auth

# Run the built binary
./out/bin/src-auth --version
```

## Common Issues

| Problem | Solution |
|---------|----------|
| "Dependency not built" | Build dependency first or build all |
| Build fails | Check error output, ensure tools installed |
| Artifacts missing | Run `validate artifacts` for details |

## Next Steps

- [Run Tests for Module](./run-tests-for-module.md) → Test your changes
- [Build Changed Modules](./build-changed-modules.md) → Efficient CI builds

## Related Commands

- [`build`](../../../../reference/commands/build/build.md) - Full command reference
- [`show artifacts`](../../../../reference/commands/show/artifacts.md) - View artifacts
- [`validate artifacts`](../../../../reference/commands/validate/artifacts.md) - Verify artifacts
