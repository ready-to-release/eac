# Command Cheat Sheet

Quick reference for the most commonly used EAC commands.

Commands are organized by workflow and use case for fast lookup.

## Quick Navigation

- [Module Development](#module-development)
- [Building and Testing](#building-and-testing)
- [Git Workflow](#git-workflow)
- [CI/CD](#cicd)
- [Release Management](#release-management)
- [Quality and Validation](#quality-and-validation)
- [Documentation](#documentation)
- [Security Scanning](#security-scanning)

---

## Module Development

### List and Explore Modules

```bash
# List all modules
r2r eac show modules

# Get modules as JSON
r2r eac get modules

# Show module dependencies
r2r eac show dependencies

# Get dependency graph as JSON
r2r eac get dependencies
```

### View Module Information

```bash
# Show configuration
r2r eac show config

# List all files with module ownership
r2r eac show files

# Show only changed files
r2r eac show files-changed

# Show staged files
r2r eac show files-staged
```

---

## Building and Testing

### Build Commands

```bash
# Build a single module
r2r eac build src-auth

# Build multiple modules
r2r eac build src-auth src-api

# Build with dependencies in order
r2r eac get execution order src-auth | xargs -L1 r2r eac build
```

### Test Commands

```bash
# Test a module (all suites)
r2r eac test src-auth

# Run specific test suite
r2r eac test suite acceptance

# List available test suites
r2r eac test list-suites

# Show test results
r2r eac show test-summary src-auth acceptance
```

### Debug Test Failures

```bash
# Parse and list failures
r2r eac test debug

# Show test timing analysis
r2r eac show test-timings
```

---

## Git Workflow

### Workspace Management

```bash
# Create new worktree for feature
r2r eac work create feature/auth

# List all workspaces
r2r eac show workspaces

# Sync workspace with main
r2r eac work pull

# Merge workspace to main
r2r eac work merge

# Remove workspace
r2r eac work remove feature/auth
```

### AI-Powered Commits

```bash
# Stage changes
git add .

# Generate commit message with AI
r2r eac work commit

# Alternative: Generate message only (no commit)
r2r eac create commit-message
```

### Pull Requests

```bash
# Create PR with AI-generated description
r2r eac create pr

# Generate squash commit message
r2r eac create squash-message main..feature/auth
```

---

## CI/CD

### Change Detection

```bash
# Get modules affected by changes
r2r eac get changed-modules

# Get modules requiring rebuild in CI
r2r eac get changed-modules-ci
```

### Pipeline Execution

```bash
# Run pipeline for module
r2r eac pipeline run src-auth

# Check CI status
r2r eac pipeline status

# Wait for CI to complete
r2r eac pipeline wait

# Orchestrate CI build
r2r eac pipeline ci
```

### Build Order

```bash
# Get execution order for dependencies
r2r eac get execution order src-api

# Build modules in order
r2r eac get execution order src-api | xargs -L1 r2r eac build
```

---

## Release Management

### Check Release Status

```bash
# Check for pending changes
r2r eac release pending

# Check for tag-pending versions
r2r eac release tag-pending

# Get current version from changelog
r2r eac release get-version
```

### Generate Release Materials

```bash
# Generate changelog from commits
r2r eac release changelog

# Validate changelog format
r2r eac validate release

# Check CI status for release
r2r eac release check-ci $(git rev-parse HEAD)
```

### Create Release

```bash
# Finalize module release (creates git tag)
r2r eac release this

# Generate calver tag for module
r2r eac release generate-module-calver src-auth

# Validate version format
r2r eac validate release-version
```

---

## Quality and Validation

### Pre-Commit Validation

```bash
# Validate everything
r2r eac validate

# Validate contracts
r2r eac validate contracts

# Validate dependencies
r2r eac validate dependencies

# Check Go module tidiness
r2r eac validate go-tidy
```

### Specification Validation

```bash
# Validate Gherkin specs
r2r eac validate specs

# Find unused step definitions
r2r eac get specs unused-steps
```

### File and Structure Validation

```bash
# Validate module file ownership
r2r eac validate module-files

# Validate module hierarchy
r2r eac validate module-hierarchy

# Validate markdown syntax
r2r eac validate markdown

# Validate architecture diagrams
r2r eac validate design
```

---

## Documentation

### Architecture Documentation

```bash
# Create architecture diagram
r2r eac create design src-auth

# Update existing diagram
r2r eac update design src-auth

# View diagrams in browser
r2r eac serve design src-auth

# Validate diagram syntax
r2r eac validate design
```

### Specifications

```bash
# Generate Gherkin spec from description
r2r eac create spec "User can login with email and password"

# Validate specifications
r2r eac validate specs
```

### Documentation Site

```bash
# Start documentation server
r2r eac serve docs

# Stop documentation server
r2r eac serve docs --stop
```

### Templates

```bash
# Install documentation templates
r2r templates install docs

# Install to custom location
r2r templates install docs --destination ./custom-docs

# Install AI prompt templates
r2r templates install ai

# Install report templates
r2r templates install reports

# Install specification templates
r2r templates install specs
```

---

## Security Scanning

### Complete Security Scan

```bash
# Run all security scans
r2r eac scan
```

### Individual Scans

```bash
# Scan for vulnerabilities
r2r eac scan --scanner vuln

# Scan for secrets
r2r eac scan --scanner secrets

# Static analysis (SAST)
r2r eac scan --scanner sast

# Infrastructure as Code scan
r2r eac scan --scanner iac

# Generate SBOM
r2r eac scan --scanner sbom

# Check compliance
r2r eac scan --scanner compliance

# Dynamic analysis (DAST)
r2r eac scan zap eac-api --target http://localhost:8080
```

---

## Common Patterns

### JSON Output for Automation

Most `get` commands output JSON for scripting:

```bash
# Get modules and process with jq
r2r eac get modules | jq '.modules[].moniker'

# Filter by type
r2r eac get modules | jq '.modules[] | select(.type == "go-library")'

# Count results
r2r eac get modules | jq '.modules | length'
```

### get vs show Duality

Information commands come in pairs:

| JSON (automation)  | Human-readable (interactive) |
| ------------------ | ---------------------------- |
| `get modules`      | `show modules`               |
| `get dependencies` | `show dependencies`          |
| `get files`        | `show files`                 |
| `get tests`        | `show tests`                 |
| `get config`       | `show config`                |

### Command Help

```bash
# Show help for any command
r2r eac show help <command>

# List all valid commands
r2r eac show valid-commands

# Get command metadata
r2r eac get valid-commands
```

---

## Keyboard-Friendly Aliases

Consider creating shell aliases for frequently used commands:

```bash
# ~/.bashrc or ~/.zshrc
alias r2r-modules='r2r eac show modules'
alias r2r-build='r2r eac build'
alias r2r-test='r2r eac test'
alias r2r-validate='r2r eac validate'
alias r2r-commit='r2r eac work commit'
alias r2r-changed='r2r eac get changed-modules'
```

---

## Common Workflows

### Full Build and Test Cycle

```bash
# 1. Make changes
# 2. Validate
r2r eac validate

# 3. Build affected modules
r2r eac get changed-modules | xargs -L1 r2r eac build

# 4. Run tests
r2r eac get changed-modules | xargs -L1 r2r eac test

# 5. Commit with AI
r2r eac work commit
```

### Release Workflow

```bash
# 1. Check pending changes
r2r eac release pending

# 2. Generate changelog
r2r eac release changelog

# 3. Validate changelog
r2r eac validate release

# 4. Check CI status
r2r eac release check-ci $(git rev-parse HEAD)

# 5. Create release tag
r2r eac release this
```

### CI Build Workflow

```bash
# 1. Get changed modules since last CI
r2r eac get changed-modules-ci

# 2. Build in dependency order
for module in $(r2r eac get changed-modules-ci | jq -r '.changed_modules[]'); do
  r2r eac pipeline run $module
done

# 3. Wait for completion
r2r eac pipeline wait
```

---

## Quick Reference Tables

### Most Common Commands

| Command               | Purpose              |
| --------------------- | -------------------- |
| `show modules`        | List all modules     |
| `build <module>`      | Build a module       |
| `test <module>`       | Test a module        |
| `validate`            | Validate everything  |
| `work commit`         | Commit with AI       |
| `get changed-modules` | Find changed modules |
| `show help <cmd>`     | Get command help     |

### Output Format Commands

| JSON Output        | Formatted Output    |
| ------------------ | ------------------- |
| `get modules`      | `show modules`      |
| `get dependencies` | `show dependencies` |
| `get files`        | `show files`        |
| `get config`       | `show config`       |
| `get tests`        | `show tests`        |

### Validation Commands

| Command                 | What It Checks         |
| ----------------------- | ---------------------- |
| `validate`              | Everything             |
| `validate contracts`    | Contract schemas       |
| `validate dependencies` | Module dependencies    |
| `validate specs`        | Gherkin specifications |
| `validate markdown`     | Markdown syntax        |
| `validate design`       | Architecture diagrams  |

---

## See Also

- [Full Command Reference](./index.md) - Complete command documentation
- [Command Taxonomy](./overview/command-taxonomy.md) - How commands are organized
- [Common Flags](./overview/common-flags.md) - Global options
- [Output Formats](./overview/output-formats.md) - JSON vs formatted output
