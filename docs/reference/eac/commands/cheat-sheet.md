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
eac show modules

# Get modules as JSON
eac get modules

# Show module dependencies
eac show dependencies

# Get dependency graph as JSON
eac get dependencies
```

### View Module Information

```bash
# Show configuration
eac show config

# List all files with module ownership
eac show files

# Show only changed files
eac show files-changed

# Show staged files
eac show files-staged
```

---

## Building and Testing

### Build Commands

```bash
# Build a single module
eac build src-auth

# Build multiple modules
eac build src-auth src-api

# Build multiple modules
eac build src-auth src-api
```

### Test Commands

```bash
# Test a module (all suites)
eac test src-auth

# Run specific test suite
eac test suite acceptance

# List available test suites
eac test list-suites

# Show test results
eac show test-summary src-auth acceptance
```

### Debug Test Failures

```bash
# Parse and list failures
eac test debug

# Show test timing analysis
eac show test-timings
```

---

## Git Workflow

### Workspace Management

```bash
# Create new worktree for feature
eac work create feature/auth

# List all workspaces
eac show workspaces

# Sync workspace with main
eac work pull

# Merge workspace to main
eac work merge

# Remove workspace
eac work remove feature/auth
```

### AI-Powered Commits

```bash
# Stage changes
git add .

# Generate commit message with AI
eac work commit

# Alternative: Generate message only (no commit)
eac create commit-message
```

### Pull Requests

```bash
# Create PR with AI-generated description
eac create pr

# Generate squash commit message
eac create squash-message main..feature/auth
```

---

## CI/CD

### Change Detection

```bash
# Get modules affected by changes
eac get changed-modules

# Get modules requiring rebuild in CI
eac get changed-modules-ci
```

### Pipeline Execution

```bash
# Run pipeline for module
eac pipeline run src-auth

# Check CI status
eac pipeline status

# Wait for CI to complete
eac pipeline wait

# Orchestrate CI build
eac pipeline ci
```

---

## Release Management

### Check Release Status

```bash
# Check for pending changes
eac release pending

# Check for tag-pending versions
eac release tag-pending

# Get current version from changelog
eac release get-version
```

### Generate Release Materials

```bash
# Generate changelog from commits
eac release changelog

# Validate changelog format
eac validate release

# Check CI status for release
eac release check-ci $(git rev-parse HEAD)
```

### Create Release

```bash
# Finalize module release (creates git tag)
eac release this

# Generate calver tag for module
eac release generate-module-calver src-auth

# Validate version format
eac validate release-version
```

---

## Quality and Validation

### Pre-Commit Validation

```bash
# Validate everything
eac validate

# Validate contracts
eac validate contracts

# Validate dependencies
eac validate dependencies

# Check Go module tidiness
eac validate go-tidy
```

### Specification Validation

```bash
# Validate Gherkin specs
eac validate specs

# Find unused step definitions
eac get specs unused-steps
```

### File and Structure Validation

```bash
# Validate module file ownership
eac validate module-files

# Validate module hierarchy
eac validate module-hierarchy

# Validate markdown syntax
eac validate markdown

# Validate architecture diagrams
eac validate design
```

---

## Documentation

### Architecture Documentation

```bash
# Create architecture diagram
eac create design src-auth

# Update existing diagram
eac update design src-auth

# View diagrams in browser
eac serve design src-auth

# Validate diagram syntax
eac validate design
```

### Specifications

```bash
# Generate Gherkin spec from description
eac create spec "User can login with email and password"

# Validate specifications
eac validate specs
```

### Documentation Site

```bash
# Start documentation server
eac serve docs

# Stop documentation server
eac serve docs --stop
```

### Templates

```bash
# Install documentation templates
clie templates install docs

# Install to custom location
clie templates install docs --destination ./custom-docs

# Install AI prompt templates
clie templates install ai

# Install report templates
clie templates install reports

# Install specification templates
clie templates install specs
```

---

## Security Scanning

### Complete Security Scan

```bash
# Run all security scans
eac scan
```

### Individual Scans

```bash
# Scan for vulnerabilities
eac scan --scanner vuln

# Scan for secrets
eac scan --scanner secrets

# Static analysis (SAST)
eac scan --scanner sast

# Infrastructure as Code scan
eac scan --scanner iac

# Generate SBOM
eac scan --scanner sbom

# Check compliance
eac scan --scanner compliance

# Dynamic analysis (DAST)
eac scan zap eac-api --target http://localhost:8080
```

---

## Common Patterns

### JSON Output for Automation

Most `get` commands output JSON for scripting:

```bash
# Get modules and process with jq
eac get modules | jq '.modules[].moniker'

# Filter by type
eac get modules | jq '.modules[] | select(.type == "go-library")'

# Count results
eac get modules | jq '.modules | length'
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
eac show help <command>

# List all valid commands
eac show valid-commands

# Get command metadata
eac get valid-commands
```

---

## Keyboard-Friendly Aliases

Consider creating shell aliases for frequently used commands:

```bash
# ~/.bashrc or ~/.zshrc
alias clie-modules='eac show modules'
alias clie-build='eac build'
alias clie-test='eac test'
alias clie-validate='eac validate'
alias clie-commit='eac work commit'
alias clie-changed='eac get changed-modules'
```

---

## Common Workflows

### Full Build and Test Cycle

```bash
# 1. Make changes
# 2. Validate
eac validate

# 3. Build affected modules
eac get changed-modules | xargs -L1 eac build

# 4. Run tests
eac get changed-modules | xargs -L1 eac test

# 5. Commit with AI
eac work commit
```

### Release Workflow

```bash
# 1. Check pending changes
eac release pending

# 2. Generate changelog
eac release changelog

# 3. Validate changelog
eac validate release

# 4. Check CI status
eac release check-ci $(git rev-parse HEAD)

# 5. Create release tag
eac release this
```

### CI Build Workflow

```bash
# 1. Get changed modules since last CI
eac get changed-modules-ci

# 2. Build in dependency order
for module in $(eac get changed-modules-ci | jq -r '.changed_modules[]'); do
  eac pipeline run $module
done

# 3. Wait for completion
eac pipeline wait
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
