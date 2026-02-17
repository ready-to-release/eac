# Adapters System

## Overview

The EAC adapters system provides a contract-based integration layer between EAC commands and external tools, test frameworks, and package managers. Adapters enable EAC to work with diverse technologies while maintaining a consistent interface.

## Architecture Principles

### 1. Contract-Based Integration

Adapters implement standardized contracts defined in the `contracts` module:

```text
┌─────────────────────────────────────┐
│       EAC Commands                  │
│  (build, test, scan, update)        │
└──────────────┬──────────────────────┘
               │ Uses contract interfaces
               ▼
┌─────────────────────────────────────┐
│         Contracts                   │
│  (runner, scanner, package-manager) │
└──────────────┬──────────────────────┘
               │ Implemented by
               ▼
┌─────────────────────────────────────┐
│         Adapters                    │
│  (gotest, npm, docker, etc.)        │
└──────────────┬──────────────────────┘
               │ Wraps
               ▼
┌─────────────────────────────────────┐
│      External Tools                 │
│  (go test, npm, docker, etc.)       │
└─────────────────────────────────────┘
```

### 2. Isolation and Independence

Each adapter is:

- **Self-contained**: No dependencies on other adapters
- **Independently testable**: Can be tested in isolation
- **Independently deployable**: Can be updated without affecting others
- **Stateless**: No persistent state between invocations

### 3. CLI Wrapping Pattern

Most adapters wrap existing CLI tools:

```go
// Example: npm adapter wraps npm CLI
type NpmAdapter struct {
    npmPath string
}

func (n *NpmAdapter) Install(ctx context.Context, pkg string) error {
    cmd := exec.CommandContext(ctx, n.npmPath, "install", pkg)
    return cmd.Run()
}
```

## Adapter Categories

### Test Framework Adapters

Implement the `runner` contract for test execution:

| Adapter      | Framework | Language   | Test Style       |
| ------------ | --------- | ---------- | ---------------- |
| **gotest**   | go test   | Go         | Unit/Integration |
| **godog**    | Godog     | Go         | BDD (Gherkin)    |
| **mocha**    | Mocha     | JavaScript | Unit/Integration |
| **pytest**   | pytest    | Python     | Unit/Integration |
| **behave**   | Behave    | Python     | BDD (Gherkin)    |
| **reqnroll** | Reqnroll  | .NET       | BDD (Gherkin)    |
| **cucumber** | Cucumber  | Ruby       | BDD (Gherkin)    |

**Common Interface**:

```go
type Runner interface {
    // Discover finds test suites in workspace
    Discover(ctx context.Context, workspace string) ([]TestSuite, error)

    // Execute runs tests and returns results
    Execute(ctx context.Context, suite TestSuite) (TestResults, error)
}
```

### Package Manager Adapters

Manage dependencies for different languages:

| Adapter    | Tool   | Language           | Purpose              |
| ---------- | ------ | ------------------ | -------------------- |
| **npm**    | npm    | JavaScript/Node.js | Package management   |
| **pip**    | pip    | Python             | Package installation |
| **nuget**  | NuGet  | .NET               | Package management   |
| **dotnet** | dotnet | .NET               | SDK and packages     |

**Common Interface**:

```go
type PackageManager interface {
    // Install installs a package
    Install(ctx context.Context, pkg string) error

    // Update updates packages
    Update(ctx context.Context) error

    // List lists installed packages
    List(ctx context.Context) ([]Package, error)
}
```

### Infrastructure Adapters

Integration with development infrastructure:

| Adapter    | Service          | Purpose                   |
| ---------- | ---------------- | ------------------------- |
| **docker** | Docker Engine    | Container runtime         |
| **gh**     | GitHub           | Version control and CI/CD |
| **ai**     | Anthropic/OpenAI | AI-powered automation     |

**Example: Container Runtime**:

```go
type ContainerRuntime interface {
    // Pull pulls a container image
    Pull(ctx context.Context, image string) error

    // Run creates and runs a container
    Run(ctx context.Context, config RunConfig) error

    // Build builds an image from Dockerfile
    Build(ctx context.Context, config BuildConfig) error
}
```

### Utility Adapters

Supporting adapters for special purposes:

| Adapter | Purpose                                          |
| ------- | ------------------------------------------------ |
| **tui** | Terminal UI components (spinners, progress bars) |
| **eac** | Nested EAC execution (EAC-to-EAC)                |

## Adapter Lifecycle

### 1. Initialization

Adapters are initialized at EAC startup:

```go
// Adapter factory
func NewGotestAdapter() (*GotestAdapter, error) {
    goPath, err := exec.LookPath("go")
    if err != nil {
        return nil, fmt.Errorf("go not found: %w", err)
    }

    return &GotestAdapter{
        goPath: goPath,
    }, nil
}
```

### 2. Registration

Adapters register with the framework:

```go
// Registry maintains available adapters
type AdapterRegistry struct {
    runners  map[string]Runner
    scanners map[string]Scanner
}

// Register test runner adapter
func (r *AdapterRegistry) RegisterRunner(name string, runner Runner) {
    r.runners[name] = runner
}
```

### 3. Invocation

Commands invoke adapters through contracts:

```go
// Command uses contract interface, not specific adapter
func runTests(runner Runner, workspace string) error {
    suites, err := runner.Discover(ctx, workspace)
    if err != nil {
        return err
    }

    for _, suite := range suites {
        results, err := runner.Execute(ctx, suite)
        if err != nil {
            return err
        }
        // Process results
    }
    return nil
}
```

## Error Handling

Adapters translate tool-specific errors to contract errors:

```go
// Tool-specific error
if exitCode != 0 {
    // Translate to contract error
    return &runner.TestFailedError{
        Suite:   suite.Name,
        Message: stderr.String(),
    }
}
```

## Output Parsing

Adapters parse tool output and convert to standard formats:

### Example: gotest Adapter

```go
func (g *GotestAdapter) Execute(ctx context.Context, suite TestSuite) (TestResults, error) {
    // Run: go test -json ./...
    cmd := exec.Command("go", "test", "-json", suite.Path)
    output, err := cmd.Output()

    // Parse JSON output
    var results TestResults
    for _, line := range bytes.Split(output, []byte("\n")) {
        var event GoTestEvent
        json.Unmarshal(line, &event)

        // Convert to standard TestResult
        results.Tests = append(results.Tests, TestResult{
            Name:     event.Test,
            Status:   mapStatus(event.Action),
            Duration: event.Elapsed,
        })
    }

    return results, nil
}
```

## Adapter Configuration

Adapters can be configured via `.eac/repository.yml`:

```yaml
adapters:
  gotest:
    enabled: true
    config:
      coverage: true
      race: true
      timeout: "5m"

  npm:
    enabled: true
    config:
      registry: "https://registry.npmjs.org"
      cache_dir: ".npm-cache"
```

## Testing Adapters

### Unit Tests

Test adapter logic in isolation:

```go
func TestGotestAdapter_Discover(t *testing.T) {
    adapter := &GotestAdapter{goPath: "/usr/bin/go"}

    suites, err := adapter.Discover(context.Background(), "/workspace")

    assert.NoError(t, err)
    assert.NotEmpty(t, suites)
}
```

### Integration Tests

Test adapter integration with actual tools:

```go
func TestGotestAdapter_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    adapter, _ := NewGotestAdapter()

    suite := TestSuite{
        Name: "core",
        Path: "./go/core",
    }

    results, err := adapter.Execute(context.Background(), suite)

    assert.NoError(t, err)
    assert.True(t, results.Passed > 0)
}
```

### Mock Adapters

Use mocks for testing commands without external dependencies:

```go
type MockRunner struct {
    DiscoverFunc func(ctx context.Context, workspace string) ([]TestSuite, error)
    ExecuteFunc  func(ctx context.Context, suite TestSuite) (TestResults, error)
}

func (m *MockRunner) Discover(ctx context.Context, workspace string) ([]TestSuite, error) {
    return m.DiscoverFunc(ctx, workspace)
}
```

## Adding New Adapters

To add a new adapter:

### 1. Identify Contract

Choose the appropriate contract:

- Test framework → `runner` contract
- Package manager → `package-manager` contract
- Scanner → `scanner` contract
- Container runtime → `container-runtime` contract

### 2. Create Module

```bash
mkdir -p go/adapters/myAdapter
```

### 3. Implement Contract

```go
// go/adapters/myAdapter/myAdapter.go
package myadapter

import "github.com/ready-to-release/eac/go/contracts/runner/0.1.0/runner"

type MyAdapter struct {
    toolPath string
}

func (a *MyAdapter) Discover(ctx context.Context, workspace string) ([]runner.TestSuite, error) {
    // Implementation
}

func (a *MyAdapter) Execute(ctx context.Context, suite runner.TestSuite) (runner.TestResults, error) {
    // Implementation
}
```

### 4. Register Adapter

```go
// Register in EAC startup
func initAdapters(registry *AdapterRegistry) error {
    myAdapter, err := myadapter.New()
    if err != nil {
        return err
    }

    registry.RegisterRunner("myadapter", myAdapter)
    return nil
}
```

### 5. Add Tests

```go
// go/adapters/myAdapter/myAdapter_test.go
func TestMyAdapter_Execute(t *testing.T) {
    // Unit tests
}
```

### 6. Document

Create `docs/reference/eac/modules/adapters/myadapter.md`

## Adapter Best Practices

### 1. Fail Fast

Validate prerequisites at initialization:

```go
func New() (*MyAdapter, error) {
    toolPath, err := exec.LookPath("mytool")
    if err != nil {
        return nil, fmt.Errorf("mytool not found: %w", err)
    }

    return &MyAdapter{toolPath: toolPath}, nil
}
```

### 2. Respect Context

Honor context cancellation:

```go
func (a *MyAdapter) Execute(ctx context.Context, suite TestSuite) (TestResults, error) {
    cmd := exec.CommandContext(ctx, a.toolPath, "test")
    // Context cancellation automatically kills the process
    return cmd.Run()
}
```

### 3. Log Appropriately

Use structured logging:

```go
log.Info("running tests",
    "adapter", "gotest",
    "suite", suite.Name,
    "path", suite.Path,
)
```

### 4. Handle Errors Gracefully

Provide helpful error messages:

```go
if err != nil {
    return fmt.Errorf("failed to run %s tests in %s: %w",
        suite.Name, suite.Path, err)
}
```

## Performance Considerations

### Parallel Execution

Adapters should support parallel execution:

```go
func (a *MyAdapter) Execute(ctx context.Context, suite TestSuite) (TestResults, error) {
    if suite.Parallel {
        // Run tests in parallel
        return a.runParallel(ctx, suite)
    }
    return a.runSequential(ctx, suite)
}
```

### Caching

Leverage tool caching when available:

```go
// npm adapter uses npm cache
cmd := exec.Command("npm", "install", "--prefer-offline")
```

### Timeouts

Implement reasonable timeouts:

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
defer cancel()
```

## See Also

- [Adapters Module](../modules/adapters.md) - Adapter module documentation
- [Contracts Module](../modules/contracts.md) - Contract definitions
- [Commands Module](../modules/commands.md) - Commands using adapters
