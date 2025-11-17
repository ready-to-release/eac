# Get Commands Reference

The `get` commands provide **canonical structured data** in machine-readable formats (YAML/JSON/TOML) for automation and integration.

## Architecture Principle

```sh
get commands  # CANONICAL DATA SOURCE
show commands # HUMAN-READABLE DISPLAY (calls get logic)
```

All `get` commands follow a consistent pattern:

1. Generate structured data from contracts/discovery
2. Support multiple output formats
3. Provide stable, automation-friendly output
4. No side effects (read-only operations)

---

## Common Flags

All `get` commands support these flags:

| Flag        | Description    | Default   |
| ----------- | -------------- | --------- |
| `--as-yaml` | Output as YAML | (default) |
| `--as-json` | Output as JSON |           |
| `--as-toml` | Output as TOML |           |

---

## Command Index

| Command                                     | Purpose                             |
| ------------------------------------------- | ----------------------------------- |
| [get modules](#get-modules)                 | Get all module contracts            |
| [get dependencies](#get-dependencies)       | Get module dependency graph         |
| [get execution order](#get-execution-order) | Get build/test execution order      |
| [get files](#get-files)                     | Get repository files with ownership |
| [get changed modules](#get-changed-modules) | Get modules affected by changes     |
| [get environments](#get-environments)       | Get all environment contracts       |
| [get suite](#get-suite)                     | Get test suite with all tests       |
| [get tests](#get-tests)                     | Get all tests in repository         |

---

## Module Commands

### `get modules`

**Purpose**: Get all module contracts as structured data.

**Usage**:

```bash
# YAML output (default)
get modules
# JSON output
get modules --as-json

# TOML output
get modules --as-toml
```

**Output Structure**:

```yaml
- moniker: src-cli
  name: "CLI Application"
  type: go-cli
  description: "Command-line interface"
  version: "0.1.0"
  dependencies: []

- moniker: src-commands
  name: "Commands Module"
  type: go-commands
  dependencies:
    - src-core

- moniker: src-core
  name: "Core Library"
  type: go-library
  dependencies: []
```

**Use Cases**:

- Build automation scripts
- Generate dependency graphs
- Validate module structure
- CI/CD pipeline configuration
- Module discovery for tooling

**Related Commands**:

- `show modules` - Human-readable table format
- `show moduletypes` - Group modules by type

---

### `get dependencies`

**Purpose**: Get module dependency graph in structured format.

**Usage**:

```bash
get dependencies
get dependencies --as-json
```

**Output Structure**:

```yaml
dependencies:
  src-cli:
    depends_on:
      - src-core
    dependents: []

  src-commands:
    depends_on:
      - src-core
    dependents: []

  src-core:
    depends_on: []
    dependents:
      - src-cli
      - src-commands
```

**Use Cases**:

- Visualize dependency graph
- Detect circular dependencies
- Plan refactoring efforts
- Determine build order
- Impact analysis

**Related Commands**:

- `show dependencies` - Formatted table view
- `get execution order` - Ordered module list

---

### `get execution order`

**Purpose**: Get correct build/test execution order based on dependencies.

**Usage**:

```bash
# All modules in dependency order
get execution order

# Specific modules only
get execution order src-cli src-commands
```

**Output Structure**:

```yaml
execution_order:
  - src-core        # No dependencies, runs first
  - src-commands    # Depends on src-core
  - src-cli         # Depends on src-core
```

**Use Cases**:

- Build scripts (build in correct order)
- Test execution (test dependencies first)
- Deployment pipelines
- Incremental builds
- Parallel execution planning

**Algorithm**: Topological sort with dependency resolution

**Related Commands**:

- `get dependencies` - Full dependency graph
- `get changed modules` - Affected modules only

---

### `get files`

**Purpose**: Get all repository files with module ownership information.

**Usage**:

```bash
# All files
get files

# Specific formats
get files --as-json
```

**Output Structure**:

```yaml
files:
  - path: "src/cli/main.go"
    modules:
      - src-cli
    status: "tracked"

  - path: "src/core/testing/suites.go"
    modules:
      - src-core
    status: "tracked"

  - path: "specs/src-cli/cli-invocation/specification.feature"
    modules:
      - src-cli
    status: "tracked"
```

**Use Cases**:

- File ownership queries
- Module boundary validation
- Code organization analysis
- Automated file categorization
- Build artifact collection

**Related Commands**:

- `show files` - Table format with ownership
- `show files staged` - Only staged files
- `show files changed` - Only modified files

---

### `get changed modules`

**Purpose**: Get list of modules affected by current git changes.

**Usage**:

```bash
# Modules with uncommitted changes
get changed modules
```

**Output Structure**:

```yaml
changed_modules:
  - src-cli
  - src-core

changed_files_count: 5
```

**Use Cases**:

- Incremental CI/CD (only test affected modules)
- Smart build systems
- Code review scope
- Impact analysis
- Targeted testing

**Detection Logic**:

1. Run `git status --porcelain`
2. Map changed files to modules
3. Return unique module list

**Related Commands**:

- `show files changed` - See which files changed
- `get execution order` - Build order for changed modules

---

## Environment Commands

### `get environments`

**Purpose**: Get all environment contract definitions.

**Usage**:

```bash
get environments
get environments --as-json
```

**Output Structure**:

```yaml
- moniker: local01
  name: "Local Environment 01 - Docker Container"
  level: "L2"
  type: "docker"
  env_tags:
    - "local"
    - "isolated"
  system_deps:
    - "@deps:docker"

- moniker: plte01
  name: "PLTE Environment 01 - Kubernetes"
  level: "L3"
  type: "plte"
  env_tags:
    - "k8s"
    - "staging"
    - "plte"
  system_deps:
    - "@deps:kubectl"
    - "@deps:helm"
```

**Use Cases**:

- Environment provisioning scripts
- CI/CD environment selection
- Test environment documentation
- Dependency installation automation
- Infrastructure as code

**Environment Levels**:

- **L0**: In-process, very fast unit tests
- **L1**: Fast unit tests with Go toolchain
- **L2**: Local/Docker emulated systems
- **L3**: PLTE (Production-Like Test Environment)
- **L4**: Production environments

**Related Commands**:

- `show environments` - Table with summaries

---

## Test Commands

### `get suite`

**Purpose**: Get complete test suite with all tests and metadata.

**Usage**:

```bash
# Specific suite
get suite acceptance
get suite commit
get suite production-verification

# Different formats
get suite acceptance --as-json
```

**Output Structure**:

```yaml
suite_moniker: acceptance
suite_name: "PLTE Acceptance Tests"
description: "Stage 5-6 - Installation, Operational, and Performance Verification"
total_discovered: 487
production_tests:
  - moniker: src-cli_install-test_test-install-command-create-config-file
    test_name: "TestInstallCommand_CreateConfigFile"
    type: gotest
    module: src-cli
    module_type: go-cli
    level:
      - "@L1"
    verification:
      - "@ov"
    system_deps:
      - "@deps:go"
    module_deps: []
    is_ignored: false
    is_manual: false
    is_gxp: false
    is_critical_aspect: false

  - moniker: src-cli_cli-invocation_version-flag-displays-version
    test_name: "Version flag displays version"
    type: godog
    module: src-cli
    module_type: go-cli
    level:
      - "@L2"
    verification:
      - "@ov"
    system_deps:
      - "@deps:go"
      - "@deps:docker"
    module_deps:
      - "@depm:src-cli"

framework_tests: [...]  # Tests excluded from suite
validation_errors: {}   # Any validation failures
```

**Predefined Suites**:

| Suite                     | Purpose                | Selection Criteria |
| ------------------------- | ---------------------- | ------------------ |
| `commit`                  | Pre-commit checks      | L0-L2 tests        |
| `acceptance`              | PLTE acceptance        | @iv, @ov, @pv      |
| `production-verification` | Production smoke tests | L4 + @piv          |

**Use Cases**:

- Test execution planning
- CI/CD test selection
- Test coverage analysis
- Quality metrics
- Test documentation generation

**Data Enrichment**:

1. Discovers all tests
2. Applies global inferences
3. Infers deps from module types
4. Infers deps from environments
5. Selects tests matching suite
6. Validates test tags
7. Generates monikers

**Related Commands**:

- `show suite` - Formatted table display
- `test suite` - Execute suite tests

---

### `get tests`

**Purpose**: Get **all** tests in the repository with full metadata.

**Usage**:

```bash
# All tests
get tests

# Different formats
get tests --as-json
get tests --as-toml
```

**Output Structure**:

```yaml
total_tests: 487
tests:
  - moniker: src-cli_install-test_test-install-command-create-config-file
    test_name: "TestInstallCommand_CreateConfigFile"
    type: gotest
    file_path: "C:\\projects\\eac\\src\\cli\\cmd\\install_test.go"
    module: src-cli
    module_type: go-cli
    level:
      - "@L1"
    verification:
      - "@ov"
    system_deps:
      - "@deps:go"
    module_deps: []
    is_ignored: false
    skip_reason: ""
    is_manual: false
    risk_controls: []
    is_gxp: false
    is_critical_aspect: false

  - moniker: src-cli_cli-invocation_version-flag-displays-version
    test_name: "Version flag displays version"
    type: godog
    file_path: "C:\\projects\\eac\\specs\\src-cli\\cli-invocation\\specification.feature"
    module: src-cli
    module_type: go-cli
    level:
      - "@L2"
    verification:
      - "@ov"
    system_deps:
      - "@deps:go"
    module_deps:
      - "@depm:src-cli"
```

**Test Monikers**:

- **GoTest**: `module_test-file_TestName`
- **Godog**: `module_feature-name_scenario-name`

**Use Cases**:

- Complete test inventory
- Test discovery for IDEs
- Test analytics and metrics
- Custom test selection
- Test documentation generation
- Quality dashboards

**Discovery Includes**:

- All `*_test.go` files in `src/`
- All `*.feature` files in `specs/`
- Full tag enrichment via inference

**Related Commands**:

- `show tests` - Table format with summaries
- `get suite` - Filtered subset for a suite

---

## Output Format Examples

### YAML (Default)

```yaml
- moniker: src-cli
  name: "CLI Application"
  type: go-cli
```

### JSON

```json
[
  {
    "moniker": "src-cli",
    "name": "CLI Application",
    "type": "go-cli"
  }
]
```

### TOML

```toml
[[modules]]
moniker = "src-cli"
name = "CLI Application"
type = "go-cli"
```

---

## Automation Patterns

### Pattern 1: CI/CD Test Selection

```bash
# Get changed modules
CHANGED=$(get changed modules --as-json | jq -r '.changed_modules[]')

# Get execution order
ORDER=$(get execution order $CHANGED --as-json | jq -r '.execution_order[]')

# Run tests for each module
for module in $ORDER; do
    eac test module $module
done
```

### Pattern 2: Environment Provisioning

```bash
# Get environment requirements
get environments --as-json | \
  jq -r '.[] | select(.level == "L2") | .system_deps[]' | \
  sort -u > required-deps.txt

# Install dependencies
cat required-deps.txt | while read dep; do
    install-dependency $dep
done
```

### Pattern 3: Test Reporting

```bash
# Get all tests with metadata
get tests --as-json > tests.json

# Generate metrics
jq '.tests | group_by(.module) |
    map({module: .[0].module, count: length})' tests.json
```

### Pattern 4: Dependency Visualization

```bash
# Export to GraphViz
get dependencies --as-json | \
  python scripts/deps-to-dot.py | \
  dot -Tpng > dependencies.png
```

---

## Common Workflows

### Workflow 1: Incremental Build

```bash
# 1. Find what changed
get changed modules --as-json > changed.json

# 2. Get build order
cat changed.json | jq -r '.changed_modules[]' | \
  xargs get execution order --as-json > order.json

# 3. Build in order
cat order.json | jq -r '.execution_order[]' | \
  while read module; do
    eac build module $module
  done
```

### Workflow 2: Test Selection for PR

```bash
# 1. Get all L0-L1 tests (fast)
get suite commit --as-json > commit-tests.json

# 2. Extract test files
cat commit-tests.json | \
  jq -r '.production_tests[].file_path' | \
  sort -u > test-files.txt

# 3. Run tests
go test $(cat test-files.txt)
```

### Workflow 3: Environment Setup

```bash
# 1. Get PLTE environment
get environments --as-json | \
  jq '.[] | select(.moniker == "plte01")' > plte01.json

# 2. Extract dependencies
cat plte01.json | jq -r '.system_deps[]'

# 3. Provision based on deps
# @deps:kubectl → install kubectl
# @deps:helm → install helm
```

---

## Error Handling

All `get` commands return:

- **Exit Code 0**: Success
- **Exit Code 1**: Error (with message to stderr)

Example error handling:

```bash
if ! get modules --as-json > modules.json 2> error.log; then
    echo "Failed to get modules"
    cat error.log
    exit 1
fi
```

---

## Performance Considerations

| Command            | Typical Time | Notes                      |
| ------------------ | ------------ | -------------------------- |
| `get modules`      | < 50ms       | Embedded contracts         |
| `get dependencies` | < 100ms      | In-memory graph            |
| `get files`        | ~500ms       | File system walk           |
| `get suite`        | ~2s          | Full discovery + inference |
| `get tests`        | ~2s          | Full discovery + inference |

**Caching**: Currently no caching (stateless). Consider caching for `get tests` in CI.

---

## See Also

- [Show Commands Reference](show-commands.md)
- [Core Packages Architecture](../architecture/core-packages.md)
- [Module Contracts](../contracts/modules.md)
- [Environment Contracts](../contracts/environments.md)
- [Test Suites](../specifications/test-suites.md)
