# Command Reference

{{ page_breadcrumb() }}

Complete technical reference for all EAC commands.

## Quick Access

- [Command Taxonomy](./overview/command-taxonomy.md) - How commands are organized
- [All Categories](./categories/index.md) - Browse by category
- [Naming Conventions](./overview/naming-conventions.md) - Command naming patterns
- [Common Flags](./overview/common-flags.md) - Global options
- [Output Formats](./overview/output-formats.md) - JSON vs human-readable output

## Command Categories

| Category | Commands | Purpose |
|----------|----------|---------|
| [create](./categories/create.md) | 7 | AI-powered generation (commits, specs, designs, PRs) |
| [get](./categories/get.md) | 18 | JSON output for automation and scripting |
| [show](./categories/show.md) | 17 | Human-readable output for interactive use |
| [validate](./categories/validate.md) | 20 | Contract and dependency validation |
| [work](./categories/work.md) | 6 | Workspace management (git worktrees) |
| [test](./categories/test.md) | 4 | Testing and test suite management |
| [build](./categories/other.md#build) | 1 | Module building |
| [pipeline](./categories/pipeline.md) | 6 | CI/CD orchestration |
| [release](./categories/release.md) | 8 | Release management and versioning |
| [scan](./categories/scan.md) | 8 | Security scanning (SAST, secrets, vulnerabilities) |
| [serve](./categories/serve.md) | 2 | Local development servers |
| [templates](./categories/templates.md) | 7 | Template management |
| [update](./categories/update.md) | 1 | Update operations |
| [other](./categories/other.md) | 3 | Help, init, extension metadata |

**Total**: 108 commands

## Common Workflows

### Module Development

- [show modules](./show/modules.md) - List all modules
- [build](./other/build.md) - Build modules
- [test](./test/test.md) - Test modules
- [validate dependencies](./validate/dependencies.md) - Check contracts

### CI/CD

- [get changed-modules-ci](./get/changed-modules-ci.md) - Find changed modules
- [get execution order](./get/execution-order.md) - Get build order
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
   - [Command Taxonomy](./overview/command-taxonomy.md) - Understand how commands are organized
   - [Naming Conventions](./overview/naming-conventions.md) - Learn command naming patterns
   - [Common Flags](./overview/common-flags.md) - Global options all commands accept

2. **Explore by category**:
   - Browse the [Categories Index](./categories/index.md) to find commands by function
   - Start with [show commands](./categories/show.md) for exploration
   - Use [get commands](./categories/get.md) for automation

3. **Try common commands**:

   ```bash
   # Discover modules
   r2r eac show modules

   # Get help for any command
   r2r eac help <command>

   # Build a module
   r2r eac build <module>

   # Run tests
   r2r eac test <module>
   ```

### Looking for a Specific Command?

- **By name**: Use the search function or browse categories
- **By purpose**: See [Command Taxonomy](./overview/command-taxonomy.md#finding-the-right-command)
- **By output format**: See [Output Formats](./overview/output-formats.md)

## Command Reference Pages

Each command has a dedicated reference page with:

- **Overview**: Command purpose and use cases
- **Syntax**: Full command syntax with options
- **Arguments & Flags**: Detailed parameter documentation
- **Output**: Output format and schema (for JSON commands)
- **Examples**: Common usage patterns
- **Error Handling**: Common errors and solutions
- **Related Commands**: Links to related functionality

**Example**: [create commit-message](./create/commit-message.md)

## Output Formats

EAC commands produce two types of output:

### JSON Output (get commands)

Structured, machine-readable output for automation:

```bash
$ r2r eac get modules
{
  "modules": [
    {
      "moniker": "eac-commands",
      "type": "go-commands",
      "path": "go/eac/commands",
      "dependencies": ["eac-core"],
      "files": 45
    }
  ]
}
```

**Process with jq**:

```bash
r2r eac get modules | jq -r '.modules[].moniker'
```

**See**: [Get Commands](./categories/get.md), [Output Formats](./overview/output-formats.md)

### Formatted Output (show commands)

Human-readable tables and text for interactive use:

```bash
$ r2r eac show modules
┌───────────────┬─────────────┬────────────────────┬──────┐
│ Moniker       │ Type        │ Path               │ Files│
├───────────────┼─────────────┼────────────────────┼──────┤
│ eac-commands  │ go-commands │ go/eac/commands    │   45 │
│ eac-core      │ go-library  │ go/eac/core        │   32 │
└───────────────┴─────────────┴────────────────────┴──────┘
```

**See**: [Show Commands](./categories/show.md), [Output Formats](./overview/output-formats.md)

## Category Highlights

### AI-Powered Commands (create)

Generate content using AI:

- **[create commit-message](./create/commit-message.md)** - Semantic commit messages from staged changes
- **[create pr](./create/pr.md)** - Pull request descriptions from branch diff
- **[create spec](./create/spec.md)** - Gherkin specifications from natural language
- **[create design](./create/design.md)** - Architecture diagrams with AI assistance

**Setup**: Run `r2r eac init` to configure your AI provider

**See**: [Create Commands Category](./categories/create.md), [Init Command](./other/init.md)

### Information Commands (get/show)

Retrieve repository information:

- **get/show modules** - Module contracts and metadata
- **get/show dependencies** - Dependency graphs
- **get/show files** - File-to-module ownership mappings
- **get/show config** - EAC configuration

**Rule**: Use `get` for JSON (automation), `show` for tables (interactive)

**See**: [Get Commands](./categories/get.md), [Show Commands](./categories/show.md)

### Quality Commands (validate/scan/test)

Ensure code quality and security:

- **validate** - Check contracts, dependencies, and specifications
- **scan** - Security scans (secrets, vulnerabilities, SAST, IaC)
- **test** - Run unit, integration, and acceptance tests

**Use in**: Pre-commit hooks, CI/CD pipelines, release gates

**See**: [Validate Commands](./categories/validate.md), [Scan Commands](./categories/scan.md), [Test Commands](./categories/test.md)

### Workflow Commands (work/release/pipeline)

Manage development workflows:

- **work** - Git worktree workspace management for parallel development
- **release** - Release preparation, changelogs, and versioning
- **pipeline** - CI/CD orchestration and monitoring

**See**: [Work Commands](./categories/work.md), [Release Commands](./categories/release.md), [Pipeline Commands](./categories/pipeline.md)

## Integration Examples

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Fast validation (< 10 min)
r2r eac validate specs || exit 1
r2r eac validate go-tidy || exit 1
r2r eac scan secrets || exit 1
r2r eac test --short || exit 1

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
          MODULES=$(r2r eac get changed-modules-ci | jq -r '.changed_modules | join(" ")')
          echo "modules=$MODULES" >> $GITHUB_OUTPUT

      - name: Build and Test
        run: |
          for module in ${{ steps.changed.outputs.modules }}; do
            r2r eac build $module
            r2r eac test $module
          done

      - name: Security Scan
        run: |
          r2r eac scan compliance
          r2r eac scan vuln
```

### Release Script

```bash
#!/bin/bash
# release.sh

set -e

# Check for changes
r2r eac release pending || {
  echo "No changes to release"
  exit 0
}

# Generate changelog
r2r eac release changelog

# Validate
r2r eac validate release

# Check CI
r2r eac release check-ci $(git rev-parse HEAD)

# Create release
r2r eac release this

echo "✓ Release created"
```

## For Interactive Use

See [How-to Guides](../../how-to-guides/eac/commands/index.md) for practical, task-oriented documentation:

- [Commit Command Guide](../../how-to-guides/eac/commands/commit-command.md) - Generate AI commit messages
- [Init Command Guide](../../how-to-guides/eac/commands/init-command.md) - Setup AI provider
- [Workspace Commands Guide](../../how-to-guides/eac/commands/areas/workspace-commands.md) - Use git worktrees
- [Show Commands Guide](../../how-to-guides/eac/commands/show-commands.md) - Explore repository

## Contributing

### Adding New Commands

When adding new commands to EAC:

1. Follow [Naming Conventions](./overview/naming-conventions.md)
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

See [Command Reference Template](../../contributing/command-reference-template.md) for the standard format.

## See Also

### Overview

- [Command Taxonomy](./overview/command-taxonomy.md) - Organization and categories
- [Naming Conventions](./overview/naming-conventions.md) - Naming rules
- [Common Flags](./overview/common-flags.md) - Global options
- [Output Formats](./overview/output-formats.md) - JSON vs formatted

### Categories

- [All Categories](./categories/index.md) - Browse all command categories
- [Create Commands](./categories/create.md) - AI-powered generation
- [Get Commands](./categories/get.md) - JSON output
- [Show Commands](./categories/show.md) - Formatted output

### How-to Guides

- [Command Guides](../../how-to-guides/eac/commands/index.md) - Task-oriented guides
- [Command Areas](../../how-to-guides/eac/commands/areas/index.md) - Functional area guides

{{ diataxis_footer() }}
