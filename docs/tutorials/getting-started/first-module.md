# Your First Module

Learn how to create a module in this repository - the fundamental unit of code organization that defines ownership, dependencies, and build behavior.

**Prerequisites:** [Quick Start Guide](./quick-start.md), Go 1.21+ installed

## What You'll Learn

By the end of this tutorial, you'll be able to:

- Understand what a module is and why it matters
- Choose the right module type for your code
- Create a module contract in `.r2r/eac/repository.yml`
- Define file ownership patterns
- Build and test your module
- Validate module configuration
- Understand module dependencies and build order

## What is a Module?

A **module** is an independently buildable unit of code with explicit identity, file ownership, dependencies, and build configuration. Modules are the foundation of the repository's modular architecture.

**Key characteristics:**

- **Identity** - Unique moniker (e.g., `eac-core`, `my-service`)
- **Type** - Classification providing build/test behavior (e.g., `go`, `container`, `typescript`)
- **File Ownership** - Explicit patterns defining which files belong to this module
- **Dependencies** - Declared dependencies on other modules
- **Build Contract** - How to build, test, and verify the module

**Why modules matter:**

- Enable **independent development** - Work on modules in isolation
- Enable **incremental builds** - Only rebuild changed modules and dependents
- Ensure **clear ownership** - Each file belongs to exactly one module
- Support **dependency management** - Build in correct topological order
- Enable **parallel testing** - Test modules independently

**Example:** The `eac-core` module contains core domain libraries, owns all `.go` files in `go/eac/core/`, depends on `logging-go`, and builds as a Go library.

!!! info "Learn More"
    For comprehensive information about modules, see:

    - [Modules Reference](../../reference/r2r-eac/modules.md) - Complete module system documentation
    - [Module Contracts](../../reference/r2r-eac/contracts.md#modules-contract) - Contract specification and validation
    - [Module Types](../../reference/r2r-eac/module-types-reference.md) - Available types and capabilities

## Step 1: Choose Your Module Type

Before creating a module, select the appropriate type based on your language and purpose:

| Language/Purpose | Module Type | Build Support | Test Support |
|------------------|-------------|---------------|--------------|
| Go library/service | `go` | ✅ Full (cross-compile, version inject) | ✅ gotest, godog |
| TypeScript/JavaScript | `typescript` | ✅ npm, tsc | ✅ mocha, cucumber-js |
| Any (containerized) | `container` | ✅ Docker buildx | Depends on container |
| Documentation | `docs` | ✅ MkDocs | ❌ No tests |
| Configuration/static files | `static` | ❌ No build | ❌ No tests |

For this tutorial, we'll create a **Go module** (`type: go`) - a simple greeting service.

!!! tip "Choosing Types"
    - Use `go` for all Go code (libraries, services, tools)
    - Use `container` for Python, Rust, Java, or multi-language applications
    - See [Module Types Reference](../../reference/r2r-eac/module-types-reference.md) for all available types

## Step 2: Plan Your Module

Let's create a module called `greet-service`:

- **Moniker:** `greet-service` (unique identifier)
- **Type:** `go` (Go module)
- **Location:** `go/examples/greet-service/` (source code directory)
- **Purpose:** Provide greeting functionality
- **Dependencies:** None (standalone example)

## Step 3: Create the Directory Structure

Create the module directory and initialize the Go module:

```bash
# Create directory
mkdir -p go/examples/greet-service

# Create Go module files
cd go/examples/greet-service

# Initialize Go module (in workspace mode, this is optional but recommended)
go mod init github.com/ready-to-release/eac/go/examples/greet-service
```

## Step 4: Write the Implementation

Create `go/examples/greet-service/greet.go`:

```go
package greet

import "fmt"

// Greet returns a personalized greeting message.
// If name is empty, returns a default greeting.
func Greet(name string) string {
    if name == "" {
        return "Hello, Guest!"
    }
    return fmt.Sprintf("Hello, %s!", name)
}
```

## Step 5: Write Unit Tests

Create `go/examples/greet-service/greet_test.go`:

```go
package greet

import "testing"

func TestGreet(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "with valid name",
            input:    "Alice",
            expected: "Hello, Alice!",
        },
        {
            name:     "with empty name",
            input:    "",
            expected: "Hello, Guest!",
        },
        {
            name:     "with name containing special characters",
            input:    "Bob-123",
            expected: "Hello, Bob-123!",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Greet(tt.input)
            if result != tt.expected {
                t.Errorf("Greet(%q) = %q, want %q", tt.input, result, tt.expected)
            }
        })
    }
}
```

Verify tests pass:

```bash
go test ./go/examples/greet-service
```

## Step 6: Register the Module

Open `.r2r/eac/repository.yml` and add the module contract:

```yaml
modules:
  # ... existing modules ...

  - moniker: greet-service
    name: Greeting Service Example
    type: go
    description: Simple greeting service demonstrating module creation
    files:
      root: go/examples/greet-service
      source:
        - "**/*.go"
        - "!**/*_test.go"
      tests:
        - "**/*_test.go"
      config:
        - "go.mod"
        - "go.sum"
```

**Understanding the contract:**

- `moniker` - Unique identifier for the module (required)
- `name` - Human-readable name (required)
- `type` - Module type reference (required, must exist in `module-types.yml`)
- `description` - Module purpose (optional but recommended)
- `files.root` - Base directory containing module files (required)
- `files.source` - Glob patterns for source files (inherits from type defaults if not specified)
- `files.tests` - Glob patterns for test files
- `files.config` - Configuration files (go.mod, go.sum)

!!! note "File Ownership Rule"
    Each file must be claimed by **exactly one module**. The validation system ensures no overlapping ownership.

## Step 7: Validate the Module

Verify the module contract is correct:

```bash
# Validate contract schema
r2r eac validate contracts

# Validate file ownership (no overlaps)
r2r eac validate module-files

# Show all modules (your module should appear)
r2r eac show modules
```

Expected output from `show modules`:

```text
Moniker         Name                    Type  Dependencies
greet-service   Greeting Service Example go    (none)
```

## Step 8: Build the Module

Build your module using the R2R CLI:

```bash
r2r eac build greet-service
```

This will:

1. Load the module contract
2. Resolve dependencies (none for this module)
3. Build the Go package
4. Create a marker file in `out/build/greet-service/.built`
5. Cache the build result

!!! tip "Build Artifacts"
    Go library modules create a `.built` marker file. Executable modules (`go-cli` type) create platform-specific binaries.

## Step 9: Test the Module

Run tests for your module:

```bash
r2r eac test module greet-service
```

This will:

1. Run `go test` for all test files
2. Generate test results in `out/test/greet-service/`
3. Display test summary

## Step 10: Query Module Information

Explore information about your module:

```bash
# Show module files
r2r eac get files greet-service

# Show module details (JSON)
r2r eac get modules --moniker=greet-service

# Show build artifacts
r2r eac show artifacts greet-service
```

## Understanding Module Dependencies

If your module depends on other modules, declare dependencies in the contract:

```yaml
- moniker: my-service
  type: go
  depends_on:
    - eac-core      # Depends on core libraries
    - logging-go    # Depends on logging library
  files:
    root: go/my/service
```

**Dependency rules:**

- Dependencies must exist in `repository.yml`
- No circular dependencies allowed
- Build order determined by topological sort
- Changed modules trigger rebuild of dependents

View dependency graph:

```bash
r2r eac show dependencies
```

!!! info "Learn More About Dependencies"
    See [Dependency Management](../../reference/r2r-eac/modules.md#dependency-management) for complete information.

## Best Practices

1. **Use descriptive monikers** - Choose clear, lowercase names with hyphens (e.g., `user-service`, not `UserService`)
2. **Choose the correct type** - Match module type to your language and purpose
3. **Be explicit with file patterns** - Define clear ownership boundaries
4. **Declare all dependencies** - Make dependency relationships explicit
5. **Keep modules focused** - Single responsibility per module
6. **Validate early and often** - Run `validate contracts` before committing changes

**Good module:**

```yaml
- moniker: auth-service
  name: Authentication Service
  type: go
  description: Handles user authentication and session management
  depends_on: [config-go, logging-go]
  files:
    root: go/services/auth
```

**Avoid:**

```yaml
- moniker: AuthService          # Use kebab-case
  type: go-app                  # Type doesn't exist
  files:
    root: src/auth              # Inconsistent with repo structure
```

## Common Module Types

Beyond basic `go` modules, you can create:

**Executable application:**

```yaml
- moniker: my-cli
  type: go
  build:
    artifacts:
      - id: linux-amd64
        type: executable
        pattern: "my-cli-linux-amd64"
  files:
    root: go/my/cli
```

**Container image:**

```yaml
- moniker: my-api
  type: container
  files:
    root: containers/my-api
    source:
      - "Dockerfile"
      - "**/*"
```

**Documentation:**

```yaml
- moniker: api-docs
  type: docs
  files:
    root: docs/api
```

See [Module Types Reference](../../reference/r2r-eac/module-types-reference.md) for all available types.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Module not found" | Verify moniker in `repository.yml` matches command argument |
| "File claimed by multiple modules" | Adjust `files` patterns to eliminate overlaps |
| "Dependency not found" | Ensure dependency exists in `repository.yml` |
| "Circular dependency detected" | Refactor to break the cycle (extract shared code to new module) |
| "Build failed" | Check Go compilation errors, verify go.mod is correct |

Run validation to diagnose issues:

```bash
r2r eac validate
```

## What You Learned

Congratulations! You've successfully:

- ✅ Understood what a module is and its key characteristics
- ✅ Chosen the appropriate module type (`go`)
- ✅ Created module directory structure and Go code
- ✅ Written table-driven unit tests
- ✅ Registered the module in `.r2r/eac/repository.yml`
- ✅ Validated module configuration
- ✅ Built and tested the module with R2R commands
- ✅ Queried module information

## Key Concepts Covered

- **Module architecture** - Independently buildable units with explicit contracts
- **Module types** - `go`, `container`, `typescript`, `static`, `docs`
- **File ownership** - Glob patterns defining module boundaries
- **Module contracts** - YAML configuration in `repository.yml`
- **Validation system** - Schema and file ownership validation
- **Build system** - Topological build order based on dependencies
- **Commands** - `validate`, `build`, `test`, `show modules`, `get files`

## Next Steps

### Continue Learning

- **Previous:** [Your First Specification](./first-specification.md) - Write Gherkin specifications
- **Next:** [Understanding Test Suites](./understanding-test-suites.md) - Organize tests by level and type
- **Advanced:** [Multi-Module Development](../advanced-practices/multi-module-development.md) - Work with multiple related modules

### Apply What You Learned

Now that you can create modules, you can accomplish these tasks:

- **[Create modules](../../how-to-guides/eac/modules/creating-modules.md)** - Detailed how-to guide
- **[Build modules](../../how-to-guides/eac/commands/build-test-validate/build-single-module.md)** - Build individual modules
- **[Test modules](../../how-to-guides/eac/commands/build-test-validate/run-tests-for-module.md)** - Run tests for modules
- **Explore dependencies** - Use `show dependencies` to understand module relationships

### Dive Deeper

- [Modules Reference](../../reference/r2r-eac/modules.md) - Complete module system documentation
- [Module Types Reference](../../reference/r2r-eac/module-types-reference.md) - All available types and capabilities
- [Contracts Reference](../../reference/r2r-eac/contracts.md) - Contract specification and validation
- [Repository Layout](../../reference/r2r-eac/repository-layout.md) - How modules fit in the repository structure
