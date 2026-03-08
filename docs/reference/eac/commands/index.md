# Command Reference

This section documents **EAC commands** - the automation tools provided by the EAC CLI.

## EAC Commands

Complete technical reference for all EAC commands (invoked as `eac <command>`).

## Quick Access

- [Language Support](./language-support.md) - Which commands work with which languages
- [Command Taxonomy](./overview/command-taxonomy.md) - How commands are organized
- [All Categories](./categories/index.md) - Browse by category
- [Naming Conventions](./overview/naming-conventions.md) - Command naming patterns
- [Common Flags](./overview/common-flags.md) - Global options
- [Output Formats](./overview/output-formats.md) - JSON vs human-readable output

## Command Categories

| Category                               | Purpose                                              |
| -------------------------------------- | ---------------------------------------------------- |
| [build](./categories/build.md)         | Module building and compilation                      |
| [create](./categories/create.md)       | AI-powered generation (commits, specs, designs, PRs) |
| [get](./categories/get.md)             | JSON output for automation and scripting             |
| [help](./categories/help.md)           | Display help information                             |
| [init](./categories/init.md)           | Project initialization                               |
| [pipeline](./categories/pipeline.md)   | CI/CD orchestration                                  |
| [release](./categories/release.md)     | Release management and versioning                    |
| [scan](./categories/scan.md)           | Security scanning (SAST, secrets, vulnerabilities)   |
| [serve](./categories/serve.md)         | Local development servers                            |
| [show](./categories/show.md)           | Human-readable output for interactive use            |
| [templates](./categories/templates.md) | Template management                                  |
| [test](./categories/test.md)           | Testing and test suite management                    |
| [update](./categories/update.md)       | Update operations                                    |
| [validate](./categories/validate.md)   | Contract and dependency validation                   |
| [work](./categories/work.md)           | Workspace management (git worktrees)                 |

## Common Workflows

### Module Development

- [show modules](./show/modules.md) - List all modules
- [build](./build/build.md) - Build modules
- [test](./test/test.md) - Test modules
- [validate dependencies](./validate/dependencies.md) - Check contracts

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
- **By purpose**: See [Command Taxonomy](./overview/command-taxonomy.md#finding-the-right-command)
- **By output format**: See [Output Formats](./overview/output-formats.md)

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

**See**: [Get Commands](./categories/get.md), [Output Formats](./overview/output-formats.md)

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

**See**: [Show Commands](./categories/show.md), [Output Formats](./overview/output-formats.md)

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

- [Command Guides](../../../how-to-guides/eac/commands/index.md) - Task-oriented guides
