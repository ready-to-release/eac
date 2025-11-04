# multi-module: feat: establish foundational mono-repository infrastructure

## Summary

This commit establishes foundational infrastructure for the
mono-repository by introducing deployable unit contracts, module
definitions, and initializing the CLI module with Go workspace
configuration. These contracts define versioning strategies,
dependencies, and source paths for all major modules, enabling
trunk-based development and semantic versioning across the
repository.

## Files affected

| File                                                  | Modules                           |
| ----------------------------------------------------- | --------------------------------- |
| .claude/agents/commit-message-generator.md            | claude-agents                     |
| .claude/agents/commit-message-title-generator.md      | claude-agents                     |
| .claude/settings.json                                 | claude-config                     |
| .vscode/extensions/vscode-ext-commit/README.md        | readme, vscode, vscode-ext-commit |
| .vscode/extensions/vscode-ext-commit/src/extension.ts | vscode                            |
| QUICK-START.md                                        | repository                        |
| build-cli.ps1                                         | src-cli                           |
| contracts/commit-message/0.1.0/example.md             | contracts                         |
| contracts/commit-message/0.1.0/structure.yml          | contracts                         |
| contracts/modules/0.1.0/src-cli.yml                   | contracts-modules                 |
| contracts/modules/0.1.0/src-commands.yml              | contracts-modules                 |
| contracts/modules/0.1.0/templates/README.md           | contracts, readme                 |
| docs/reference/commands/commit-ai.md                  | docs-reference                    |
| importer.ps1                                          | repository                        |
| importer.sh                                           | repository                        |
| scripts/pwsh/go-invoker/go.psm1                       | repository                        |
| scripts/sh/go-invoker/go.sh                           | repository                        |
| scripts/sh/vscode/README.md                           | readme                            |

---

## contracts

contracts: chore: changed setting

## src-cli

src-cli: feat: add deployable unit contract

Defines module metadata including name, type, root location,
versioning scheme, and dependencies for the CLI module. Enables
automated version management and deployment tracking.

```yaml
moniker: "cli-cli"
name: "The Cli of Cli's cli"
type: "cli"
root: "src/cli"
versioning:
  version_scheme: "MAJOR.MINOR.PATCH"
used_by:
  - "claude-config"
  - "src-mcp-vscode"
```

```yaml
paths:
  - '**/src-cli/**'
```

---

## cli

cli: feat: initialize go module workspace

Establishes Go module configuration for the CLI package with
standard module path and Go version specification. Provides
foundation for dependency management and module compilation.
