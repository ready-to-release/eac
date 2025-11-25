# Build and Test Commands

**Problem**: Building and testing modules in a monorepo requires understanding each module's type and build tooling.

**Solution**: Use `build` and `test` commands that automatically dispatch to the correct tooling based on module contracts.

## Key Benefits

- Type-based dispatch (Go, MkDocs, containers, specs, etc.)
- Consistent interface across all module types
- Real-time output and proper exit codes
- Multiple output formats (text, Cucumber JSON, JUnit XML)
- Platform-specific builds (Windows, Linux, macOS)

## Quick Start

```bash
# Build a single module (verbose output)
r2r eac build src-commands

# Build multiple modules (parallel)
r2r eac build src-commands src-core

# Build all modules
r2r eac build

# Test a single module (verbose output)
r2r eac test src-commands

# Test multiple modules (parallel)
r2r eac test src-commands src-core

# Test all modules
r2r eac test

# Test with JUnit output (for CI/CD)
r2r eac test src-core --as-junit
```

## Command Reference

### build

Build one or more modules by moniker.

```bash
r2r eac build [module1] [module2] ... [options]

# Options:
--tidy-first           # Run 'go mod tidy' before building (default for local)
--no-tidy              # Skip 'go mod tidy' (default for CI)

# Examples:
r2r eac build                          # Build all modules
r2r eac build src-commands             # Build single module (verbose)
r2r eac build src-cli src-core         # Build multiple modules (parallel)
r2r eac build --tidy-first src-cli     # Build with go mod tidy first
```

**Supported module types:**

- `go-cli` - Go CLI applications
- `go-commands` - Go command packages
- `go-mcp` - Go MCP servers
- `go-library` - Go libraries
- `mkdocs-site` - MkDocs documentation sites
- `containers` - Docker containers
- `specifications` - Gherkin specs validation
- `contracts` - Contract validation
- `markdown` - Markdown validation
- And more...

### test

Test one or more modules by moniker.

```bash
r2r eac test [module1] [module2] ... [options]

# Options:
--as-cucumber          # Output in Cucumber JSON format (default)
--as-junit             # Output in JUnit XML format
--suite <name>         # Filter tests by suite (default: "commit")

# Examples:
r2r eac test                           # Test all modules
r2r eac test src-commands              # Test single module (verbose)
r2r eac test src-cli src-core          # Test multiple modules (parallel)
r2r eac test src-core --as-junit       # Test with JUnit output
r2r eac test src-commands --suite integration
```

**Supported module types:**

- `go-cli` - Run `go test`
- `go-commands` - Run `go test`
- `go-mcp` - Run `go test`
- `go-library` - Run `go test`
- `go-tests` - Run test suites

### test suite

Run a specific test suite.

```bash
r2r eac test suite <name> [options]

# Examples:
r2r eac test suite integration
r2r eac test suite e2e
r2r eac test suite smoke --as-junit
```

### test list-suites

List all available test suites.

```bash
r2r eac test list-suites

# Output:
# Available test suites:
# - integration (5 tests)
# - e2e (12 tests)
# - smoke (3 tests)
```

## Typical Workflows

### Local Development

```bash
# Build before testing
r2r eac build src-auth
r2r eac test src-auth

# Or combine with validation
r2r eac build src-auth && \
r2r eac test src-auth && \
r2r eac validate
```

### Continuous Integration

```bash
# Build all modules
r2r eac build

# Test with JUnit output
r2r eac test --as-junit > test-results.xml

# Validate contracts
r2r eac validate

# Exit with failure if any step fails
```

### TDD Workflow

```bash
# 1. Write spec
r2r eac specs create "Feature description"

# 2. Run tests (failing)
r2r eac test src-feature

# 3. Implement code
# ... edit src/feature/*.go ...

# 4. Build
r2r eac build src-feature

# 5. Run tests (passing)
r2r eac test src-feature

# 6. Commit
r2r eac work commit --all
```

## Module Type Dispatch

### Go Modules

**Build:**

- Runs `go build`
- Outputs to `out/build/<moniker>/`
- Platform-specific binaries for go-cli

**Test:**

- Runs `go test ./...`
- Table-driven tests
- Outputs coverage reports

### MkDocs Sites

**Build:**

- Runs `mkdocs build`
- Validates markdown links
- Generates static site in `out/build/<moniker>/site/`

**Test:**

- Validates markdown syntax
- Checks internal links
- Verifies navigation structure

### Specifications

**Build:**

- Validates Gherkin syntax
- Checks against contracts
- Reports validation errors

**Test:**

- Runs BDD step definitions
- Cucumber JSON output
- Test reports in `out/reports/`

### Containers

**Build:**

- Builds Docker images
- Tags appropriately
- Pushes to registry (if configured)

**Test:**

- Container smoke tests
- Health checks
- Integration tests

## Output Formats

### Standard Output (Default)

```text
Building module: src-commands (type: go-commands)
Module root: src/commands
Output directory: C:\projects\eac\out\build\src-commands
Build log: C:\projects\eac\out\build\src-commands\build.log
Tidy mode: enabled (default for local builds)

=== go-commands: src-commands ===
🔄 Running go mod tidy...
✅ go mod tidy completed
🔄 Running go generate...
✅ go generate completed

ℹ️  This module uses 'go run .' and is never compiled to a binary
ℹ️  Auto-built during testing (no explicit build needed)
```

### Cucumber JSON (Default for tests)

```json
{
  "description": "Module test results",
  "elements": [
    {
      "id": "feature-1",
      "name": "Authentication",
      "type": "scenario",
      "steps": [
        {
          "keyword": "Given",
          "name": "user exists",
          "result": {"status": "passed"}
        }
      ]
    }
  ]
}
```

### JUnit XML

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="Module Tests">
  <testsuite name="src-commands" tests="10" failures="0" errors="0" time="2.5">
    <testcase name="TestExample" classname="commands" time="0.25"/>
  </testsuite>
</testsuites>
```

## Build Outputs

### Directory Structure

```text
out/
├── build/                        # Build outputs per module
│   ├── src-cli/
│   │   ├── build.log
│   │   ├── r2r-linux
│   │   ├── r2r-windows.exe
│   │   ├── r2r-darwin-amd64
│   │   └── r2r-darwin-arm64
│   ├── docs/
│   │   ├── build.log
│   │   └── site/
│   └── orchestrator.log          # Multi-module build log
└── reports/                      # Test reports
    ├── test-results.json
    └── coverage.html
```

## Integration Patterns

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Building and testing changed modules..."

# Get changed files
CHANGED_FILES=$(git diff --cached --name-only)

# Determine affected modules
MODULES=$(r2r eac get-changed-modules)

# Build and test each module
for module in $MODULES; do
  echo "Building $module..."
  r2r eac build $module || exit 1

  echo "Testing $module..."
  r2r eac test $module || exit 1
done

echo "✅ All checks passed"
```

### GitHub Actions

```yaml
name: Build and Test

on: [push, pull_request]

jobs:
  build-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build all modules
        run: r2r eac build

      - name: Test all modules
        run: r2r eac test --as-junit > test-results.xml

      - name: Publish test results
        uses: dorny/test-reporter@v1
        if: always()
        with:
          name: Test Results
          path: test-results.xml
          reporter: java-junit
```

### Make Integration

```makefile
.PHONY: build test clean

build:
	r2r eac build

test:
	r2r eac test

validate:
	r2r eac validate

ci: build test validate
	@echo "✅ CI pipeline completed"

clean:
	rm -rf out/
```

## Best Practices

- **Build before testing**: Ensure code compiles before running tests
- **Use module contracts**: Define module types correctly
- **CI/CD output formats**: Use `--as-junit` for CI systems
- **Validate contracts**: Run `r2r eac validate` regularly
- **Watch build times**: Monitor execution time in output
- **Check exit codes**: Non-zero indicates failure

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Module not found | Check moniker in module contract |
| Unknown module type | Verify module type in contract YAML |
| Build fails | Check module dependencies, run with verbose output |
| Tests fail | Review test output, check test files exist |
| Permission denied | Check file permissions, may need sudo |
| Go not found | Install Go ≥ 1.21 |
| Output format error | Use valid format: `--as-cucumber` or `--as-junit` |

## Advanced Usage

### Selective Module Building

```bash
# Build only Go modules
r2r eac get-modules --type go-* | while read module; do
  r2r eac build $module
done

# Build in dependency order
r2r eac get-execution-order src-cli | while read module; do
  r2r eac build $module
done
```

### Parallel Builds

```bash
# Build multiple specific modules (built-in parallel)
r2r eac build src-commands src-core src-cli
```

## Summary

**Build workflow:**

1. `r2r eac build <moniker>` - Build single module (verbose output)
2. `r2r eac build <m1> <m2>` - Build multiple modules (parallel)
3. `r2r eac build` - Build all modules
4. Check output in `out/build/<moniker>/`

**Test workflow:**

1. `r2r eac test <moniker>` - Test single module (verbose output)
2. `r2r eac test <m1> <m2>` - Test multiple modules (parallel)
3. `r2r eac test` - Test all modules
4. Use `--as-junit` for CI/CD integration
5. Check reports in `out/reports/`

Module contracts enable consistent build and test operations across heterogeneous module types.
