# Adapters Overview

The adapters module group provides integration with external tools, test frameworks, and package managers. Adapters translate between EAC's contract-based interfaces and the specific APIs and CLIs of external tools.

## Purpose

Adapters enable EAC to:

- **Execute tests** across multiple languages and frameworks (Go, JavaScript, Python, .NET, Ruby)
- **Manage packages** using language-specific package managers (npm, pip, NuGet)
- **Orchestrate containers** via Docker and container runtimes
- **Integrate with GitHub** for repository operations
- **Leverage AI providers** for intelligent automation
- **Provide interactive UIs** via terminal user interfaces

## Modules

The adapter system consists of the following adapters:

| Adapter      | Purpose                                     | Technology                |
| ------------ | ------------------------------------------- | ------------------------- |
| **ai**       | AI provider integration (Claude, GPT, etc.) | Anthropic API, OpenAI API |
| **behave**   | Python BDD test framework integration       | Behave (Gherkin)          |
| **cucumber** | Ruby BDD test framework integration         | Cucumber (Gherkin)        |
| **docker**   | Container runtime integration               | Docker Engine API         |
| **dotnet**   | .NET build and package management           | dotnet CLI                |
| **eac**      | EAC-to-EAC adapter for nested commands      | EAC CLI                   |
| **gh**       | GitHub API and CLI integration              | GitHub CLI (gh)           |
| **godog**    | Go BDD test framework integration           | Godog (Gherkin)           |
| **gotest**   | Go native test framework integration        | go test                   |
| **mocha**    | JavaScript test framework integration       | Mocha                     |
| **npm**      | Node.js package management                  | npm                       |
| **nuget**    | .NET package management                     | NuGet                     |
| **pip**      | Python package management                   | pip                       |
| **pytest**   | Python test framework integration           | pytest                    |
| **reqnroll** | .NET BDD test framework integration         | Reqnroll (Gherkin)        |
| **tui**      | Terminal user interface components          | Bubble Tea                |

## Architecture

Adapters implement contract-based interfaces defined in the `contracts` module:

```text
┌─────────────────────────────────────┐
│         EAC Commands                │
│    (build, test, scan, update)      │
└──────────────┬──────────────────────┘
               │
    ┌──────────▼──────────┐
    │   contracts         │
    │ (Interfaces/Schemas)│
    └──────────┬──────────┘
               │
     ┌─────────┼─────────┐
     │         │         │
┌────▼───┐ ┌──▼───┐ ┌──▼────┐
│ gotest │ │ npm  │ │docker │  ...
│adapter │ │adapter│ │adapter│
└────┬───┘ └──┬───┘ └──┬────┘
     │        │        │
┌────▼────────▼────────▼────┐
│   External Tools/CLIs     │
│ (go test, npm, docker)    │
└───────────────────────────┘
```

### Key Design Principles

1. **Contract-Based**: Adapters implement standardized contracts (runner, scanner, package manager)
2. **Isolation**: Each adapter is independent and self-contained
3. **CLI Wrapping**: Most adapters wrap existing CLI tools
4. **Error Translation**: Adapters translate tool-specific errors to contract errors

### Adapter Interface Example

```go
// Example: Runner contract (implemented by test adapters)
type Runner interface {
    // Discover finds test suites in the workspace
    Discover(ctx context.Context, workspace string) ([]TestSuite, error)

    // Execute runs tests and returns results
    Execute(ctx context.Context, suite TestSuite) (TestResults, error)
}
```

## Adapter Categories

### Test Framework Adapters

Adapters for running tests:

- **Go**: gotest (native), godog (BDD)
- **JavaScript**: mocha
- **Python**: pytest, behave (BDD)
- **.NET**: reqnroll (BDD)
- **Ruby**: cucumber (BDD)

### Package Manager Adapters

Adapters for dependency management:

- **JavaScript/Node.js**: npm
- **Python**: pip
- **.NET**: nuget, dotnet

### Infrastructure Adapters

Adapters for external services:

- **Containers**: docker (Docker Engine)
- **Version Control**: gh (GitHub)
- **AI Services**: ai (Claude, GPT)

### Utility Adapters

Supporting adapters:

- **UI**: tui (terminal interfaces)
- **Nested Execution**: eac (EAC-to-EAC)

## Contract Implementation

Adapters implement one or more contracts:

| Adapter                       | Contracts Implemented          |
| ----------------------------- | ------------------------------ |
| gotest, mocha, pytest, behave | `runner` (test execution)      |
| npm, pip, nuget, dotnet       | `package-manager`              |
| docker                        | `container-runtime`            |
| ai                            | `ai-provider`                  |
| gh                            | `vcs` (version control system) |
| tui                           | `tui` (terminal UI)            |

## Adding New Adapters

To add a new adapter:

1. **Create module**: `go/adapters/{name}/`
2. **Implement contract**: Implement required contract interfaces
3. **Add registration**: Register adapter in EAC startup
4. **Write tests**: Add adapter tests
5. **Document**: Create `docs/reference/eac/modules/adapters/{name}.md`

See [Adapters System](../architecture/adapters-system.md) for detailed architecture guidance.

## Dependencies

### Common Dependencies

All adapters depend on:

- **contracts**: Contract definitions
- **core**: Core utilities and configuration
- **clibase**: CLI utilities and error handling

### Adapter-Specific Dependencies

- **Test adapters**: May depend on language-specific SDKs
- **Package managers**: Depend on CLI tools (npm, pip, etc.)
- **Infrastructure adapters**: Depend on external SDKs (Docker SDK, GitHub API)

## See Also

- [Contracts Module](contracts.md) - Contract system documentation
- [EAC Architecture](../architecture/index.md) - Overall architecture
- [Modules Index](../index.md) - Complete module reference
