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
# Build a module
r2r eac build module src-commands

# Test a module
r2r eac test module src-commands

# Build all modules
r2r eac build modules

# Test with JUnit output (for CI/CD)
r2r eac test module src-core --as-junit
```

## Command Reference

### build module

Build a single module by moniker.

```bash
r2r eac build module <moniker> [options]

# Options:
--windows-only         # Build only for Windows
--linux-only           # Build only for Linux
--macos-only           # Build only for macOS

# Examples:
r2r eac build module src-commands
r2r eac build module src-cli --windows-only
r2r eac build module docs
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

### build modules

Build multiple modules.

```bash
r2r eac build modules [options]

# Options:
--windows-only         # Build only for Windows
--linux-only           # Build only for Linux
--macos-only           # Build only for macOS

# Examples:
r2r eac build modules
r2r eac build modules --linux-only
```

### test module

Test a single module by moniker.

```bash
r2r eac test module <moniker> [options]

# Options:
--as-cucumber          # Output in Cucumber JSON format (default)
--as-junit             # Output in JUnit XML format
--no-generate          # Skip test summary generation
--generate-only        # Generate summary without running tests

# Examples:
r2r eac test module src-commands
r2r eac test module src-core --as-junit
r2r eac test module src-auth --as-cucumber
r2r eac test module src-api --generate-only
```

**Supported module types:**
- `go-cli` - Run `go test`
- `go-commands` - Run `go test`
- `go-mcp` - Run `go test`
- `go-library` - Run `go test`
- `go-tests` - Run test suites

### test modules

Test multiple modules.

```bash
r2r eac test modules [options]

# Same options as test module
--as-cucumber          # Cucumber JSON format
--as-junit             # JUnit XML format

# Example:
r2r eac test modules
r2r eac test modules --as-junit
```

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
r2r eac build module src-auth
r2r eac test module src-auth

# Or combine with validation
r2r eac build module src-auth && \
r2r eac test module src-auth && \
r2r eac validate
```

### Continuous Integration

```bash
# Build all modules
r2r eac build modules

# Test with JUnit output
r2r eac test modules --as-junit > test-results.xml

# Validate contracts
r2r eac validate

# Exit with failure if any step fails
```

### Platform-Specific Builds

```bash
# Build for distribution
r2r eac build module src-cli --windows-only
r2r eac build module src-cli --linux-only
r2r eac build module src-cli --macos-only

# Outputs in out/bin/<platform>/
```

### TDD Workflow

```bash
# 1. Write spec
r2r eac specs create "Feature description"

# 2. Run tests (failing)
r2r eac test module src-feature

# 3. Implement code
# ... edit src/feature/*.go ...

# 4. Build
r2r eac build module src-feature

# 5. Run tests (passing)
r2r eac test module src-feature

# 6. Commit
r2r eac work commit --all
```

## Module Type Dispatch

### Go Modules

**Build:**
- Runs `go build`
- Outputs to `out/bin/`
- Platform-specific binaries

**Test:**
- Runs `go test ./...`
- Table-driven tests
- Outputs coverage reports

### MkDocs Sites

**Build:**
- Runs `mkdocs build`
- Validates markdown links
- Generates static site in `out/site/`

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

```
Building module: src-commands
Module type: go-commands
Output dir: out

Running: go build -o out/bin/commands ./src/commands
Build successful

Execution time: 1.234s
✅ Build completed successfully
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

```
out/
├── bin/                          # Compiled binaries
│   ├── windows/
│   │   └── r2r.exe
│   ├── linux/
│   │   └── r2r
│   └── darwin/
│       └── r2r
├── site/                         # Built documentation
│   └── index.html
├── reports/                      # Test reports
│   ├── test-results.json
│   └── coverage.html
└── logs/                         # Build logs
    └── build-src-commands.log
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
  r2r eac build module $module || exit 1

  echo "Testing $module..."
  r2r eac test module $module || exit 1
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
        run: r2r eac build modules

      - name: Test all modules
        run: r2r eac test modules --as-junit > test-results.xml

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
	r2r eac build modules

test:
	r2r eac test modules

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
- **Platform-specific builds**: Use flags for distribution builds
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
  r2r eac build module $module
done

# Build in dependency order
r2r eac get-execution-order src-cli | while read module; do
  r2r eac build module $module
done
```

### Parallel Builds

```bash
# Build modules in parallel (GNU parallel)
r2r eac get-modules | parallel -j 4 r2r eac build module {}
```

### Custom Build Scripts

```bash
# Pre-build hook
if [ -f "src/module/.prebuild.sh" ]; then
  ./src/module/.prebuild.sh
fi

# Build
r2r eac build module src-module

# Post-build hook
if [ -f "src/module/.postbuild.sh" ]; then
  ./src/module/.postbuild.sh
fi
```

## Summary

**Build workflow:**
1. `r2r eac build module <moniker>` - Build single module
2. `r2r eac build modules` - Build all modules
3. Check output in `out/bin/` or `out/site/`

**Test workflow:**
1. `r2r eac test module <moniker>` - Test single module
2. `r2r eac test modules` - Test all modules
3. Use `--as-junit` for CI/CD integration
4. Check reports in `out/reports/`

Module contracts enable consistent build and test operations across heterogeneous module types.
