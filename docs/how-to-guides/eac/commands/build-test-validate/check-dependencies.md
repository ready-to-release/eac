# Check Dependencies

## What You'll Accomplish

Verify module dependencies match contracts and are properly configured.

## Prerequisites

- Module contracts defined (module.yml)
- Go modules configured (go.mod)

## Steps

### 1. Validate Dependencies

```bash
eac validate dependencies
```

**What happens**: Checks all module dependencies match their contracts

### 2. View Dependency Graph

```bash
eac show dependencies
```

**What happens**: Displays visual dependency graph

### 3. Check Go Module Tidiness

```bash
eac validate go-tidy
```

**What happens**: Verifies go.mod and go.sum are tidy

### 4. Fix Issues

```bash
# If go.mod not tidy
go mod tidy

# If dependencies mismatch
# Update module.yml dependencies section
```

## Example Scenario

After adding new module dependency:

```bash
# Check dependencies
eac validate dependencies

# Output:
# ✗ Module src-api: depends on src-auth v1.2.0 but using v1.1.0
# ✗ Module src-db: not listed in go.mod

# View current state
eac show dependencies | grep src-api

# Fix go.mod
go get github.com/org/src-auth@v1.2.0
go mod tidy

# Validate again
eac validate dependencies
# ✓ All dependencies valid

# Ensure go.mod is tidy
eac validate go-tidy
# ✓ Go modules are tidy
```

## Dependency Patterns

```bash
# Get build dependencies only
eac get build-deps src-api

# Get full dependency tree as JSON
eac get dependencies | jq '.dependencies["src-api"]'
```

## Common Issues

| Problem             | Solution                        |
| ------------------- | ------------------------------- |
| Version mismatch    | Update go.mod to match contract |
| Missing dependency  | Add to module.yml and go.mod    |
| Circular dependency | Refactor module structure       |

## Next Steps

- [Build Single Module](./build-single-module.md) → Build with deps

## Related Commands

- [`validate dependencies`](../../../../reference/eac/commands/validate/dependencies.md) - Check contracts
- [`show dependencies`](../../../../reference/eac/commands/show/dependencies.md) - View graph
- [`validate go-tidy`](../../../../reference/eac/commands/validate/go-tidy.md) - Check tidiness
