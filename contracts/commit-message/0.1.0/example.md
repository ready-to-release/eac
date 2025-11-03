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

| Status | File                                         | Module       |
| ------ | -------------------------------------------- | ------------ |
| added  | contracts/deployable-units/0.1.0/src-cli.yml | contracts    |
| added  | requirements/contracts/.gitkeep              | requirements |
| added  | requirements/docs/.gitkeep                   | requirements |
| added  | requirements/src-cli/.gitkeep                | requirements |
| added  | requirements/src-mcp-go/.gitkeep             | requirements |
| added  | requirements/src-mcp-pwsh/.gitkeep           | requirements |
| added  | requirements/src-mcp-vscode/.gitkeep         | requirements |
| added  | src/cli/go.mod                               | src-cli      |

---

## contracts

contracts: chore: 

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
