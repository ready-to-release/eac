# Commands Overview

The commands module provides the modular command framework for EAC. The monolithic commands module has been refactored into seven specialized command components within a single module, each focused on a specific domain.

## Components

The commands module consists of the following components:

| Module         | Purpose                                                        | Location                  |
| -------------- | -------------------------------------------------------------- | ------------------------- |
| **base**       | Base command infrastructure and shared utilities               | `go/commands/base/`       |
| **build**      | Build automation commands (compile, package, containerize)     | `go/commands/build/`      |
| **lint**       | Code quality and linting commands                              | `go/commands/lint/`       |
| **repository** | Repository management commands (git workflows, branching)      | `go/commands/repository/` |
| **scan**       | Security scanning and vulnerability detection                  | `go/commands/scan/`       |
| **test**       | Test execution and test management commands                    | `go/commands/test/`       |
| **update**     | Update and maintenance commands (dependencies, docs, evidence) | `go/commands/update/`     |

## Architecture

The commands framework follows a modular architecture:

```text
┌─────────────────────────────────────────────┐
│              EAC CLI / MCP Server           │
└─────────────────┬───────────────────────────┘
                  │
    ┌─────────────┴─────────────┐
    │    commands/base          │
    │  (Command Registration)   │
    └─────────────┬─────────────┘
                  │
      ┌───────────┼───────────┐
      │           │           │
┌─────▼──┐  ┌────▼───┐  ┌───▼────┐
│ build  │  │  test  │  │ update │  ...
│commands│  │commands│  │commands│
└────────┘  └────────┘  └────────┘
      │           │           │
      └───────────┼───────────┘
                  │
    ┌─────────────▼─────────────┐
    │     core / contracts      │
    └───────────────────────────┘
```

### Key Design Principles

1. **Domain Separation**: Each command module focuses on a specific domain (build, test, scan, etc.)
2. **Shared Base**: Common command infrastructure lives in `commands/base`
3. **Independent Registration**: Each module registers its own commands
4. **Minimal Coupling**: Modules depend only on `commands/base`, `core`, and `contracts`

### Command Registration

Commands are registered using the base command framework:

```go
// Each command module provides a Register function
func Register(registry *base.CommandRegistry) error {
    registry.Add(&base.Command{
        Name:    "build",
        Handler: handleBuild,
        // ...
    })
    return nil
}
```

The EAC CLI and MCP server call each module's `Register` function at startup.

## Dependencies

### Common Dependencies

All command modules depend on:

- **commands/base**: Command registration and execution framework
- **core**: Core utilities (contracts, configuration, workspace detection)
- **contracts**: Contract definitions and schemas
- **clibase**: CLI framework (Cobra integration, flags, output)

### Module-Specific Dependencies

Individual command modules may also depend on:

- **Adapters**: Test runners, package managers, container runtimes
- **External Tools**: Git, Docker, language-specific tools

## Command Discovery

To list all available commands:

```bash
# List all EAC commands
eac --help

# List commands in a specific category
eac build --help
eac test --help
```

## Module Documentation

Individual command module documentation will be available at:

- `commands/base.md` - Base command infrastructure
- `commands/build.md` - Build commands
- `commands/lint.md` - Lint commands
- `commands/repository.md` - Repository commands
- `commands/scan.md` - Scan commands
- `commands/test.md` - Test commands
- `commands/update.md` - Update commands

## See Also

- [EAC Architecture](../architecture/index.md) - Overall EAC architecture
- [Modules Index](../index.md) - Complete module reference
- [Commands Reference](../commands/index.md) - Command usage documentation
