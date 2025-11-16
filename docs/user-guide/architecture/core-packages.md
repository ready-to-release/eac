# Core Packages Architecture

This document provides a comprehensive overview of the `src/core` packages that form the foundation of the EAC (Enterprise Architecture Contracts) system.

## Overview

The `src/core` directory contains the core domain logic, contract definitions, and shared functionality used throughout the system. These packages are organized by domain responsibility and follow clean architecture principles.

## Package Structure

```text
src/core/
├── contracts/          # Contract definitions and management
│   ├── modules/        # Module contract definitions
│   └── reports/        # Contract reporting and aggregation
├── environments/       # Environment contract definitions
├── repository/         # Repository file operations and discovery
└── testing/           # Test discovery, inference, and validation
```

## Core Packages

### 1. `contracts/modules` - Module Contract System

**Purpose**: Define and manage module contracts that describe the architecture's modular structure.

**Key Types**:

- `ModuleContract` - Represents a single module's contract
- `Registry` - In-memory registry for fast module lookups
- `BaseContract` - Common contract fields (moniker, name, type, etc.)

**Responsibilities**:

- Load module contracts from YAML files
- Validate module contract structure
- Provide module lookup by moniker
- Generate glob patterns for file matching

**Key Functions**:

```go
// Load all module contracts for a version
func LoadModuleContracts(version string) (*Registry, error)

// Get a module by moniker
func (r *Registry) Get(moniker string) (*ModuleContract, bool)

// Get all modules of a specific type
func (r *Registry) GetByType(moduleType string) []*ModuleContract
```

**Contract Location**: `contracts/modules/<version>/*.yml`

**Example Contract**:

```yaml
moniker: src-cli
name: "CLI Application"
type: go-cli
description: "Command-line interface for the system"
```

---

### 2. `contracts/reports` - Contract Reporting

**Purpose**: Generate reports and aggregations from contracts for use by commands.

**Key Types**:

- `ModuleReport` - Aggregated module contract data
- `ModuleFileOwnership` - Mapping of files to owning modules

**Responsibilities**:

- Aggregate contract data for reporting
- Build file-to-module mappings
- Provide canonical contract data to commands

**Key Functions**:

```go
// Get module contracts with registry
func GetModuleContracts(repoRoot, version string) (*ModuleReport, error)

// Get file ownership information
func GetFileOwnership(repoRoot, version string) (map[string][]string, error)
```

**Usage Pattern**:

```go
// Commands use reports to get structured data
report, err := contractsreports.GetModuleContracts(repoRoot, "0.1.0")
registry := report.Registry  // Use for lookups
```

---

### 3. `environments` - Environment Contracts

**Purpose**: Define test execution environments across L0-L4 levels.

**Key Types**:

- `Environment` - Single environment definition
- `EnvironmentContract` - Collection of all environments
- `Metadata` - Contract metadata (version, description)

**Responsibilities**:

- Load environment contracts from embedded YAML
- Validate environment definitions
- Provide environment lookup by moniker
- Generate test tags (@env:\<moniker>)

**Key Functions**:

```go
// Load environment contract
func LoadEnvironmentContract() (*EnvironmentContract, error)

// Get specific environment
func (c *EnvironmentContract) GetEnvironment(moniker string) (*Environment, error)

// Get test tag for environment
func (e *Environment) GetTestTag() string  // Returns "@env:<moniker>"
```

**Environment Levels**:

- **L0**: Very fast unit tests (l00-01, l00-02)
- **L1**: Fast unit tests (l01-01, l01-02)
- **L2**: Local/Docker (local01, local02)
- **L3**: PLTE (plte01, plte02)
- **L4**: Production (production, production-inactive)

**Contract Location**: `src/core/environments/contracts/environments/0.1.0/environments.yml`

**Example Environment**:

```yaml
moniker: local01
name: "Local Environment 01 - Docker Container"
level: "L2"
type: "docker"
env_tags:
  - "local"
  - "isolated"
system_deps:
  - "@deps:docker"
```

---

### 4. `repository` - Repository Operations

**Purpose**: Discover and manage repository files and their relationships to modules.

**Key Types**:

- `FileInfo` - File metadata with module ownership
- `RepositoryFile` - File with additional context

**Responsibilities**:

- Walk repository directory structure
- Match files to modules using contracts
- Provide file listings with ownership
- Support filtered queries (staged, changed, etc.)

**Key Functions**:

```go
// Get all files with module ownership
func GetRepositoryFilesWithModules(
    includeAll, stagedOnly, changedOnly bool,
    repoRoot, version string,
) ([]FileInfo, error)

// Get changed modules based on git status
func GetChangedModules(repoRoot, version string) ([]string, error)
```

**Usage Scenarios**:

- `show files` - Display all repository files
- `show files staged` - Show only staged files
- `show files changed` - Show only modified files
- `get changed modules` - Determine affected modules

---

### 5. `testing` - Test System

**Purpose**: Comprehensive test discovery, inference, validation, and suite management.

This is the largest and most complex core package, handling the entire test lifecycle.

#### 5.1 Test Discovery

**Files**: `discovery.go`

**Responsibilities**:

- Discover Go tests (`*_test.go` files)
- Discover Godog/Gherkin tests (`*.feature` files)
- Parse test metadata and tags
- Build test references with file locations

**Key Functions**:

```go
// Discover all tests in repository
func DiscoverAllTests(rootPath string) ([]TestReference, error)

// Discover Go tests in a package
func DiscoverGoTestTags(pkgPath string) ([]TestReference, error)

// Discover Godog features
func DiscoverGodogFeatures(specsPath string) ([]TestReference, error)
```

**TestReference Structure**:

```go
type TestReference struct {
    FilePath  string   // Absolute path to test file
    Type      string   // "gotest" or "godog"
    TestName  string   // Test/scenario name
    Tags      []string // All tags (levels, deps, verification, etc.)
    IsIgnored bool     // @ignore tag present
    SkipReason string  // From @skip:<reason>
    // ... additional metadata
}
```

#### 5.2 Test Inference

**Files**: `inference.go`

**Responsibilities**:

- Apply inference rules to enrich test tags
- Infer levels based on test type and verification tags
- Infer system dependencies from module dependencies
- Infer system dependencies from environment tags
- Derive operational verification (@ov) when no other verification present

**Inference Pipeline**:

```text
1. ApplyInferences() - Apply suite-specific or global inferences
   ↓
2. InferSystemDepsFromModuleDeps() - Module type → system deps
   ↓
3. InferSystemDepsFromEnv() - Environment → system deps
   ↓
4. DeriveOperationalVerification() - Add @ov if needed
```

**Example Inferences**:

- `gotest` + no level → Add `@L1`
- `godog` + no level → Add `@L2`
- `@iv` → Add `@L3`
- `@piv` → Add `@L4`
- `@depm:src-cli` + go-cli module → Add `@deps:go`
- `@env:local01` → Add `@deps:docker`

**Key Functions**:

```go
// Apply inference rules
func ApplyInferences(tests []TestReference, inferences []Inference) []TestReference

// Infer from module dependencies
func InferSystemDepsFromModuleDeps(tests []TestReference, registry *modules.Registry) []TestReference

// Infer from environment tags
func InferSystemDepsFromEnv(tests []TestReference, envContract *environments.EnvironmentContract) []TestReference

// Derive @ov tag
func DeriveOperationalVerification(tags []string) []string
```

#### 5.3 Test Suites

**Files**: `suites.go`, `suite_report.go`

**Responsibilities**:

- Load test suite definitions
- Select tests matching suite criteria
- Generate suite reports with full metadata
- Support multiple selectors (AnyOf, AllOf, Exclude)

**Key Types**:

```go
type TestSuite struct {
    Moniker     string        // Suite identifier
    Name        string        // Human-readable name
    Description string        // Purpose
    Selectors   []TagSelector // Selection criteria
    Inferences  []Inference   // Suite-specific inferences
}

type SuiteReport struct {
    SuiteMoniker      string
    SuiteName         string
    Description       string
    ProductionTests   []SuiteTestEntry  // Tests for production use
    FrameworkTests    []SuiteTestEntry  // Framework/internal tests
    TotalDiscovered   int
    ValidationErrors  map[string][]string
}

type SuiteTestEntry struct {
    Moniker          string   // Unique test identifier
    TestName         string   // Original test name
    Type             string   // gotest/godog
    Module           string   // Owning module
    ModuleType       string   // Module type
    Level            []string // @L0-@L4 tags
    Verification     []string // @ov/@iv/@pv/@piv/@ppv
    SystemDeps       []string // @deps:* tags
    ModuleDeps       []string // @depm:* tags
    // ... additional fields
}
```

**Selection Logic**:

```go
// Select tests matching suite criteria
func (s *TestSuite) SelectTests(allTests []TestReference) []TestReference

// Check if test matches selector
func (s *TagSelector) Matches(tags []string) bool
```

**Predefined Suites**:

- `commit` - Pre-commit checks (L0-L2)
- `acceptance` - PLTE acceptance tests (@iv, @ov, @pv)
- `production-verification` - Production verification (@L4 + @piv)

#### 5.4 Test Validation

**Files**: `validation.go`

**Responsibilities**:

- Validate test tags against contracts
- Check for multiple level tags
- Check for multiple verification tags
- Validate GxP requirements
- Validate risk controls

**Validation Rules**:

- Must have exactly one level tag (@L0-@L4)
- Must have exactly one verification tag (@ov/@iv/@pv/@piv/@ppv)
- @gxp tests must have @risk-control:gxp-\*
- @critical-aspect must be used with @gxp
- All tags must be defined in tags contract

**Key Functions**:

```go
// Validate single test
func ValidateTestReference(test TestReference, repoRoot string) []string

// Validate all tests
func ValidateAllPostInference(tests []TestReference, repoRoot string) map[string][]string

// Validate GxP requirements
func ValidateGxPRequirements(test TestReference) []string
```

#### 5.5 Test Monikers

**Files**: `moniker.go`

**Responsibilities**:

- Generate unique identifiers for tests
- Different formats for gotest vs godog
- Stable, human-readable names

**Moniker Formats**:

- **GoTest**: `module_test-file_TestName`
  - Example: `src-cli_install-test_test-install-command-create-config-file`
- **Godog**: `module_feature-name_scenario-name`
  - Example: `src-cli_cli-invocation_version-flag-displays-version`

**Key Functions**:

```go
// Generate test moniker
func GenerateTestMoniker(testRef TestReference, module string) string

// Convert to kebab-case
func toKebabCase(s string) string
```

#### 5.6 GxP and Risk Controls

**Files**: `gxp.go`, `traceability.go`

**Responsibilities**:

- Filter GxP tests
- Generate traceability matrices
- Link tests to risk controls
- Generate GxP compliance reports

**Key Types**:

```go
type TraceabilityEntry struct {
    TestName     string
    RiskControl  string
    Level        string
    Verification string
}

type GxPReport struct {
    TotalGxPTests      int
    CriticalAspects    int
    RiskControlCoverage map[string][]string
}
```

---

## Common Patterns

### Pattern 1: Contract Loading

```go
// Load → Validate → Use
contract, err := LoadSomethingContract()
if err := contract.Validate(); err != nil {
    // Handle error
}
// Use contract
```

### Pattern 2: Registry Pattern

```go
// Build registry → Fast lookups
registry := BuildRegistry(contracts)
module, exists := registry.Get(moniker)
```

### Pattern 3: Pipeline Pattern

```go
// Discover → Infer → Select → Validate
tests := DiscoverAllTests(root)
tests = ApplyInferences(tests, inferences)
tests = InferSystemDepsFromModuleDeps(tests, registry)
selected := suite.SelectTests(tests)
errors := ValidateAll(selected, root)
```

### Pattern 4: Test Entry Conversion

```go
// TestReference (raw) → SuiteTestEntry (enriched)
entry := SuiteTestEntry{
    Moniker: GenerateTestMoniker(test, module),
    TestName: test.TestName,
    // ... extract and categorize tags
}
```

---

## Data Flow

```text
Repository Files
      ↓
[Discovery] ← Contracts (modules, environments, tags)
      ↓
TestReferences (raw test data)
      ↓
[Inference] ← Module Registry, Environment Contract
      ↓
TestReferences (enriched with inferred tags)
      ↓
[Suite Selection] ← Suite Definition
      ↓
Selected Tests
      ↓
[Validation] ← Tag Contracts
      ↓
[Report Generation] ← File-Module Map
      ↓
SuiteTestEntries (final structured data)
      ↓
Commands (get/show) → User
```

---

## Dependency Graph

```text
commands/
    ↓
contracts/reports
    ↓
contracts/modules

commands/
    ↓
testing/
    ↓
├── contracts/modules (for module lookups)
├── environments/ (for env lookups)
└── repository/ (for file discovery)

environments/ (standalone)
repository/ (minimal dependencies)
```

---

## Extension Points

### Adding New Test Types

1. Update `discovery.go` with new discovery function
2. Update `inference.go` with type-specific inferences
3. Update `moniker.go` with new moniker format
4. Update tag contracts

### Adding New Inference Rules

1. Add to `GetGlobalInferences()` in `inference.go`
2. Or add to suite-specific inferences in YAML
3. Rules apply in order during pipeline

### Adding New Environments

1. Edit `contracts/environments/0.1.0/environments.yml`
2. Add new environment with moniker, level, type, deps
3. System automatically infers deps when `@env:<moniker>` used

### Adding New Validation Rules

1. Add validation function in `validation.go`
2. Call from `ValidateTestReference()`
3. Return error strings for reporting

---

## Testing the Core Packages

Each core package has comprehensive unit tests:

- `contracts/modules/*_test.go` - Contract loading and validation
- `environments/contracts_test.go` - Environment contract tests
- `testing/discovery_test.go` - Test discovery
- `testing/inference_test.go` - Inference rules
- `testing/validation_test.go` - Validation rules
- `testing/suites_test.go` - Suite selection

Run tests:

```bash
cd src/core/testing
go test -v

cd src/core/environments
go test -v
```

---

## Performance Considerations

1. **Registry Pattern**: O(1) module lookups via map
2. **Embedded Contracts**: Compiled into binary, no file I/O at runtime
3. **Pipeline**: Single pass through tests, applying all transformations
4. **Lazy Loading**: Reports generated on-demand, not cached globally
5. **Parallel Discovery**: Can discover Go tests and Godog tests concurrently

---

## Best Practices

1. **Immutability**: TestReferences are copied when enriched, never mutated
2. **Contract-Driven**: All behavior defined by contracts, not hardcoded
3. **Single Responsibility**: Each package has one clear purpose
4. **Type Safety**: Strong typing throughout, minimal use of `interface{}`
5. **Error Handling**: Errors bubble up with context, no silent failures
6. **Validation**: Validate early and often, fail fast with clear messages

---

## See Also

- [Get Commands Reference](../commands/get-commands.md)
- [Show Commands Reference](../commands/show-commands.md)
- [Module Contracts](../contracts/modules.md)
- [Environment Contracts](../contracts/environments.md)
- [Test Tagging System](../specifications/test-tagging.md)
