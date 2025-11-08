# multi-module: feat: establish foundational mono-repository infrastructure

This commit establishes foundational infrastructure for the
mono-repository by introducing deployable unit contracts, module
definitions, and initializing the CLI module with Go workspace
configuration. These contracts define versioning strategies,
dependencies, and source paths for all major modules, enabling
trunk-based development and semantic versioning across the
repository.

## contracts

contracts: feat: add commit message contract specification

Added structure.yml defining conventional commit format, semantic
types, line limits, and module-specific summaries. Updated validation
rules to enforce 72-character line limits and conventional commit
header format for both top-level and module summaries.

---

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
