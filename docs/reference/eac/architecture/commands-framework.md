# Commands Framework

## Overview

The EAC commands framework has evolved from a monolithic commands module to a modular, domain-separated architecture. This refactoring improves maintainability, testability, and allows for independent evolution of command domains.

## Architecture Evolution

### Before: Monolithic Commands

Previously, all commands resided in a single `eac-commands` module:

```text
go/commands/
├── build.go
├── test.go
├── lint.go
├── scan.go
├── repository.go
├── update.go
└── ... (100+ files)
```

**Challenges**:

- Large, unwieldy codebase
- Tight coupling between unrelated commands
- Difficult to test in isolation
- Unclear ownership and boundaries

### After: Modular Commands

Commands are now organized into seven focused modules:

```text
go/commands/
├── base/           # Common infrastructure
├── build/          # Build automation
├── lint/           # Code quality
├── repository/     # Repository management
├── scan/           # Security scanning
├── test/           # Test execution
└── update/         # Maintenance and updates
```

**Benefits**:

- Clear domain separation
- Independent testing and deployment
- Reduced coupling
- Easier to understand and maintain

## Module Structure

### commands/base

**Purpose**: Provides foundational infrastructure for all commands

**Responsibilities**:

- Command registration and discovery
- Common flags and parameters
- Output formatting (table, JSON, YAML)
- Error handling and reporting
- Shared utilities

**Key Types**:

```go
// CommandRegistry manages command registration
type CommandRegistry struct {
    commands map[string]*Command
}

// Command represents a CLI command
type Command struct {
    Name        string
    Description string
    Flags       []Flag
    Handler     HandlerFunc
}
```

### commands/build

**Purpose**: Build automation and artifact generation

**Responsibilities**:

- Module compilation
- Artifact generation (executables, images)
- Dependency resolution
- Build caching
- Cross-platform builds

**Key Commands**:

- `build`: Build a module
- `build --deps`: Build with dependencies
- `build --skip-cache=local:state`: Force rebuild (ignore cached state)

### commands/lint

**Purpose**: Code quality and linting

**Responsibilities**:

- Code formatting (gofmt, prettier)
- Static analysis (go vet, eslint)
- Import organization
- Documentation checks

**Key Commands**:

- `lint`: Run linters
- `lint --fix`: Auto-fix issues

### commands/repository

**Purpose**: Repository and release management

**Responsibilities**:

- Git operations
- Release workflows
- Changelog generation
- Version management
- Tagging

**Key Commands**:

- `release-pending`: Check pending releases
- `release-this`: Create release
- `work-create`: Create work branch

### commands/scan

**Purpose**: Security scanning and vulnerability detection

**Responsibilities**:

- Dependency scanning
- Secret detection
- Container scanning
- Security reporting

**Key Commands**:

- `scan`: Run security scans
- `scan --severity high`: Filter by severity

### commands/test

**Purpose**: Test execution and management

**Responsibilities**:

- Test discovery
- Test execution (unit, integration, BDD)
- Result reporting
- Coverage tracking
- Parallel execution

**Key Commands**:

- `test`: Run tests
- `test-suite`: Run test suite
- `test --tags @L0`: Run tagged tests

### commands/update

**Purpose**: Maintenance and updates

**Responsibilities**:

- Dependency updates
- Documentation updates
- Evidence updates
- Cache management

**Key Commands**:

- `update-docs`: Update documentation
- `update-evidence`: Update test evidence
- `update-go-tidy`: Update Go dependencies

## Command Registration

Commands use an **init-based registration pattern** with import side-effects:

### Registration Flow

```text
1. Command Module          2. Import File           3. Build Registry
   (register.go)              (imports_*.go)           (allcmds/)

   ┌──────────────┐        ┌──────────────┐        ┌──────────────┐
   │ func init()  │        │ import _     │        │ BuildRegistry│
   │   register   │───────>│  "commands/  │───────>│   calls all  │
   │   Commands() │        │   build"     │        │   providers  │
   └──────────────┘        └──────────────┘        └──────────────┘
```

### Step 1: Each Command Module Registers via init()

```go
// go/commands/build/register.go
package build

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
    registry.RegisterProvider(Commands)  // Called automatically at startup
}
```

### Step 2: Import Files Control Module Linking

```go
// go/cli/eac/imports_build.go
//go:build !lite

package main

import _ "github.com/ready-to-release/eac/go/commands/build"  // Blank import triggers init()
```

**Build tags** (`!lite`) allow conditional inclusion of command groups.

### Step 3: allcmds Package Builds Registry

```go
// go/cli/eac/allcmds/allcmds.go
func BuildRegistry() (*registry.CommandRegistry, error) {
    reg := registry.NewCommandRegistry()

    for _, provider := range registry.Providers() {  // All registered providers
        if err := reg.RegisterAll(provider()...); err != nil {
            return nil, err
        }
    }

    registry.SetGlobal(reg)
    return reg, nil
}
```

### Import File Organization

Import files in `go/cli/eac/` control which command groups are linked:

- `imports_build.go` → `commands/build`
- `imports_lint.go` → `commands/lint`
- `imports_repository.go` → `commands/repository`
- `imports_scan.go` → `commands/scan`
- `imports_test_action.go` → `commands/test`
- `imports_update.go` → `commands/update`

This pattern enables:

- **Automatic registration** via `init()` functions
- **Conditional compilation** via build tags
- **Decoupled modules** - no explicit registration calls in main
- **Easy extension** - add new command module by creating import file

## Command Execution Flow

```text
┌─────────────────────────────────────┐
│       User Invocation               │
│     eac build core                  │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      EAC CLI / MCP Server           │
│  Parse command line arguments       │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      Command Registry               │
│  Lookup command by name             │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      commands/base                  │
│  Validate flags and parameters      │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      commands/build                 │
│  Execute build logic                │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      Output Formatter               │
│  Format and display results         │
└─────────────────────────────────────┘
```

## Dependency Architecture

```text
┌─────────────────────────────────────┐
│       EAC CLI / MCP Server          │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴──────────┐
    │   commands/base     │
    └──────────┬──────────┘
               │
      ┌────────┼────────┐
      │        │        │
┌─────▼──┐ ┌──▼───┐ ┌─▼────┐
│ build  │ │ test │ │ scan │  ...
└────┬───┘ └──┬───┘ └─┬────┘
     │        │       │
     └────────┼───────┘
              │
    ┌─────────▼─────────┐
    │  core / contracts │
    └───────────────────┘
```

**Dependency Rules**:

1. All command modules depend on `commands/base`
2. Command modules depend on `core` and `contracts`
3. Command modules MAY depend on `adapters` (for tool integration)
4. Command modules MUST NOT depend on other command modules

## Testing Strategy

### Unit Tests

Each command module has its own unit tests:

```text
go/commands/build/
├── build.go
├── build_test.go
├── artifacts.go
└── artifacts_test.go
```

### Integration Tests

Integration tests verify command interactions:

```bash
# Test build command
eac build test-module

# Verify artifacts were created
ls out/build/test-module/
```

### BDD Tests

Gherkin specifications for command behavior:

```gherkin
Feature: Build Command

  Scenario: Build a Go module
    Given a Go module "test-module"
    When I run "eac build test-module"
    Then the build should succeed
    And artifacts should exist in "out/build/test-module/"
```

## Migration Strategy

The migration from monolithic to modular commands followed this approach:

1. **Create `commands/base`**: Extract common infrastructure
2. **Identify domains**: Group related commands
3. **Create domain modules**: One module per domain
4. **Move commands**: Migrate commands to appropriate modules
5. **Update registrations**: Update command registration
6. **Update tests**: Adapt tests for new structure
7. **Update documentation**: Document new architecture

## Benefits Realized

### Before Refactoring

- **Module count**: 1 monolithic module
- **Lines of code**: ~15,000 LOC
- **Test time**: 45 seconds (all tests run together)
- **Coupling**: High (commands interdependent)

### After Refactoring

- **Module count**: 7 focused modules
- **Lines of code**: ~15,000 LOC (same total, better organized)
- **Test time**: 15 seconds (parallel execution per module)
- **Coupling**: Low (clear module boundaries)

### Maintainability Improvements

- **Clear ownership**: Each domain has a focused module
- **Easier testing**: Test modules in isolation
- **Independent evolution**: Modules can evolve separately
- **Better organization**: Related code grouped together

## Future Evolution

The modular architecture enables future enhancements:

### Command Plugins

Allow external commands to register:

```go
// External plugin
func Register(registry *base.CommandRegistry) error {
    return registry.Register(&base.Command{
        Name:    "custom-command",
        Handler: handleCustom,
    })
}
```

### Command Composition

Commands can invoke other commands:

```go
func handleBuildAndTest(ctx context.Context, args []string) error {
    // Build first
    if err := build.Execute(ctx, args); err != nil {
        return err
    }

    // Then test
    return test.Execute(ctx, args)
}
```

### Command Chains

Build workflows from command chains:

```yaml
workflows:
  - name: ci
    steps:
      - command: lint
      - command: build
      - command: test
      - command: scan
```

## See Also

- [Commands Module](../modules/commands.md) - Command module documentation
- [EAC Architecture](index.md) - Overall architecture
