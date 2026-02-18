# Command Implementation Guide

This guide explains how commands are actually implemented in the EAC CLI, based on the modular commands framework.

## Overview

EAC commands use an **init-based registration pattern** where command modules self-register at startup through import side-effects. This design enables:

- **Automatic discovery** - No explicit registration calls needed
- **Modular organization** - Commands grouped by domain
- **Conditional compilation** - Build tags control which commands are included
- **Decoupled architecture** - Command modules don't depend on CLI main package

---

## Directory Structure

```text
go/
├── cli/eac/                    # EAC CLI application
│   ├── main.go                 # CLI entry point
│   ├── allcmds/                # Registry builder
│   │   └── allcmds.go          # BuildRegistry() function
│   ├── imports_build.go        # Triggers build module import
│   ├── imports_lint.go         # Triggers lint module import
│   ├── imports_repository.go  # Triggers repository module import
│   ├── imports_scan.go         # Triggers scan module import
│   ├── imports_test_action.go # Triggers test module import
│   └── imports_update.go       # Triggers update module import
│
└── commands/                   # Command modules
    ├── base/                   # Shared command infrastructure
    ├── build/                  # Build commands
    │   ├── register.go         # Self-registration via init()
    │   ├── build.go            # Build command implementation
    │   └── commands.go         # Command definitions
    ├── lint/                   # Lint commands
    ├── repository/             # Repository commands (create, serve, etc.)
    ├── scan/                   # Security scan commands
    ├── test/                   # Test commands
    └── update/                 # Update commands
```

---

## Command Registration Flow

### Step 1: Command Module Self-Registers

Each command module has a `register.go` file that registers a provider function:

```go
// go/commands/build/register.go
package build

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
    // Called automatically when package is imported
    registry.RegisterProvider(Commands)
}
```

The `Commands` function returns all commands in the module:

```go
// go/commands/build/commands.go
package build

import core "github.com/ready-to-release/eac/contracts/core/0.1.0"

func Commands() []core.CommandProvider {
    return []core.CommandProvider{
        &buildCommand{},
    }
}
```

### Step 2: Import File Triggers Registration

Import files use **blank imports** to trigger init() functions:

```go
// go/cli/eac/imports_build.go
//go:build !lite

package main

import _ "github.com/ready-to-release/eac/go/commands/build"
```

**Build tags** control conditional inclusion:
- `//go:build !lite` - Include in full build, exclude in lite build
- Enables minimal CLIs for specific use cases

### Step 3: Registry Builder Collects All Commands

The `allcmds` package builds the final registry:

```go
// go/cli/eac/allcmds/allcmds.go
func BuildRegistry() (*registry.CommandRegistry, error) {
    reg := registry.NewCommandRegistry()

    // Call all registered providers
    for _, provider := range registry.Providers() {
        if err := reg.RegisterAll(provider()...); err != nil {
            return nil, err
        }
    }

    registry.SetGlobal(reg)
    flags.SetRegistry(reg)
    return reg, nil
}
```

### Step 4: CLI Main Invokes Command

```go
// go/cli/eac/main.go
func main() {
    // Build command registry from all imported modules
    reg, err := allcmds.BuildRegistry()
    if err != nil {
        log.Fatal(err)
    }

    // Execute command based on CLI arguments
    exitCode := reg.Execute(os.Args[1:])
    os.Exit(exitCode)
}
```

---

## Implementing a New Command

### 1. Create Command Implementation

```go
// go/commands/mygroup/mycommand.go
package mygroup

import (
    "context"
    core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

type myCommand struct{}

var _ core.SimpleCommandPort = (*myCommand)(nil)

func (c *myCommand) Name() string {
    return "my command"
}

func (c *myCommand) Metadata() core.CommandMetadata {
    return core.CommandMetadata{
        CanonicalName: "my-command",
        Short:         "Brief description of command",
        Long:          "Detailed explanation of what the command does...",
        Args:          "arguments",
        Flags: []core.FlagSpec{
            {
                Name:  "flag-name",
                Type:  "string",
                Usage: "Description of flag",
            },
        },
    }
}

func (c *myCommand) Execute(ctx context.Context, req *core.CommandRequest) int {
    // Command logic here
    log.Info("Executing my command")
    return 0  // Exit code: 0 = success, non-zero = error
}
```

### 2. Add Command to Module

```go
// go/commands/mygroup/commands.go
package mygroup

import core "github.com/ready-to-release/eac/contracts/core/0.1.0"

func Commands() []core.CommandProvider {
    return []core.CommandProvider{
        &myCommand{},
        // Other commands in this module...
    }
}
```

### 3. Create Registration File

```go
// go/commands/mygroup/register.go
package mygroup

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
    registry.RegisterProvider(Commands)
}
```

### 4. Add Import File to CLI

```go
// go/cli/eac/imports_mygroup.go
//go:build !lite

package main

import _ "github.com/ready-to-release/eac/go/commands/mygroup"
```

### 5. Build and Test

```bash
# Rebuild CLI with new command
go build -o out/eac ./go/cli/eac

# Test command
./out/eac my command --help
```

---

## Command Types

EAC supports different command types via interface implementation:

### SimpleCommandPort (Most Common)

```go
type SimpleCommandPort interface {
    Name() string
    Metadata() core.CommandMetadata
    Execute(context.Context, *core.CommandRequest) int
}
```

**Use for**: Most commands that need flags, args, and structured metadata

### RunnableCommandPort (Advanced)

```go
type RunnableCommandPort interface {
    SimpleCommandPort
    Run(context.Context) error
}
```

**Use for**: Commands needing additional lifecycle methods

---

## Command Patterns

### Pattern 1: Simple Data Retrieval

```go
func (c *getModulesCommand) Execute(ctx context.Context, req *core.CommandRequest) int {
    modules, err := repository.LoadModules()
    if err != nil {
        log.Errorf("Failed to load modules: %v", err)
        return 1
    }

    // Output as JSON
    json.NewEncoder(os.Stdout).Encode(modules)
    return 0
}
```

### Pattern 2: Build/Test Orchestration

```go
func (c *buildCommand) Execute(ctx context.Context, req *core.CommandRequest) int {
    // Parse module arguments
    monikers := req.Args

    // Resolve components to UoWs
    uows, err := resolver.ResolveForBuild(monikers)
    if err != nil {
        return 1
    }

    // Execute UoWs with orchestrator
    orch := orchestrator.New(config, buildWorker)
    results, err := orch.Run(uows)
    if err != nil {
        return 1
    }

    return orchestrator.GetExitCode(results)
}
```

### Pattern 3: Validation Commands

```go
func (c *validateCommand) Execute(ctx context.Context, req *core.CommandRequest) int {
    validators := []Validator{
        &ContractsValidator{},
        &SpecsValidator{},
        &DependenciesValidator{},
    }

    failed := false
    for _, v := range validators {
        if err := v.Validate(); err != nil {
            log.Errorf("Validation failed: %v", err)
            failed = true
        }
    }

    if failed {
        return 1
    }
    log.Info("All validations passed")
    return 0
}
```

---

## Flag Handling

### Declaring Flags

Flags are declared in `Metadata()`:

```go
func (c *buildCommand) Metadata() core.CommandMetadata {
    return core.CommandMetadata{
        CanonicalName: "build",
        Flags: []core.FlagSpec{
            {
                Name:         "skip-cache",
                Type:         "bool",
                Usage:        "Skip build cache",
                DefaultValue: "false",
            },
            {
                Name:  "output",
                Type:  "string",
                Usage: "Output directory",
            },
        },
    }
}
```

### Parsing Flags

Use the `flags` package to parse:

```go
import "github.com/ready-to-release/eac/go/clibase/flags"

func (c *buildCommand) Execute(ctx context.Context, req *core.CommandRequest) int {
    skipCache := flags.GetBool("skip-cache")
    output := flags.GetString("output")

    log.Infof("Building with skipCache=%v, output=%s", skipCache, output)
    return 0
}
```

---

## Logging

Use the global logger from `go/clibase/logger`:

```go
import "github.com/ready-to-release/eac/go/clibase/logger"

var log = logger.Get()

func (c *command) Execute(ctx context.Context, req *core.CommandRequest) int {
    log.Info("Starting command")
    log.Debugf("Debug info: %v", data)
    log.Warnf("Warning: %v", issue)
    log.Errorf("Error: %v", err)
    return 0
}
```

---

## Exit Codes

Standard exit codes:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Validation failure |
| 3 | Build failure |
| 4 | Test failure |
| 5 | Scan found issues |

---

## Testing Commands

### Unit Tests

```go
// go/commands/mygroup/mycommand_test.go
func TestMyCommand(t *testing.T) {
    cmd := &myCommand{}

    ctx := context.Background()
    req := &core.CommandRequest{
        Args: []string{"arg1", "arg2"},
    }

    exitCode := cmd.Execute(ctx, req)
    assert.Equal(t, 0, exitCode)
}
```

### BDD Tests

```gherkin
# go/cli/eac/specs/commands/my-command.feature
Feature: My Command

  Scenario: Successfully execute command
    When I run "eac my command arg1 arg2"
    Then the exit code is 0
    And the output contains "Success"
```

---

## Best Practices

### 1. Keep Commands Focused

Each command should do one thing well. Complex workflows should be composed of multiple commands.

### 2. Use Consistent Naming

- Command names: lowercase with spaces (`build`, `test suite`)
- Canonical names: lowercase with hyphens (`build`, `test-suite`)
- Package names: lowercase, no hyphens (`build`, `testsuite`)

### 3. Fail Fast

Return early on errors:

```go
func (c *command) Execute(ctx context.Context, req *core.CommandRequest) int {
    if err := validate(); err != nil {
        log.Errorf("Validation failed: %v", err)
        return 1  // Exit early
    }

    // Continue with main logic
    return 0
}
```

### 4. Provide Helpful Error Messages

```go
log.Errorf("Failed to load module 'core': file not found at %s", path)
```

### 5. Support JSON Output

Commands should support `--json` flag for automation:

```go
if flags.GetBool("json") {
    json.NewEncoder(os.Stdout).Encode(result)
} else {
    fmt.Println(result.Format())
}
```

---

## Related Documentation

- [Commands Framework](./commands-framework.md) - Modular command architecture
- [Build Execution System](./build-execution.md) - UoW orchestration
- [Adapters System](./adapters-system.md) - External tool integration

---

## FAQ

**Q: Why use init() instead of explicit registration?**

A: Init-based registration enables automatic discovery and makes it easy to add new command modules without modifying the CLI main package. It also allows conditional compilation via build tags.

**Q: Can I have subcommands?**

A: Yes. Use compound names like `"serve design"` for subcommands. The command name includes the space.

**Q: How do I share code between commands?**

A: Put shared logic in the `commands/base/` package or create utility packages within your command module.

**Q: Can commands call other commands?**

A: No. Commands should be independent. Use shared libraries for common functionality.

**Q: How do I debug command registration?**

A: Add logging to `allcmds.BuildRegistry()` to see which providers are registered.
