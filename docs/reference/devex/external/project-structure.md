# Project Structure

Recommended directory layout for EAC-enabled projects.

## Basic Structure

Minimal setup for any project:

```
your-project/
├── .r2r/
│   ├── r2r-cli.yml              # R2R CLI configuration
│   └── eac/
│       └── repository.yml        # Module definitions
├── .gitignore                    # Include .r2r/**/*.personal.yml
└── ... (your code)
```

## Full Structure

Complete setup with all features:

```
your-project/
├── .r2r/
│   ├── r2r-cli.yml              # R2R CLI configuration
│   ├── r2r-cli.local.yml        # Local overrides (gitignored)
│   ├── cache/                   # Build cache (gitignored)
│   └── eac/
│       ├── repository.yml        # Module definitions
│       ├── module-types.yml      # Custom type templates
│       ├── environments.yml      # Test environments
│       ├── test-suites.yml       # Test suite definitions
│       ├── testing-tags.yml      # Valid test tags
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

| File | Purpose |
| ---- | ------- |
| `.r2r/r2r-cli.yml` | Extension registry |
| `.r2r/eac/repository.yml` | Module definitions |
| `.r2r/eac/module-types.yml` | Type templates |
| `.r2r/eac/environments.yml` | Test environments |
| `.r2r/eac/test-suites.yml` | Test suites |
| `.r2r/eac/ai-provider.yml` | AI config (no secrets) |

### Never Commit (Add to .gitignore)

| File/Directory | Purpose |
| -------------- | ------- |
| `.r2r/r2r-cli.local.yml` | Local CLI overrides |
| `.r2r/eac/*.personal.yml` | Personal configs with secrets |
| `.r2r/cache/` | Build cache |

### Recommended .gitignore Entries

```gitignore
# R2R/EAC local configuration
.r2r/r2r-cli.local.yml
.r2r/eac/*.personal.yml
.r2r/cache/
```

## Module Organization

### Go Projects

```
your-project/
├── .r2r/eac/repository.yml
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

```
your-project/
├── .r2r/eac/repository.yml
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

```
your-project/
├── .r2r/eac/repository.yml
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

```
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

```
release/
├── <module-moniker>/
│   └── CHANGELOG.md
└── ...
```

Or at repository root:

```
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
r2r eac validate-module-files
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
r2r init
r2r install eac
r2r eac init
```

### Step 2: Define Modules

Create `.r2r/eac/repository.yml` with your modules.

### Step 3: Validate

```bash
r2r eac validate
r2r eac show-modules
```

### Step 4: Test Commands

```bash
r2r eac build <module>
r2r eac test <module>
```

## See Also

- [Configuration](./configuration.md) - Configuration reference
- [Getting Started](./getting-started.md) - Initial setup guide
- [Modules Reference](../../eac/modules/) - Module system details
