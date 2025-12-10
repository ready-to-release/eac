# Creating Module Types

{{ page_breadcrumb() }}

**Problem**: You have multiple modules with the same structure (e.g., Python services, React apps) and want to define consistent defaults.

**Solution**: Create a module type in `.r2r/eac/module-types.yml` that provides templates for common patterns.

## What is a Module Type?

A module type defines:

- **Build dependencies**: What tools are needed (go, docker, npm)
- **Capabilities**: What the module can do (executable, container, api_service)
- **Default file patterns**: Standard source, config, and test locations

Modules inherit these defaults, reducing repetition.

## Quick Start

Add a type to `.r2r/eac/module-types.yml`:

```yaml
types:
  - name: python-service
    description: Python microservice with FastAPI
    build_deps: [docker]
    capabilities: [python_module, container, api_service]
    files:
      source: ["**/*.py", "requirements.txt"]
      config: ["pyproject.toml", "Dockerfile"]
      tests: ["tests/**/*.py"]
```

Use it in a module:

```yaml
modules:
  - moniker: user-api
    name: User API Service
    type: python-service
    files:
      root: src/user-api
```

## Type Definition Structure

### Minimal Type

```yaml
types:
  - name: my-type
    description: What this type represents
```

### Full Type

```yaml
types:
  - name: python-service
    description: Python microservice with FastAPI

    # What tools are needed to build
    build_deps:
      - docker        # Requires Docker
      - python        # Requires Python (for local dev)

    # What this module type can do
    capabilities:
      - python_module
      - container
      - api_service
      - openapi

    # Default file patterns (modules can override)
    files:
      source:
        - "**/*.py"
        - "requirements.txt"
        - "pyproject.toml"
      config:
        - "Dockerfile"
        - "docker-compose.yml"
        - ".env.example"
      tests:
        - "tests/**/*.py"
        - "pytest.ini"
      assets:
        - "static/**/*"
        - "templates/**/*"
```

## Build Dependencies

Available system dependencies:

| Dependency | Description |
|------------|-------------|
| `go` | Go toolchain (1.21+) |
| `docker` | Docker runtime |
| `npm` | Node.js / npm |
| `git` | Git CLI |
| `gh-cli` | GitHub CLI |
| `az-cli` | Azure CLI |

```yaml
build_deps:
  - go
  - docker
```

Modules of this type will only build if these dependencies are available.

## Capabilities

Capabilities describe what a module can do:

| Capability | Description |
|------------|-------------|
| `go_module` | Go module with go.mod |
| `executable` | Produces a binary |
| `cross_compile` | Can build for multiple platforms |
| `container` | Produces a Docker image |
| `documentation` | Documentation site |
| `serveable` | Can be served locally |
| `api_service` | Exposes an API |

```yaml
capabilities:
  - go_module
  - executable
  - cross_compile
```

Handlers use capabilities to determine build behavior.

## File Patterns

Default patterns applied to all modules of this type:

```yaml
files:
  source:           # Source code files
    - "**/*.go"
    - "!**/*_test.go"
  config:           # Configuration files
    - "go.mod"
    - "go.sum"
  tests:            # Test files
    - "**/*_test.go"
  assets:           # Static assets
    - "assets/**/*"
```

Modules can override any pattern:

```yaml
modules:
  - moniker: my-module
    type: go
    files:
      root: src/my-module
      source:                    # Override source pattern
        - "**/*.go"
        - "internal/**/*.go"
```

## Base Module Types

The unified type system uses four base types:

### Go

```yaml
- name: go
  description: Go module
  capabilities: [go_module]
  files:
    source: ["**/*.go", "**/*.go.txt"]
    config: ["go.mod", "go.sum"]
    tests: ["**/*_test.go"]
```

### Container

```yaml
- name: container
  description: Docker container module
  capabilities: [buildx]
  files:
    assets: ["Dockerfile", "**/*"]
```

### Node.js Package

```yaml
- name: node-package
  description: Node.js/TypeScript package
  build_deps: [npm]
  capabilities: [npm_module]
  files:
    source: ["src/**/*.ts", "src/**/*.js"]
    config: ["package.json", "tsconfig.json"]
    tests: ["tests/**/*.ts", "*.test.ts"]
```

### Python Service

```yaml
- name: python-service
  description: Python FastAPI service
  build_deps: [docker]
  capabilities: [python_module, container, api_service]
  files:
    source: ["**/*.py", "requirements.txt"]
    config: ["Dockerfile", "pyproject.toml"]
    tests: ["tests/**/*.py"]
```

### Documentation Site

```yaml
- name: mkdocs-site
  description: MkDocs documentation site
  build_deps: [docker]
  capabilities: [documentation, serveable, container]
  files:
    source: ["docs/**/*.md"]
    config: ["mkdocs.yml"]
    assets: ["docs/assets/**/*"]
```

### Configuration Only

```yaml
- name: configuration
  description: Configuration files (no build)
  build_deps: []
  capabilities: []
  files:
    config: ["**/*.yml", "**/*.yaml", "**/*.json"]
```

## Viewing Types

```bash
# List all types
r2r eac show moduletypes

# Get type details as JSON
r2r eac get modules --format=json | jq '.modules[].type' | sort -u
```

## Best Practices

1. **Descriptive names**: Use clear, lowercase names with hyphens
2. **Accurate dependencies**: Only list truly required build deps
3. **Minimal capabilities**: List only capabilities the type actually has
4. **Sensible defaults**: File patterns should cover 90% of cases
5. **Allow overrides**: Modules can always override defaults

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Type not found | Check spelling in module-types.yml |
| Build fails | Verify build_deps are installed |
| Wrong files matched | Adjust file patterns in type or module |
| Missing capability | Add to type's capabilities list |

## See Also

- [Creating Modules](creating-modules.md) - Use types in module definitions

{{ diataxis_footer() }}
