# Command Reference

This section documents **EAC commands** - the automation tools provided by the EAC CLI.

## EAC Commands

Complete technical reference for all EAC commands (invoked as `eac <command>`).

## Quick Access

- [Language Support](../language-support.md) - Which commands work with which languages
- [Command Taxonomy](../overview/command-taxonomy.md) - How commands are organized
- [Naming Conventions](../overview/naming-conventions.md) - Command naming patterns
- [Common Flags](../overview/common-flags.md) - Global options
- [Output Formats](../overview/output-formats.md) - JSON vs human-readable output

## Command Categories

| Category                            | Purpose                                              |
| ----------------------------------- | ---------------------------------------------------- |
| [build](./build/index.md)          | Module building and compilation                      |
| [create](./create/index.md)        | AI-powered generation (commits, specs, designs, PRs) |
| [get](./get/index.md)              | JSON output for automation and scripting             |
| [help](./help/index.md)            | Display help information                             |
| [init](./init/index.md)            | Project initialization                               |
| [pipeline](./pipeline/index.md)    | CI/CD orchestration                                  |
| [release](./release/index.md)      | Release management and versioning                    |
| [scan](./scan/index.md)            | Security scanning (SAST, secrets, vulnerabilities)   |
| [serve](./serve/index.md)          | Local development servers                            |
| [show](./show/index.md)            | Human-readable output for interactive use            |
| [templates](./templates/index.md)  | Template management                                  |
| [test](./test/index.md)            | Testing and test suite management                    |
| [update](./update/index.md)        | Update operations                                    |
| [validate](./validate/index.md)    | Contract and dependency validation                   |
| [work](./work/index.md)            | Workspace management (git worktrees)                 |

## Core Commands

Core commands are top-level commands that don't require a category prefix:

| Command | Description |
|---------|-------------|
| [build](./build/build.md) | Build one or more modules |
| [lint](./lint.md) | Run linters on modules |
| [scan](./scan/scan.md) | Run security scanners on modules |
| [test](./test/test.md) | Run tests for modules |
| [extension-meta](./extension-meta.md) | Output extension metadata for CLI integration |
| [init](./init/init.md) | Initialize a new EAC repository |
| [help](./help/help.md) | Display help information |

### Linting Code

```bash
# Lint all modules
eac lint

# Lint specific module
eac lint eac-commands

# Lint with auto-fix
eac lint --fix
```

The lint command automatically selects appropriate linters based on module component types (Go, Markdown, etc.).

## Finding Commands by Task

### Development Tasks

**I want to...**

- **...see what modules exist**: [show modules](./show/modules.md)
- **...check dependencies**: [show dependencies](./show/dependencies.md)
- **...create a workspace**: [work create](./work/create.md)
- **...write specifications**: [create spec](./create/spec.md)
- **...generate diagrams**: [create design](./create/design.md)

### Quality and Validation

**I want to...**

- **...scan for secrets**: [scan](./scan/scan.md) with `--scanner secrets`
- **...check for vulnerabilities**: [scan](./scan/scan.md) with `--scanner vuln`

### Building and Testing

**I want to...**

- **...build a module**: [build](./build/build.md)
- **...run tests**: [test](./test/test.md)
- **...run a test suite**: [test suite](./test/suite.md)
- **...see test results**: [show test-summary](./show/test-summary.md)

### Committing and PRs

**I want to...**

- **...commit changes**: [work commit](./work/commit.md) or [get commit-message](./get/commit-message.md)
- **...create a PR**: [create pr](./create/pr.md)
- **...see changed files**: [show files-changed](./show/files-changed.md)
- **...get changed modules**: [get changed-modules](./get/changed-modules.md)

### Release and Deployment

**I want to...**

- **...check for release changes**: [release pending](./release/pending.md)
- **...generate changelog**: [release changelog](./release/changelog.md)
- **...create a release**: [release this](./release/this.md)
- **...check CI status**: [release check-ci](./release/check-ci.md)
- **...run pipelines**: [pipeline run](./pipeline/run.md)

### Documentation

**I want to...**

- **...serve docs locally**: [serve docs](./serve/docs.md)
- **...view architecture diagrams**: [serve design](./serve/design.md)
- **...manage templates**: [templates commands](./templates/index.md)

## Category Patterns

### Information Retrieval (get/show)

Most information commands come in pairs:

| get (JSON)         | show (Formatted)    | Information      |
| ------------------ | ------------------- | ---------------- |
| `get modules`      | `show modules`      | Module contracts |
| `get dependencies` | `show dependencies` | Dependency graph |
| `get files`        | `show files`        | File ownership   |
| `get config`       | `show config`       | Configuration    |
| `get tests`        | `show tests`        | Test information |

**Rule**: Use `get` for automation, `show` for interactive and templated markdown use.

### Quality Gates (validate/scan/test)

Quality commands all validate different aspects:

- **validate**: Repository structure and contracts
- **scan**: Security issues and compliance
- **test**: Functional correctness

**Use in**: Pre-commit hooks, CI pipelines, release gates

### Workflow (work/pipeline/release)

Workflow commands manage different stages:

- **work**: Local development (git worktrees)
- **pipeline**: CI/CD execution
- **release**: Version management

## Common Workflows

### Module Development

- [show modules](./show/modules.md) - List all modules
- [build](./build/build.md) - Build modules
- [test](./test/test.md) - Test modules

### CI/CD

- [get changed-modules-ci](./get/changed-modules-ci.md) - Find changed modules
- [pipeline run](./pipeline/run.md) - Execute pipelines
- [show build-summary](./show/build-summary.md) - Display results

### Release Management

- [release pending](./release/pending.md) - Check for changes
- [release changelog](./release/changelog.md) - Update changelog
- [release check-ci](./release/check-ci.md) - Verify CI status
- [release this](./release/this.md) - Create release

### Workspace Development

- [work create](./work/create.md) - Create workspace
- [work commit](./work/commit.md) - Commit with AI messages
- [work pull](./work/pull.md) - Sync with main
- [work merge](./work/merge.md) - Merge to main

## Getting Started

### New to EAC Commands?

1. **Start with the basics**:
   - [Command Taxonomy](../overview/command-taxonomy.md) - Understand how commands are organized
   - [Naming Conventions](../overview/naming-conventions.md) - Learn command naming patterns
   - [Common Flags](../overview/common-flags.md) - Global options all commands accept

2. **Explore by category**:
   - Browse the [Command Categories](#command-categories) table above
   - Start with [show commands](./show/index.md) for exploration
   - Use [get commands](./get/index.md) for automation

3. **Try common commands**:

   ```bash
   # Discover modules
   eac show modules

   # Get help for any command
   eac help <command>

   # Build a module
   eac build <module>

   # Run tests
   eac test <module>
   ```

### Looking for a Specific Command?

- **By name**: Use the search function or browse categories
- **By purpose**: See [Command Taxonomy](../overview/command-taxonomy.md#finding-the-right-command)
- **By output format**: See [Output Formats](../overview/output-formats.md)

## Command Reference Pages

Each command has a dedicated reference page with:

- **Overview**: Command purpose and use cases
- **Syntax**: Full command syntax with options
- **Arguments and Flags**: Detailed parameter documentation
- **Output**: Output format and schema (for JSON commands)
- **Examples**: Common usage patterns
- **Error Handling**: Common errors and solutions
- **Related Commands**: Links to related functionality

**Example**: [get commit-message](./get/commit-message.md)

## Output Formats

EAC commands produce two types of output:

### JSON Output (get commands)

Structured, machine-readable output for automation:

```bash
$ eac get modules
{
  "modules": [
    {
      "moniker": "eac-commands",
      "type": "go-commands",
      "path": "go/cli/eac",
      "dependencies": ["eac-core"],
      "files": 45
    }
  ]
}
```

**Process with jq**:

```bash
eac get modules | jq -r '.modules[].moniker'
```

**See**: [Get Commands](./get/index.md), [Output Formats](../overview/output-formats.md)

### Formatted Output (show commands)

Human-readable tables and text for interactive use:

```bash
$ eac show modules
┌───────────────┬─────────────┬────────────────────┬──────┐
│ Moniker       │ Type        │ Path               │ Files│
├───────────────┼─────────────┼────────────────────┼──────┤
│ eac-commands  │ go-commands │ go/cli/eac    │   45 │
│ eac-core      │ go-library  │ go/core        │   32 │
└───────────────┴─────────────┴────────────────────┴──────┘
```

**See**: [Show Commands](./show/index.md), [Output Formats](../overview/output-formats.md)

## Integration Examples

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Fast validation (< 10 min)
eac validate specs || exit 1
eac validate go-tidy || exit 1
eac scan --scanner secrets || exit 1
eac test --short || exit 1

echo "✓ Pre-commit checks passed"
```

### CI/CD Pipeline

```yaml
name: Build and Test

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Get Changed Modules
        id: changed
        run: |
          MODULES=$(eac get changed-modules-ci | jq -r '.changed_modules | join(" ")')
          echo "modules=$MODULES" >> $GITHUB_OUTPUT

      - name: Build and Test
        run: |
          for module in ⟪ steps.changed.outputs.modules ⟫; do
            eac build $module
            eac test $module
          done

      - name: Security Scan
        run: |
          eac scan --scanner compliance
          eac scan --scanner vuln
```

### Release Script

```bash
#!/bin/bash
# release.sh

set -e

# Check for changes
eac release pending || {
  echo "No changes to release"
  exit 0
}

# Generate changelog
eac release changelog

# Validate
eac validate release

# Check CI
eac release check-ci $(git rev-parse HEAD)

# Create release
eac release this

echo "✓ Release created"
```

## For Interactive Use

See [How-to Guides](../../../how-to-guides/eac/commands/index.md) for practical, task-oriented documentation:

- [Commit Command Guide](../../../how-to-guides/eac/commands/development-workflow/make-commits-with-ai.md) - Generate AI commit messages
- [Init Command Guide](../../../how-to-guides/local-setup/configure-ai.md) - Setup AI provider
- [Workspace Commands Guide](./work/index.md) - Use git worktrees
- [Show Commands Guide](../../../how-to-guides/eac/commands/getting-started/explore-your-repository.md) - Explore repository

## Contributing

### Adding New Commands

When adding new commands to EAC:

1. Follow [Naming Conventions](../overview/naming-conventions.md)
2. Choose appropriate category (verb-first)
3. Implement both `get` (JSON) and `show` (formatted) variants for information commands
4. Add comprehensive help text
5. Update this reference documentation

### Documentation Standards

Command reference pages should include:

- Clear purpose statement
- Complete syntax with all flags
- Output schema (for JSON commands)
- Common use cases and examples
- Error handling guidance
- Links to related commands

## See Also

### Overview

- [Command Taxonomy](../overview/command-taxonomy.md) - Organization and categories
- [Naming Conventions](../overview/naming-conventions.md) - Naming rules
- [Common Flags](../overview/common-flags.md) - Global options
- [Output Formats](../overview/output-formats.md) - JSON vs formatted

### How-to Guides

- [Command Guides](../../../how-to-guides/eac/commands/index.md) - Task-oriented guides
