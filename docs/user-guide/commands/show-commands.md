# Show Commands Reference

The `show` commands provide **human-readable output** in markdown table formats for interactive use and reporting.

## Architecture Principle

```
get commands  = CANONICAL DATA SOURCE (YAML/JSON/TOML)
show commands = HUMAN-READABLE DISPLAY (formatted tables)
```

All `show` commands follow a consistent pattern:
1. Call the corresponding `get` logic to retrieve structured data
2. Format the data into markdown tables for readability
3. Add summaries, statistics, and contextual information
4. Display directly to stdout for terminal viewing
5. No side effects (read-only operations)

---

## Command Index

| Command | Purpose |
|---------|---------|
| [show modules](#show-modules) | Show all module contracts in table format |
| [show moduletypes](#show-moduletypes) | Show module types grouped by count |
| [show dependencies](#show-dependencies) | Show dependency graph with statistics |
| [show files](#show-files) | Show repository files with ownership |
| [show files staged](#show-files-staged) | Show staged files with ownership |
| [show files changed](#show-files-changed) | Show modified files with ownership |
| [show environments](#show-environments) | Show environment contracts in table |
| [show suite](#show-suite) | Show test suite with full details |
| [show tests](#show-tests) | Show all tests in repository |

---

## Module Commands

### `show modules`

**Purpose**: Display all module contracts in a human-readable table format.

**Usage**:
```bash
eac show modules
```

**Output Format**:
```
| Moniker      | Type         | Root Path      |
|--------------|--------------|----------------|
| src-cli      | go-cli       | src/cli        |
| src-commands | go-commands  | src/commands   |
| src-core     | go-library   | src/core       |
```

**Use Cases**:
- Quick module overview
- Verify module structure
- Share module list in documentation
- Terminal-based exploration

**Related Commands**:
- `get modules` - Structured data for automation
- `show moduletypes` - Group by type

---

### `show moduletypes`

**Purpose**: Display module types grouped by count with summary statistics.

**Usage**:
```bash
eac show moduletypes
```

**Output Format**:
```
| Module Type  | Count |
|--------------|-------|
| go-cli       | 1     |
| go-commands  | 1     |
| go-library   | 1     |
|--------------|-------|
| Total Types  | 3     |
```

**Use Cases**:
- Understand module type distribution
- Identify dominant patterns
- Architecture review
- Refactoring planning

**Features**:
- Alphabetically sorted types
- Count per type
- Footer with total type count

**Related Commands**:
- `show modules` - See individual modules
- `get modules` - Filter by type programmatically

---

## Dependency Commands

### `show dependencies`

**Purpose**: Display module dependency graph with statistics and execution order.

**Usage**:
```bash
eac show dependencies
```

**Output Format**:
```markdown
# Module Dependency Graph

## Statistics

| Metric                         | Value |
|--------------------------------|-------|
| Total Modules                  | 3     |
| Total Dependencies             | 2     |
| Root Modules (no dependencies) | 1     |
| Leaf Modules (no dependents)   | 2     |
| Max Dependencies               | 1     |
| Max Dependents                 | 2     |

## Module Dependencies

| Module       | Depends On | Used By              |
|--------------|------------|----------------------|
| src-cli      | src-core   | -                    |
| src-commands | src-core   | -                    |
| src-core     | -          | src-cli, src-commands|

## Execution Order

Total layers: 2

| Layer   | Modules (can run in parallel) | Count |
|---------|-------------------------------|-------|
| Layer 0 | src-core                      | 1     |
| Layer 1 | src-cli, src-commands         | 2     |
```

**Sections**:
1. **Statistics**: Overview metrics
2. **Module Dependencies**: Bidirectional relationships
3. **Execution Order**: Topological sort with parallel layers

**Use Cases**:
- Understand module relationships
- Plan build/test execution
- Identify refactoring opportunities
- Detect bottlenecks in dependency chains
- Document architecture

**Related Commands**:
- `get dependencies` - Structured dependency data
- `get execution order` - Just the execution order

---

## File Commands

### `show files`

**Purpose**: Show all repository files with their module ownership.

**Usage**:
```bash
eac show files
```

**Output Format**:
```
| File                           | Modules      |
|--------------------------------|--------------|
| src/cli/main.go                | src-cli      |
| src/core/testing/suites.go     | src-core     |
| specs/src-cli/cli-invocation/  | src-cli      |
| specification.feature          |              |
| docs/README.md                 | NONE         |
```

**Features**:
- Shows all tracked files
- Sorted by module ownership
- Files with no owner show "NONE"
- Multiple modules shown as comma-separated list

**Use Cases**:
- Understand file organization
- Verify module boundaries
- Find unowned files
- Module migration planning

**Related Commands**:
- `get files` - Structured file data
- `show files staged` - Only staged files
- `show files changed` - Only modified files

---

### `show files staged`

**Purpose**: Show git-staged files with their module ownership.

**Usage**:
```bash
eac show files staged
```

**Output Format**:
```
| File                        | Modules      |
|-----------------------------|--------------|
| src/cli/cmd/install.go      | src-cli      |
| src/core/testing/suites.go  | src-core     |
```

**Use Cases**:
- Review changes before commit
- Understand which modules are affected
- Verify commit scope
- Pre-commit validation

**Git Integration**:
- Uses `git diff --cached --name-only`
- Only shows staged files
- Empty output if nothing staged

**Related Commands**:
- `get changed modules` - Get affected modules
- `show files changed` - Unstaged changes
- `commit-ai` - Generate commit message

---

### `show files changed`

**Purpose**: Show modified (unstaged) files with their module ownership.

**Usage**:
```bash
eac show files changed
```

**Output Format**:
```
| File                        | Modules      |
|-----------------------------|--------------|
| src/cli/main.go             | src-cli      |
| src/core/testing/suites.go  | src-core     |
```

**Use Cases**:
- See uncommitted work
- Identify affected modules
- Incremental development tracking
- Code review preparation

**Git Integration**:
- Uses `git diff --name-only HEAD`
- Shows modified but unstaged files
- Empty output if no changes

**Related Commands**:
- `show files staged` - Staged changes
- `get changed modules` - Affected modules list

---

## Environment Commands

### `show environments`

**Purpose**: Display all environment contracts in a human-readable table with summaries.

**Usage**:
```bash
eac show environments
```

**Output Format**:
```markdown
# Environment Contracts

**Version**: 0.1.0
**Description**: Environment execution contracts for testing
**Total Environments**: 10

| Moniker  | Name                                | Level | Type       | System Dependencies        | Env Tags         |
|----------|-------------------------------------|-------|------------|----------------------------|------------------|
| l00-01   | L0 Environment 01 - In-process Unit | L0    | unit       |                            | in-process, very-fast |
| l00-02   | L0 Environment 02 - In-process Unit | L0    | unit       |                            | in-process, very-fast |
| l01-01   | L1 Environment 01 - Go Unit Tests   | L1    | unit       | @deps:go                   | go, fast         |
| l01-02   | L1 Environment 02 - Go Unit Tests   | L1    | unit       | @deps:go                   | go, fast         |
| local01  | Local Environment 01 - Docker       | L2    | docker     | @deps:docker               | local, isolated  |
| local02  | Local Environment 02 - Docker       | L2    | docker     | @deps:docker               | local, isolated  |
| plte01   | PLTE Environment 01 - Kubernetes    | L3    | plte       | @deps:kubectl, @deps:helm  | k8s, staging     |
| plte02   | PLTE Environment 02 - Kubernetes    | L3    | plte       | @deps:kubectl, @deps:helm  | k8s, staging     |
| production         | Production Environment        | L4    | production | @deps:kubectl, @deps:helm  | live, prod       |
| production-inactive| Production Inactive           | L4    | production | @deps:kubectl, @deps:helm  | live, inactive   |

## Summary by Level

- **L0 (Very Fast Unit)**: 2 environments
- **L1 (Fast Unit)**: 2 environments
- **L2 (Local/Docker)**: 2 environments
- **L3 (PLTE)**: 2 environments
- **L4 (Production)**: 2 environments

## Summary by Type

- **unit**: 4 environments
- **docker**: 2 environments
- **plte**: 2 environments
- **production**: 2 environments
```

**Sections**:
1. **Contract Metadata**: Version and description
2. **Environment Table**: All environments with details
3. **Summary by Level**: Count per L0-L4
4. **Summary by Type**: Count per environment type

**Use Cases**:
- Understand test execution environments
- Plan test infrastructure
- Select appropriate environment for tests
- Document environment contracts
- Verify environment coverage

**Environment Test Tags**:
- Test tag format: `@env:<moniker>`
- Example: `@env:local01`, `@env:plte01`
- System dependencies are automatically inferred

**Related Commands**:
- `get environments` - Structured environment data

---

## Test Commands

### `show suite`

**Purpose**: Display detailed information about a test suite in markdown table format.

**Usage**:
```bash
# Show specific suite
eac show suite commit
eac show suite acceptance
eac show suite production-verification
```

**Available Suites**:
- `commit` - Pre-commit checks (L0-L2 tests)
- `acceptance` - PLTE acceptance tests (@iv, @ov, @pv)
- `production-verification` - Production verification (@L4 + @piv)

**Output Format**:
```markdown
# Test Suite: Commit Suite

**Moniker**: `commit`
**Description**: Pre-commit checks - fast tests only (L0-L2)
**Production Tests**: 245
**Framework Tests**: 15 (excluded from display)
**Total Discovered**: 487

## Selection Criteria

**Selector 1**:
  - **AnyOf**: @L0, @L1, @L2
  - **Exclude**: @ignore, @manual

## Production Tests

| # | Moniker | Test Name | Type | Module | Level | Verification | System Deps |
|---|---------|-----------|------|--------|-------|--------------|-------------|
| 1 | src-cli_install-test_test-install-command-create-config-file | TestInstallCommand_CreateConfigFile | gotest | src-cli | @L1 | @ov | @deps:go |
| 2 | src-cli_cli-invocation_version-flag-displays-version | Version flag displays version | godog | src-cli | @L2 | @ov | @deps:go, @deps:docker |
...

## Statistics

**By Type**:
  - gotest: 180
  - godog: 65

**By Module**:
  - src-cli: 85
  - src-commands: 60
  - src-core: 100

**Dependencies**:
  - System: @deps:go, @deps:docker
  - Module: @depm:src-cli, @depm:src-core
```

**Sections**:
1. **Suite Metadata**: Name, description, counts
2. **Selection Criteria**: Tag selectors used
3. **Production Tests**: Table of all tests
4. **Statistics**: Breakdowns by type, module, dependencies

**Features**:
- **Validation Warnings**: Shows tests with validation errors
- **Framework Tests Excluded**: Internal tests not shown
- **Monikers**: Unique test identifiers
- **Tag Categories**: Separated into Level, Verification, Deps

**Use Cases**:
- Review suite composition
- Understand test selection
- Verify suite coverage
- Plan test execution
- Document test suites

**Related Commands**:
- `get suite` - Structured suite data
- `test suite` - Execute suite tests
- `show tests` - All tests in repository

---

### `show tests`

**Purpose**: Show all tests in the repository in a human-readable table format.

**Usage**:
```bash
eac show tests
```

**Output Format**:
```markdown
# All Tests

**Total Tests**: 487

| # | Moniker | Type | Module | Level | Verification | System Deps |
|---|---------|------|--------|-------|--------------|-------------|
| 1 | src-cli_install-test_test-install-command-create-config-file | gotest | src-cli | @L1 | @ov | @deps:go |
| 2 | src-cli_cli-invocation_version-flag-displays-version | godog | src-cli | @L2 | @ov | @deps:go, @deps:docker |
...

## Summary

### By Type

- **gotest**: 320 tests
- **godog**: 167 tests

### By Level

- **@L0**: 50 tests
- **@L1**: 180 tests
- **@L2**: 150 tests
- **@L3**: 80 tests
- **@L4**: 27 tests

### By Module

- **src-cli**: 85 tests
- **src-commands**: 142 tests
- **src-core**: 260 tests
```

**Sections**:
1. **All Tests Table**: Complete test inventory
2. **By Type Summary**: gotest vs godog counts
3. **By Level Summary**: L0-L4 distribution
4. **By Module Summary**: Tests per module

**Discovery**:
- All `*_test.go` files in `src/`
- All `*.feature` files in `specs/`
- Full tag enrichment via inference

**Test Monikers**:
- **GoTest**: `module_test-file_TestName` (kebab-case)
- **Godog**: `module_feature-name_scenario-name` (kebab-case)
- Unique and stable identifiers

**Use Cases**:
- Complete test inventory
- Test discovery for IDEs
- Quality metrics and analytics
- Custom test selection
- Test documentation generation

**Related Commands**:
- `get tests` - Structured test data
- `show suite` - Filtered test subsets
- `get suite` - Suite with metadata

---

## Common Patterns

### Pattern 1: Module Exploration
```bash
# Start with high-level overview
eac show modules

# Dive into types
eac show moduletypes

# Understand relationships
eac show dependencies
```

### Pattern 2: File Organization Review
```bash
# See all files
eac show files

# Check current work
eac show files changed

# Review what's staged
eac show files staged
```

### Pattern 3: Test Suite Review
```bash
# See all tests
eac show tests

# Review specific suite
eac show suite commit

# Check PLTE tests
eac show suite acceptance

# Verify production tests
eac show suite production-verification
```

### Pattern 4: Environment Planning
```bash
# Review all environments
eac show environments

# Pick appropriate level for new tests
# Use @env:<moniker> in test tags
```

---

## Output Characteristics

### Table Format
- Markdown tables for terminal rendering
- Headers clearly labeled
- Consistent column ordering
- Sortable in documentation

### Summaries
- Statistics sections where relevant
- Grouped data (by type, level, module)
- Dependency listings
- Totals and counts

### Validation Warnings
- Displayed before main output
- Clear error descriptions
- Per-test breakdown
- Non-fatal (shows data anyway)

---

## Integration with Get Commands

Every `show` command has a corresponding `get` command:

| Show Command | Get Command | Use Show When | Use Get When |
|--------------|-------------|---------------|--------------|
| show modules | get modules | Human review | Automation scripts |
| show moduletypes | get modules + filter | Type overview | Programmatic filtering |
| show dependencies | get dependencies | Visual review | Dependency analysis |
| show files | get files | File browsing | Build scripts |
| show files staged | get files + filter | Pre-commit review | CI validation |
| show files changed | get files + filter | Current work review | Incremental builds |
| show environments | get environments | Environment planning | Provisioning scripts |
| show suite | get suite | Suite review | Test execution |
| show tests | get tests | Test inventory | IDE integration |

**Principle**: Use `show` for interactive terminal use, `get` for automation.

---

## Performance

| Command | Typical Time | Notes |
|---------|-------------|-------|
| `show modules` | < 50ms | Embedded contracts |
| `show moduletypes` | < 50ms | In-memory grouping |
| `show dependencies` | < 150ms | Graph + execution order |
| `show files` | ~500ms | File system walk |
| `show files staged` | ~100ms | Git command + filter |
| `show files changed` | ~100ms | Git command + filter |
| `show environments` | < 50ms | Embedded contracts |
| `show suite` | ~2s | Full discovery + inference |
| `show tests` | ~2s | Full discovery + inference |

**Note**: File and test commands perform full discovery each time (stateless).

---

## Best Practices

### 1. Use for Interactive Exploration
```bash
# Quick checks in terminal
eac show modules
eac show files changed
```

### 2. Redirect to Markdown Files
```bash
# Save for documentation
eac show dependencies > docs/architecture/dependencies.md
eac show suite commit > docs/testing/commit-suite.md
```

### 3. Combine with Grep
```bash
# Find specific modules
eac show modules | grep "src-cli"

# Find tests for a module
eac show tests | grep "src-core"
```

### 4. Use Get for Scripts
```bash
# DON'T parse show output in scripts
# DO use get commands instead
modules=$(eac get modules --as-json | jq -r '.[].moniker')
```

---

## Error Handling

All `show` commands:
- **Exit Code 0**: Success
- **Exit Code 1**: Error
- **Stderr**: Error messages
- **Stdout**: Table output (if successful)

Example:
```bash
if ! eac show modules; then
    echo "Failed to show modules"
    exit 1
fi
```

---

## Terminal Rendering

### Wide Tables
If tables exceed terminal width, they may wrap. Solutions:
```bash
# Wider terminal
eac show dependencies

# Redirect to file
eac show tests > tests.md

# Use pager
eac show tests | less -S
```

### Markdown Preview
Many terminals support markdown rendering:
- GitHub CLI (`gh`)
- Glow (`glow`)
- Bat (`bat`)

```bash
# Render with glow
eac show dependencies | glow -

# Render with bat
eac show tests | bat --language=markdown
```

---

## See Also

- [Get Commands Reference](get-commands.md)
- [Core Packages Architecture](../architecture/core-packages.md)
- [Module Contracts](../contracts/modules.md)
- [Environment Contracts](../contracts/environments.md)
- [Test Suites](../specifications/test-suites.md)
