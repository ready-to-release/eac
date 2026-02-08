# Project Structure

Recommended directory layout for EAC-enabled projects.

## Basic Structure

Minimal setup for any project:

```text
your-project/
├── .clie/
│   ├── clie-cli.yml              # CLIE CLI configuration
│   └── eac/
│       └── repository.yml        # Module definitions
├── .gitignore                    # Include .clie/**/*.personal.yml
└── ... (your code)
```

## Full Structure

Complete setup with all features:

```text
your-project/
├── .clie/
│   ├── clie-cli.yml              # CLIE CLI configuration
│   ├── clie-cli.local.yml        # Local overrides (gitignored)
│   ├── cache/                   # Build cache (gitignored)
│   └── eac/
│       ├── repository.yml        # Module definitions
│       ├── ai-provider.yml       # AI configuration
│       └── ai-provider.personal.yml  # Personal AI config (gitignored)
│
├── specs/                       # Gherkin specifications (optional)
│   ├── my-module/
│   │   └── feature.feature
│   └── ...
│
├── release/                     # Release notes (optional)
│   └── my-module/
│       └── CHANGELOG.md
│
├── docs/                        # Documentation (optional)
│   └── ...
│
└── ... (your code)
```

## Configuration Files

### Always Commit

| File                   | Purpose                |
| ---------------------- | ---------------------- |
| `.clie/clie-cli.yml`     | Extension registry     |
| `.eac/repository.yml`  | Module definitions     |
| `.eac/ai-provider.yml` | AI config (no secrets) |

System defaults (in `contracts/eac-core/0.1.0/defaults/`) are used automatically and don't need to be committed.

### Never Commit (Add to .gitignore)

| File/Directory           | Purpose                       |
| ------------------------ | ----------------------------- |
| `.clie/clie-cli.local.yml` | Local CLI overrides           |
| `.eac/*.personal.yml`    | Personal configs with secrets |
| `.clie/cache/`            | Build cache                   |

### Recommended .gitignore Entries

```gitignore
# CLIE/EAC local configuration
.clie/clie-cli.local.yml
.eac/*.personal.yml
.clie/cache/
```

## Module Organization

### Go Projects

```text
your-project/
├── .eac/repository.yml
├── cmd/
│   ├── app1/
│   │   └── main.go
│   └── app2/
│       └── main.go
├── pkg/
│   ├── core/
│   │   └── *.go
│   └── utils/
│       └── *.go
└── go.mod
```

```yaml
# repository.yml
modules:
  - moniker: pkg-core
    type: go-library
    files:
      root: pkg/core

  - moniker: pkg-utils
    type: go-library
    files:
      root: pkg/utils

  - moniker: app1
    type: go-cli
    depends_on: [pkg-core, pkg-utils]
    files:
      root: cmd/app1

  - moniker: app2
    type: go-cli
    depends_on: [pkg-core]
    files:
      root: cmd/app2
```

### Single Application

```text
your-project/
├── .eac/repository.yml
├── main.go
├── internal/
│   └── ...
└── go.mod
```

```yaml
# repository.yml
modules:
  - moniker: my-app
    type: go-cli
    files:
      root: .
      source: ["**/*.go"]
      exclude: ["**/*_test.go"]
```

### Documentation Project

```text
your-project/
├── .eac/repository.yml
├── docs/
│   ├── index.md
│   └── ...
└── mkdocs.yml
```

```yaml
# repository.yml
modules:
  - moniker: docs
    type: mkdocs-site
    files:
      root: docs
      source: ["**/*.md"]
```

## Specifications (Optional)

If using Gherkin specifications:

```text
specs/
├── <module-moniker>/
│   ├── feature-name/
│   │   └── specification.feature
│   └── another-feature/
│       └── specification.feature
└── ...
```

Convention: Spec directories match module monikers.

## Release Notes (Optional)

If using automated changelogs:

```text
release/
├── <module-moniker>/
│   └── CHANGELOG.md
└── ...
```

Or at repository root:

```text
CHANGELOG.md              # Repository-wide changelog
```

## Best Practices

### 1. Start Simple

Begin with just `repository.yml`. Add other configs as needed.

### 2. One Module Per Component

Each deployable or testable unit should be its own module.

### 3. Clear File Ownership

Every source file should belong to exactly one module. Validate with:

```bash
eac validate-module-files
```

### 4. Consistent Naming

Use kebab-case for monikers: `my-service`, `api-client`, `core-lib`.

### 5. Document Dependencies

Explicitly list all module dependencies:

```yaml
depends_on: [dep1, dep2]  # Good: explicit
# Not: rely on implicit ordering
```

## Migration from Existing Projects

### Step 1: Initialize

```bash
clie init
clie install eac
eac init
```

### Step 2: Define Modules

Create `.eac/repository.yml` with your modules.

### Step 3: Validate

```bash
eac validate
eac show-modules
```

### Step 4: Test Commands

```bash
eac build <module>
eac test <module>
```

## See Also

- [Configuration](./configuration.md) - Configuration reference
- [Getting Started](./getting-started.md) - Initial setup guide
- [Modules Reference](../../eac/modules/index.md) - Module system details
