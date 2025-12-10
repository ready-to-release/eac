# Workflows

{{ page_breadcrumb() }}

GitHub Actions workflow specifications for continuous integration and continuous delivery.

## Overview

The repository uses 19 GitHub Actions workflows organized into four categories: CI workflows for individual modules, release workflows for deployment, orchestration workflows for change detection and execution ordering, and security workflows for vulnerability scanning.

## In This Section

| Reference | Description |
|-----------|-------------|
| [Overview](./overview.md) | Workflow architecture and design principles |
| [CI Orchestration](./ci-orchestration.md) | CI orchestration system (trigger-ci.yaml) |
| [CI Modules](./ci-modules.md) | Module CI workflows (ci-*.yaml) |
| [Release Workflows](./release-workflows.md) | Release workflows (release-*.yaml) |
| [Security Workflows](./security-workflows.md) | Security scanning workflows (codeql.yaml) |
| [Scheduled Workflows](./scheduled-workflows.md) | Scheduled workflows (cron-*.yaml) |

## Workflow Categories

### CI Workflows (11)

Module-specific continuous integration workflows that build and test individual modules:

- `ci-eac-ai.yaml` - EAC AI module CI
- `ci-eac-commands.yaml` - EAC Commands module CI
- `ci-eac-core.yaml` - EAC Core module CI
- `ci-eac-mcp-commands.yaml` - EAC MCP Commands module CI
- `ci-ext-eac.yaml` - EAC Extension module CI
- `ci-r2r-cli.yaml` - R2R CLI module CI
- `ci-vscode-ext-commit.yaml` - VSCode Extension module CI
- `ci-books.yaml` - Books documentation module CI
- `ci-docs.yaml` - Docs site module CI
- `ci-scripts-cli-installer.yaml` - CLI Installer scripts CI
- `ci-scripts-implicit-cli.yaml` - Implicit CLI scripts CI

### Release Workflows (4)

Workflows that publish deployable artifacts to production:

- `release-docs.yaml` - Documentation site release (GitHub Pages)
- `release-ext-eac.yaml` - EAC Extension release (Docker Hub)
- `release-r2r-cli.yaml` - R2R CLI binary release (GitHub Releases)
- `release-books.yaml` - Books PDF release (GitHub Releases)

### Orchestration Workflows (3)

Workflows that coordinate CI/CD execution:

- `trigger-ci.yaml` - Main CI orchestrator with incremental change detection
- `trigger-release.yaml` - Release orchestrator
- `cron-full-trigger.yaml` - Scheduled full rebuild (every 2 hours)

### Security Workflows (1)

Workflows that perform security analysis:

- `codeql.yaml` - CodeQL security scanning

{{ diataxis_footer() }}
