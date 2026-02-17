# Contracts Overview

The contracts system defines versioned JSON schemas that specify the interfaces between EAC modules, adapters, and external tools. Contracts enable loose coupling, testability, and independent evolution of components.

## Purpose

Contracts provide:

- **Interface Definitions**: Standard interfaces for adapters, runners, scanners, and tools
- **Validation Schemas**: JSON Schema definitions for configuration and data structures
- **Default Values**: Sensible defaults for contract implementations
- **Version Management**: Semantic versioning for backward compatibility

## Contract Modules

The contract system consists of the following contracts:

| Contract              | Purpose                                              | Location                       | Version |
| --------------------- | ---------------------------------------------------- | ------------------------------ | ------- |
| **ai-provider**       | AI provider integration interface (Claude, GPT)      | `contracts/ai-provider/`       | 0.1.0   |
| **clie**              | CLIE CLI framework configuration and interfaces      | `contracts/clie/`              | 0.1.0   |
| **container-runtime** | Container runtime interface (Docker, Podman)         | `contracts/container-runtime/` | 0.1.0   |
| **core**              | Core configuration and environment contracts         | `contracts/core/`              | 0.1.0   |
| **docs**              | Documentation generation and rendering contracts     | `contracts/docs/`              | 0.1.0   |
| **runner**            | Test runner interface (test frameworks)              | `contracts/runner/`            | 0.1.0   |
| **scanner**           | Security scanner interface (vulnerability detection) | `contracts/scanner/`           | 0.1.0   |
| **tui**               | Terminal UI component interface                      | `contracts/tui/`               | 0.1.0   |

## Architecture

Contracts sit between commands and adapters:

```text
┌─────────────────────────────────────┐
│         EAC Commands                │
│    (build, test, scan, update)      │
└──────────────┬──────────────────────┘
               │
    ┌──────────▼──────────┐
    │     Contracts       │
    │  (JSON Schemas +    │
    │   Go Interfaces)    │
    └──────────┬──────────┘
               │
     ┌─────────┼─────────┐
     │         │         │
┌────▼───┐ ┌──▼───┐ ┌──▼────┐
│Adapter1│ │Adapter2│ │Adapter3│
└────────┘ └──────┘ └────────┘
```

### Key Design Principles

1. **Interface Segregation**: Contracts define minimal, focused interfaces
2. **Schema Validation**: JSON Schema validates configuration and data
3. **Semantic Versioning**: Contracts evolve with backward compatibility
4. **Default Values**: Contracts provide sensible defaults to reduce configuration burden

## Contract Structure

Each contract follows a consistent directory structure:

```text
contracts/{contract-name}/
├── 0.1.0/                          # Version directory
│   ├── {contract-name}.schema.json # JSON Schema definition
│   ├── defaults/                   # Default values
│   │   ├── base.yml               # Base defaults
│   │   └── ...                    # Environment-specific defaults
│   └── examples/                   # Example configurations (optional)
│       └── ...
└── README.md                       # Contract documentation
```

### Example: Core Contract

```text
contracts/core/
├── 0.1.0/
│   ├── core.schema.json           # Defines workspace, environment structure
│   ├── defaults/
│   │   ├── base.yml              # Default workspace settings
│   │   └── ci.yml                # CI environment overrides
└── README.md
```

## Versioning Strategy

Contracts use semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes (incompatible schema changes)
- **MINOR**: New features (backward-compatible additions)
- **PATCH**: Bug fixes (clarifications, documentation)

### Current Version

All contracts are currently at **version 0.1.0**, indicating pre-release stability.

### Version Evolution

When contracts evolve:

1. **Backward-compatible changes**: Increment MINOR version (0.1.0 → 0.2.0)
2. **Breaking changes**: Increment MAJOR version (0.1.0 → 1.0.0)
3. **Multiple versions**: Old versions remain available for compatibility

Example:

```text
contracts/runner/
├── 0.1.0/          # Legacy version
├── 0.2.0/          # Current version (backward-compatible)
└── README.md
```

## Contract Categories

### Infrastructure Contracts

Contracts for external infrastructure:

- **container-runtime**: Docker, Podman integration
- **clie**: CLIE CLI framework

### Tool Integration Contracts

Contracts for tool adapters:

- **runner**: Test framework integration
- **scanner**: Security scanner integration
- **ai-provider**: AI service integration

### Core Contracts

Foundational contracts:

- **core**: Workspace, environment, configuration
- **docs**: Documentation generation
- **tui**: Terminal UI components

## JSON Schema Validation

Contracts use JSON Schema for validation:

```go
import "github.com/ready-to-release/eac/go/core/contracts"

// Load contract schema
schema, err := contracts.LoadSchema("runner", "0.1.0")
if err != nil {
    return err
}

// Validate configuration
if err := schema.Validate(config); err != nil {
    return fmt.Errorf("invalid runner config: %w", err)
}
```

## Default Values

Contracts provide default values to simplify configuration:

```yaml
# contracts/core/0.1.0/defaults/base.yml
workspace:
  root: "."
  ignore:
    - ".git"
    - "node_modules"
    - ".clie"

environment:
  type: "local"
  platform: "linux"
```

Defaults are loaded automatically and merged with user configuration.

## Using Contracts

### In Commands

Commands depend on contract interfaces:

```go
import "github.com/ready-to-release/eac/go/contracts/runner/0.1.0/runner"

func runTests(r runner.Runner) error {
    suites, err := r.Discover(ctx, workspace)
    if err != nil {
        return err
    }

    results, err := r.Execute(ctx, suites[0])
    return err
}
```

### In Adapters

Adapters implement contract interfaces:

```go
import "github.com/ready-to-release/eac/go/contracts/runner/0.1.0/runner"

type GotestAdapter struct {}

func (g *GotestAdapter) Discover(ctx context.Context, workspace string) ([]runner.TestSuite, error) {
    // Implementation
}

func (g *GotestAdapter) Execute(ctx context.Context, suite runner.TestSuite) (runner.TestResults, error) {
    // Implementation
}
```

## Contract Validation

Validate contracts with:

```bash
# Validate all contract schemas
clie eac validate contracts

# Validate a specific contract
clie eac validate contracts --contract runner
```

## See Also

- [Core Module](core.md) - Core module documentation
- [EAC Architecture](../architecture/index.md) - Overall architecture
- [Modules Index](../index.md) - Complete module reference
