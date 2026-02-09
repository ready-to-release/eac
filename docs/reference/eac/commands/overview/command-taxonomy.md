# Command Taxonomy

## Overview

EAC commands are organized into a hierarchical taxonomy based on their primary function. Understanding this organization helps you quickly find the right command for your task.

## Command Categories

EAC provides many commands organized into top-level categories:

### 1. create

**Purpose**: AI-powered content generation

Create commands use AI to generate various artifacts including commit messages, specifications, architecture diagrams, and release documentation.

**Commands**:

- `create commit-message` - Generate semantic commit messages from staged changes
- `create design` - Generate architecture diagrams (workspace.dsl) with AI
- `create pr` - Create pull requests with AI-generated descriptions
- `create risk-assess` - Update OSCAL assessment-results with test and security evidence
- `create risk-profile` - Create OSCAL profiles from risk assessments using AI
- `create spec` - Generate Gherkin specifications from natural language
- `create squash-message` - Generate squash merge commit messages

**When to use**: During development and release preparation when you need AI assistance to generate structured documentation or content.

**See also**: [Create Commands Category](../categories/create.md)

---

### 2. get

**Purpose**: Structured JSON output for automation

Get commands provide machine-readable JSON output designed for CI/CD pipelines, build scripts, and automation tools.

**Key characteristics**:

- All output is valid JSON
- Designed to be piped through `jq`
- Non-zero exit codes on errors
- Cacheable and deterministic

**Command groups**:

**Module and Repository Information**:

- `get modules` - Module contracts
- `get dependencies` - Dependency graph
- `get files` - File-to-module mappings (expensive, 19k tokens)
- `get config` - EAC configuration
- `get environments` - Environment contracts

**Build and Test Information**:

- `get artifacts` - Resolved artifacts
- `get build-deps` - Build dependencies
- `get build-times` - Build performance data
- `get execution order` - Dependency-based build order
- `get suite` - Test suite information
- `get tests` - All tests in repository
- `get test-timings` - Test performance data

**Change Detection**:

- `get changed-modules` - Modules affected by local changes
- `get changed-modules-ci` - Modules requiring rebuild since last CI run

**Command Metadata**:

- `get commands` - Command metadata for shell integration
- `get valid-commands` - All valid commands in structured format

**Specifications**:

- `get specs unused-steps` - Detect unused godog step definitions

**When to use**: In CI/CD pipelines, build scripts, or when you need programmatic access to repository data.

**See also**: [Get Commands Category](../categories/get.md)

---

### 3. show

**Purpose**: Human-readable output for interactive use

Show commands provide formatted, human-readable output designed for terminal display and interactive exploration.

**Key characteristics**:

- Formatted tables and lists
- Colorized output
- Designed for human consumption
- Interactive use cases

**Command groups**:

**Module & Repository Information**:

- `show modules` - Display module contracts in table format
- `show component-types` - Module types grouped by count
- `show dependencies` - Dependency graph in table format
- `show files` - Repository files with module ownership
- `show files-changed` - Modified files with module ownership
- `show files-staged` - Staged files with module ownership
- `show config` - Configuration with defaults applied

**Build & Test Results**:

- `show build-summary` - Pretty build summary for GitHub Actions
- `show build-times` - Build timing analysis in readable format
- `show test-summary` - Pretty test summary for GitHub Actions
- `show test-timings` - Test timing analysis in readable format
- `show tests` - All tests in human-readable table
- `show suite` - Detailed test suite information

**Environment & Workspace**:

- `show environments` - Environment contracts in table format
- `show workspaces` - List all workspaces and their status
- `show artifacts` - Display artifacts with status
- `show books` - Display book configurations

**Help System**:

- `show help` - Display help information
- `show valid-commands` - All valid commands in table format

**When to use**: During interactive development and troubleshooting when you need to quickly understand repository state.

**See also**: [Show Commands Category](../categories/show.md)

---

### 4. validate

**Purpose**: Contract and quality validation

Validate commands ensure repository consistency, enforce contracts, and verify quality standards.

**Command groups**:

**Module Contracts**:

- `validate contracts` - Validate repository contracts against JSON schemas
- `validate dependencies` - Validate module dependencies match contracts
- `validate module-files` - Validate file ownership
- `validate module-hierarchy` - Validate dependency graph structure
- `validate artifacts` - Validate build artifacts exist

**Code Quality**:

- `validate go-tidy` - Ensure Go dependencies are tidy
- `validate markdown` - Validate markdown file syntax
- `validate test-tags` - Validate test tags against contract

**Specifications**:

- `validate specs` - Validate Gherkin specifications against quality contracts

**Architecture & Design**:

- `validate design` - Check workspace.dsl syntax using Structurizr CLI
- `validate books` - Validate books.yml configuration

**Release Management**:

- `validate release` - Validate changelog format and structure
- `validate release-version` - Validate release version format

**Risk & Compliance**:

- `validate control-tags` - Validate @control tags against OSCAL catalog
- `validate risk-catalog` - Validate OSCAL catalogs against schema
- `validate risk-profile` - Validate OSCAL profile documents against schema

**When to use**: In pre-commit hooks, CI pipelines, and before releases to ensure quality standards are met.

**See also**: [Validate Commands Category](../categories/validate.md)

---

### 5. work

**Purpose**: Workspace management using git worktrees

Work commands manage parallel development workspaces, enabling multiple topic branches to be worked on simultaneously without switching contexts.

**Commands**:

- `work create` - Create new workspace for parallel development
- `work commit` - Commit changes with AI-generated messages
- `work pull` - Sync workspace with latest main via rebase
- `work merge` - Merge workspace changes to main (squash by default)
- `work remove` - Remove workspace and optionally delete branches
- `show workspaces` - List all workspaces and their status

**When to use**: When you need to work on multiple features simultaneously or want isolated development environments.

**See also**: [Work Commands Category](../categories/work.md), [Workspace Commands Guide](../work/index.md)

---

### 6. test

**Purpose**: Testing and test suite management

Test commands run tests, manage test suites, and analyze test results.

**Commands**:

- `test` - Test one or more modules by moniker
- `test suite` - Run specific test suite (parallel by default)
- `test list-suites` - List all available test suites
- `test debug` - Parse test results and list failures

**When to use**: During development (unit tests), in CI pipelines (all test types), and when debugging test failures.

**See also**: [Test Commands Category](../categories/test.md)

---

### 7. build

**Purpose**: Module building

**Commands**:

- `build` - Build one or more modules by moniker

**When to use**: To compile modules in dependency order, both locally and in CI/CD pipelines.

**See also**: [Build Command Reference](../build/build.md)

---

### 8. pipeline

**Purpose**: CI/CD orchestration and monitoring

Pipeline commands orchestrate CI/CD workflows, monitor build status, and generate diagnostics.

**Commands**:

- `pipeline run` - Execute module pipelines respecting dependencies
- `pipeline ci` - CI orchestration and diagnostics
- `pipeline ci dispatch-and-wait` - Trigger and wait for GitHub workflows
- `pipeline ci summary-link` - Generate diagnostic markdown for CI summaries
- `pipeline status` - Show CI status for the head of trunk
- `pipeline wait` - Wait for GitHub workflow runs to complete

**When to use**: In CI/CD workflows, when orchestrating complex builds, and when monitoring deployment status.

**See also**: [Pipeline Commands Category](../categories/pipeline.md)

---

### 9. release

**Purpose**: Release management and versioning

Release commands manage the release process from changelog generation through tagging and validation.

**Commands**:

- `release changelog` - Generate or update changelog from commits
- `release pending` - Check if module has pending changes for release
- `release this` - Finalize changelog and prepare module for release
- `release get-version` - Extract latest version from changelog
- `release generate-module-calver` - Generate CalVer tag for a module
- `release clie` - Create SemVer tag for releasing clie
- `release check-ci` - Check CI status before releasing
- `release tag-pending` - Check for changelog versions without git tags

**When to use**: During release preparation and when creating versioned releases.

**See also**: [Release Commands Category](../categories/release.md)

---

### 10. scan

**Purpose**: Security scanning and evidence collection

Scan commands perform various security scans and generate compliance evidence.

**Commands**:

- `scan --scanner vuln` - Scan for vulnerabilities using Trivy
- `scan --scanner secrets` - Detect secrets and credentials using Trivy
- `scan --scanner sast` - Static Application Security Testing using Semgrep
- `scan --scanner iac` - Scan Infrastructure as Code for misconfigurations
- `scan --scanner compliance` - Check compliance with security standards
- `scan --scanner sbom` - Generate Software Bill of Materials
- `scan zap` - Dynamic Application Security Testing using OWASP ZAP (special subcommand)
- `scan` - Security scanning orchestrator with --scanner flag

**When to use**: In pre-commit hooks (fast scans), CI pipelines (comprehensive scans), and for compliance evidence generation.

---

### 11. serve

**Purpose**: Local development servers

Serve commands start local servers for documentation and architecture visualization.

**Commands**:

- `serve docs` - Start or stop MkDocs server
- `serve design` - View architecture diagrams in browser using Structurizr Lite

**When to use**: During documentation writing and when reviewing architecture diagrams.

**See also**: [Serve Commands Category](../categories/serve.md)

---

### 12. templates

**Purpose**: Template management for documentation and specifications

Templates commands install project templates for consistent documentation and specification creation.

**Commands**:

- `templates install` - Install templates overview
- `templates install docs` - Install documentation templates
- `templates install ai` - Install AI prompt templates
- `templates install reports` - Install report templates
- `templates install specs` - Install specification templates
- `templates` - Base templates command

**When to use**: When setting up new modules or generating standardized documentation.

**See also**: [Templates Commands Category](../categories/templates.md)

---

### 13. update (1 command)

**Purpose**: Update operations

**Commands**:

- `update design` - Update existing workspace.dsl for a module using AI

**When to use**: When refreshing architecture diagrams after code changes.

**See also**: [Update Commands Category](../categories/update.md)

---

### 14. Other Commands

Commands that don't fit into the verb-based categorization:

**Commands**:

- `help` - Display help information for commands
- `init` - Initialize AI provider configuration for the project
- `extension-meta` - Output extension metadata for clie CLI

**When to use**:

- `help`: Anytime you need command information
- `init`: First-time setup or when changing AI providers
- `extension-meta`: When integrating with clie CLI

**See also**: [Build Commands Category](../categories/build.md)

---

## Command Naming Patterns

### Verb-First Convention

Most commands follow a **verb-first naming pattern**: `<verb> <noun> [<modifier>]`

**Examples**:

- `get modules` - verb: get, noun: modules
- `show dependencies` - verb: show, noun: dependencies
- `validate contracts` - verb: validate, noun: contracts
- `create commit-message` - verb: create, noun: commit-message

### Multi-word Commands

Some commands have multi-word names connected by hyphens:

- `create commit-message`
- `get changed-modules-ci`
- `pipeline ci dispatch-and-wait`

### Hierarchical Commands

Some categories have hierarchical subcommands:

- `pipeline ci` (base)
  - `pipeline ci dispatch-and-wait`
  - `pipeline ci summary-link`
- `templates install` (base)
  - `templates install docs`
  - `templates install ai`
  - `templates install reports`
  - `templates install specs`

See [Naming Conventions](./naming-conventions.md) for detailed rules.

---

## Finding the Right Command

### By Task

**I want to...**

- **...see what modules exist**: `show modules`
- **...build a module**: `build <module>`
- **...run tests**: `test <module>` or `test suite <name>`
- **...commit my changes**: `work commit` or `create commit-message`
- **...create a pull request**: `create pr`
- **...check for errors**: `validate` commands
- **...scan for security issues**: `scan` commands
- **...create a release**: `release` commands
- **...get data for a script**: `get` commands
- **...view information**: `show` commands

### By Output Format

**I need...**

- **JSON output for automation**: Use `get` commands
- **Human-readable output**: Use `show` commands
- **Validation/checks**: Use `validate` commands
- **AI-generated content**: Use `create` commands

### By Development Stage

**I'm working on...**

- **Local development**: `work create`, `create spec`, `validate`, `test`, `build`
- **Committing changes**: `work commit`, `create commit-message`
- **Code review**: `create pr`, `show files-changed`
- **CI/CD**: `pipeline run`, `get changed-modules-ci`, `test suite`
- **Releases**: `release` commands
- **Security**: `scan` commands

---

## Command Organization Principles

### 1. Functional Grouping

Commands are grouped by their primary function (create, get, show, validate, etc.) rather than by domain (build, test, release).

This makes it easier to find commands when you know what you want to do, even if you're not sure which domain it belongs to.

### 2. Output Format Distinction

`get` vs `show` commands provide the same information in different formats:

- `get modules` → JSON for automation
- `show modules` → Table for humans

### 3. Workflow Alignment

Command categories align with common development workflows:

- Authoring: `create`, `work create`
- Validation: `validate`, `scan`, `test`
- Integration: `build`, `pipeline`
- Release: `release`

### 4. Discoverability

The `help` command and `show help` provide interactive discovery:

```bash
# Get help for any command
eac help <command>

# List all commands
eac show valid-commands

# Get command metadata
eac get commands
```

---

## See Also

- [Naming Conventions](./naming-conventions.md) - Command naming rules and patterns
- [Common Flags](./common-flags.md) - Global options available to commands
- [Output Formats](./output-formats.md) - Understanding JSON vs human-readable output
- [Categories Index](../categories/index.md) - Browse all command categories
- [Main Index](../index.md) - Command reference home
