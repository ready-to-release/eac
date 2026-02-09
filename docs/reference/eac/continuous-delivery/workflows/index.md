# Workflows

GitHub Actions workflow specifications for continuous integration and RA and continuous deployment releases.

## Overview

The repository uses GitHub Actions workflows organized into four categories:

1. CI workflows for individual modules
2. Release workflows for deployment
3. Trigger orchestration workflows for change detection and execution ordering
4. Security workflows for vulnerability scanning

## In This Section

| Reference                                           | Description                                               |
| --------------------------------------------------- | --------------------------------------------------------- |
| [Overview](./overview.md)                           | Workflow architecture and design principles               |
| [Trigger Orchestration](./trigger-orchestration.md) | Change trigger orchestration system (change-trigger.yaml) |
| [CI Workflows](./ci-workflows.md)                   | Module CI workflows (ci-\*.yaml)                          |
| [Release Workflows](./release-workflows.md)         | Release workflows (release-\*.yaml)                       |
| [Security Workflows](./security-workflows.md)       | Security scanning workflows (codeql.yaml)                 |
| [Scheduled Workflows](./scheduled-workflows.md)     | Scheduled workflows (cron-\*.yaml)                        |

## Workflow Categories

### CI Workflows

Module-specific continuous integration workflows that build and test individual modules:

- `ci-eac-ai.yaml` - EAC AI module CI
- `ci-eac-commands.yaml` - EAC Commands module CI
- `ci-eac-core.yaml` - EAC Core module CI
- `ci-eac-mcp-commands.yaml` - EAC MCP Commands module CI
- `ci-eac-ext.yaml` - EAC Extension module CI
- `ci-clie.yaml` - CLIE CLI module CI
- `ci-vscode-commit.yaml` - VSCode Extension module CI
- `ci-books.yaml` - Books documentation module CI
- `ci-docs.yaml` - Docs site module CI
- `ci-clie-installer.yaml` - CLIE CLI Installer scripts CI
- `ci-implicit-cli.yaml` - Implicit CLI scripts CI

### Release Workflows

Workflows that publish deployable artifacts to production:

- `release-docs.yaml` - Documentation site release (GitHub Pages)
- `release-eac-ext.yaml` - EAC Extension release (Docker Hub)
- `release-clie.yaml` - CLIE CLI binary release (GitHub Releases)
- `release-books.yaml` - Books PDF release (GitHub Releases)

### Orchestration Workflows (3)

Workflows that coordinate CI/CD execution:

- `change-trigger.yaml` - Main CI orchestrator with incremental change detection
- `trigger-release.yaml` - Release orchestrator
- `cron-full-trigger.yaml` - Scheduled full rebuild (every 2 hours)

### Security Workflows (1)

Workflows that perform security analysis:

- `codeql.yaml` - CodeQL security scanning
