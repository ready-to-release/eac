# Command Naming Conventions
## Overview

EAC commands follow consistent naming conventions that make them predictable and discoverable. Understanding these conventions helps you guess command names and understand their purpose.

## Core Principles

### 1. Verb-First Structure

Most commands follow the pattern: `<verb> <noun> [<modifier>]`

**Examples**:

```bash
get modules          # verb: get, noun: modules
show dependencies    # verb: show, noun: dependencies
validate contracts   # verb: validate, noun: contracts
create commit-message  # verb: create, noun: commit-message
scan secrets         # verb: scan, noun: secrets
```

**Rationale**: Verb-first naming groups commands by action, making it easier to find commands when you know what you want to do.

### 2. Semantic Verbs

Commands use semantically meaningful verbs that clearly indicate their purpose:

| Verb        | Meaning                       | Output         | Examples                               |
| ----------- | ----------------------------- | -------------- | -------------------------------------- |
| `get`       | Retrieve structured data      | JSON           | `get modules`, `get dependencies`      |
| `show`      | Display formatted information | Human-readable | `show modules`, `show dependencies`    |
| `create`    | Generate new content          | Varies         | `create spec`, `create pr`             |
| `validate`  | Check correctness             | Pass/fail      | `validate contracts`, `validate specs` |
| `scan`      | Analyze for issues            | Report         | `scan secrets`, `scan vuln`            |
| `test`      | Execute tests                 | Test results   | `test`, `test suite`                   |
| `build`     | Compile/build                 | Artifacts      | `build`                                |
| `release`   | Manage releases               | Varies         | `release changelog`, `release this`    |
| `work`      | Manage workspaces             | Varies         | `work create`, `work commit`           |
| `pipeline`  | CI/CD orchestration           | Varies         | `pipeline run`, `pipeline status`      |
| `serve`     | Start servers                 | Server process | `serve docs`, `serve design`           |
| `templates` | Install templates             | Varies         | `templates install docs`, `templates install ai` |
| `update`    | Modify existing               | Varies         | `update design`                        |

### 3. Consistent Naming Within Categories

Commands within the same category follow similar patterns:

**get commands** (all return JSON):

```bash
get modules
get dependencies
get files
get config
get environments
```

**show commands** (all return formatted tables/text):

```bash
show modules
show dependencies
show files
show config
show environments
```

**validate commands** (all perform validation):

```bash
validate contracts
validate dependencies
validate specs
validate design
```

---

## Naming Patterns

### Simple Commands

**Pattern**: `<verb> <noun>`

**Examples**:

```bash
get modules
show files
build
test
help
init
```

**When used**: For straightforward operations on a single entity type.

---

### Compound Commands

**Pattern**: `<verb> <noun-phrase>`

Noun phrases are hyphenated:

```bash
create commit-message
create squash-message
get changed-modules
show valid-commands
validate module-files
```

**Rules**:

- Use hyphens (`-`) to connect words in noun phrases
- Never use underscores (`_`) or camelCase
- Keep noun phrases concise (2-3 words max)

**Examples**:

- ✓ `commit-message` (hyphenated)
- ✗ `commit_message` (underscored)
- ✗ `commitMessage` (camelCase)

---

### Hierarchical Commands

**Pattern**: `<category> <subcategory> [<action>]`

Some commands form hierarchies:

**pipeline commands**:

```bash
pipeline run
pipeline status
pipeline wait
pipeline ci                      # Subcategory
pipeline ci dispatch-and-wait    # Subcategory + action
pipeline ci summary-link         # Subcategory + action
```

**templates commands**:

```bash
templates install
templates install docs          # With type specifier
templates install ai            # With type specifier
templates install reports       # With type specifier
templates install specs         # With type specifier
```

**release commands**:

```bash
release changelog
release pending
release this
release check-ci
release get-version
release generate-module-calver
```

**Rules**:

- Hierarchies indicate related functionality
- Base command (e.g., `pipeline ci`) may have its own behavior
- Subcommands extend the base (e.g., `pipeline ci dispatch-and-wait`)

---

### Commands with Modifiers

**Pattern**: `<verb> <noun> <modifier>`

Some commands include additional context:

```bash
get changed-modules-ci          # "ci" modifier specifies CI context
templates install docs          # "docs" specifies template type
templates install reports       # "reports" specifies template type
show files-changed              # "changed" modifier filters files
show files-staged               # "staged" modifier filters files
```

**When used**: To provide variants of a base command for specific contexts or filters.

---

## get vs show Duality

Many information retrieval commands come in pairs: `get` and `show`.

### The Pattern

```bash
get <noun>    # JSON output for automation
show <noun>   # Human-readable output for interactive use
```

### Examples

| get command              | show command              | Purpose               |
| ------------------------ | ------------------------- | --------------------- |
| `get modules`            | `show modules`            | Module information    |
| `get dependencies`       | `show dependencies`       | Dependency graph      |
| `get files`              | `show files`              | File ownership        |
| `get config`             | `show config`             | Configuration         |
| `get environments`       | `show environments`       | Environment contracts |
| `get tests`              | `show tests`              | Test information      |
| `get suite <name>`       | `show suite <name>`       | Test suite details    |
| `get artifacts <module>` | `show artifacts <module>` | Build artifacts       |
| `get build-times`        | `show build-times`        | Build performance     |
| `get test-timings`       | `show test-timings`       | Test performance      |
| `get valid-commands`     | `show valid-commands`     | Command list          |

### When to Use Which

**Use `get` when**:

- Writing CI/CD scripts
- Piping to `jq` for processing
- Integrating with other tools
- Need JSON output

**Use `show` when**:

- Exploring interactively
- Reading output directly
- Need formatted tables
- Want colorized output

**Example**:

```bash
# Automation: get + jq
r2r eac get modules | jq -r '.modules[].moniker'

# Interactive: show
r2r eac show modules
```

---

## Noun Selection Rules

### Plural vs Singular

**General rule**: Use plural for collections, singular for specific items.

**Plurals** (operate on collections):

```bash
get modules          # All modules
show dependencies    # All dependencies
get files            # All files
validate contracts   # All contracts
```

**Singulars** (operate on one item or generate one artifact):

```bash
build <module>       # One module
test <module>        # One module
create spec          # One specification
create pr            # One pull request
```

**Exceptions**:

```bash
get config           # Configuration (uncountable)
show config          # Configuration (uncountable)
help                 # Help (uncountable)
init                 # Initialization (action)
```

### Domain-Specific Terms

Use terms consistent with the project's domain:

| Concept                | Term Used   | Not Used              |
| ---------------------- | ----------- | --------------------- |
| Module identifier      | moniker     | name, id              |
| Specification file     | spec        | feature, scenario     |
| Architecture diagram   | design      | architecture, diagram |
| Git worktree workspace | workspace   | worktree, branch      |
| PLTE environment       | environment | env, target           |
| Executable unit        | command     | cmd, tool             |

---

## Abbreviations and Acronyms

### Allowed Abbreviations

These abbreviations are standard and widely understood:

```bash
ci                  # Continuous Integration
pr                  # Pull Request
iac                 # Infrastructure as Code
sast                # Static Application Security Testing
sbom                # Software Bill of Materials
zap                 # OWASP ZAP (tool name)
cli                 # Command Line Interface
```

### Avoid Over-Abbreviating

Don't abbreviate unless the abbreviation is standard:

- ✓ `create commit-message` (clear)
- ✗ `create cm` (unclear abbreviation)
- ✓ `get dependencies` (clear)
- ✗ `get deps` (abbreviation used internally but not in command names)

---

## Case Sensitivity

### Command Names

All command names are **lowercase with hyphens**:

```bash
# Correct
create commit-message
get changed-modules
show valid-commands

# Incorrect
Create-Commit-Message  # Wrong case
create_commit_message  # Wrong separator
createCommitMessage    # Wrong format
```

### Arguments

Arguments (like module monikers) maintain their defined casing:

```bash
# Correct
r2r eac build eac-commands      # Module moniker as defined
r2r eac test src-auth           # Module moniker as defined

# Incorrect
r2r eac build EAC-COMMANDS      # Wrong case
r2r eac build eac_commands      # Wrong format
```

---

## Consistency Checklist

When adding new commands, ensure:

- [ ] Verb-first structure (unless it's a noun-only command like `help`, `init`)
- [ ] Verb matches output type (`get` → JSON, `show` → formatted)
- [ ] Hyphens used for compound nouns (not underscores or camelCase)
- [ ] Plural for collections, singular for single items
- [ ] Domain-specific terms used consistently
- [ ] Only standard abbreviations used
- [ ] All lowercase with hyphens
- [ ] Follows category conventions
- [ ] If retrieving data, has both `get` and `show` variants (if appropriate)

---

## Examples by Category

### create commands

```bash
create commit-message      # Generate commit message
create pr                  # Create pull request
create spec                # Generate specification
create design              # Generate architecture diagram
create squash-message      # Generate squash message
create risk-assess         # Generate risk assessment
create risk-profile        # Generate risk profile
```

### get commands

```bash
get modules                # Module information (JSON)
get dependencies           # Dependency graph (JSON)
get files                  # File mappings (JSON)
get changed-modules        # Changed modules (JSON)
get execution order <m>    # Build order (JSON)
```

### show commands

```bash
show modules               # Module information (table)
show dependencies          # Dependency graph (table)
show files                 # File mappings (table)
show files-changed         # Changed files (table)
show config                # Configuration (formatted)
```

### validate commands

```bash
validate contracts         # Validate repository contracts
validate dependencies      # Validate module dependencies
validate specs             # Validate specifications
validate go-tidy           # Validate Go dependencies
validate design            # Validate architecture syntax
```

### scan commands

```bash
scan secrets               # Detect secrets
scan vuln                  # Scan vulnerabilities
scan sast                  # Static analysis
scan iac                   # Infrastructure scan
scan compliance            # Compliance check
```

---

## Anti-Patterns

### Don't Do This

❌ **Mixed separators**:

```bash
get_modules              # Uses underscore
getModules               # Uses camelCase
```

❌ **Verb-last**:

```bash
modules get              # Verb should be first
contracts validate       # Verb should be first
```

❌ **Unclear abbreviations**:

```bash
get mods                 # Use full word: modules
show deps                # Use full word: dependencies
val contracts            # Use full word: validate
```

❌ **Inconsistent naming**:

```bash
get module-list          # Should be: get modules
show all-dependencies    # Should be: show dependencies
```

### Do This Instead

✓ **Consistent patterns**:

```bash
get modules              # Hyphenated, verb-first
show modules             # Consistent with get
validate contracts       # Clear verb
```

---

## See Also

- [Command Taxonomy](./command-taxonomy.md) - How commands are organized into categories
- [Common Flags](./common-flags.md) - Global options available to commands
- [Output Formats](./output-formats.md) - Understanding command output formats
- [Help Command](../help/help.md) - Using the help system
