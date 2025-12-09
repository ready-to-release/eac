# Other Commands

{{ page_breadcrumb() }}

## Overview

Other commands provide essential functionality that doesn't fit neatly into other categories. This includes project initialization, build execution, help documentation, and extension metadata.

**Key Characteristics**:

- Essential project operations
- Configuration and setup
- Build execution
- Documentation access
- Extension integration

**When to use**: For project setup, building artifacts, accessing help, and extension integration.

## All Other Commands

| Command | Purpose | Use Case |
|---------|---------|----------|
| [init](../other/init.md) | Initialize AI provider configuration | Set up API keys and AI provider settings |
| [build](../other/build.md) | Build modules | Compile and package modules |
| [help](../other/help.md) | Display help information | Get command documentation |
| [show help](../show/help.md) | Show help in formatted output | Interactive help browsing |
| [extension-meta](../other/extension-meta.md) | Output extension metadata | r2r CLI integration |

## Build Commands

### Overview

The `build` command compiles and packages modules respecting dependency order. It's a fundamental command for producing deployable artifacts.

**When to use**:

- During development to compile code changes
- In CI/CD pipelines for artifact creation
- Before testing to ensure latest code is compiled
- Before deployment to produce release artifacts

### Basic Build Usage

```bash
# Build single module
r2r eac build src-auth

# Build multiple modules
r2r eac build src-auth src-api

# Build all modules
r2r eac build --all

# Build with dependencies
r2r eac build r2r-cli
# Automatically builds all dependencies first
```

### Build Order

Builds respect dependency order automatically:

```bash
# If r2r-cli depends on eac-commands which depends on eac-core
r2r eac build r2r-cli

# Builds in order:
# 1. eac-core (no dependencies)
# 2. eac-commands (depends on eac-core)
# 3. r2r-cli (depends on eac-commands)
```

### Incremental Builds

Build only changed modules:

```bash
# Get changed modules
CHANGED=$(r2r eac get changed-modules | jq -r '.changed_modules[]')

# Build only what changed
r2r eac build $CHANGED
```

### Clean Builds

Force rebuild from scratch:

```bash
# Clean before build
r2r eac build src-auth --clean

# Or manually clean
rm -rf bin/ && r2r eac build --all
```

### Build Output

Artifacts are placed in configured output directories:

```text
bin/
├── eac              # eac-commands binary
├── r2r              # r2r-cli binary
└── lib/
    ├── auth.a       # src-auth library
    └── api.a        # src-api library
```

### Build Performance

```bash
# Show build times
r2r eac show build-times

# Find slow builds
r2r eac get build-times | jq '[.builds[]] | sort_by(.duration) | reverse'

# Parallel builds (if supported)
r2r eac build --parallel
```

### Build in CI/CD

```bash
# Build changed modules in CI
r2r eac get changed-modules-ci | jq -r '.changed_modules[]' | while read module; do
  r2r eac build $module
done

# Validate artifacts exist
r2r eac validate artifacts r2r-cli
```

### Build Integration

```bash
# Build and test
r2r eac build src-auth && r2r eac test src-auth

# Build and run
r2r eac build r2r-cli && ./bin/r2r --version

# Build before deployment
r2r eac build --all && r2r eac scan vuln
```

### Common Build Issues

**Build fails with dependency errors**:

```bash
# Check dependency order
r2r eac get execution order src-auth

# Build dependencies first
r2r eac build eac-core eac-commands src-auth
```

**Stale artifacts**:

```bash
# Clean and rebuild
rm -rf bin/ pkg/
r2r eac build --all
```

**Missing dependencies**:

```bash
# Validate dependencies
r2r eac validate dependencies

# Check module contracts
r2r eac show modules
```

## Initialization Commands

### init Command

Configure AI provider for EAC commands:

```bash
# Interactive configuration
r2r eac init

# Prompts for:
# - AI provider (Anthropic, OpenAI, Google, etc.)
# - API key
# - Model selection
# - Configuration file location
```

**Configuration file locations**:

- `.eac/ai-config.yml` - Project-specific
- `~/.eac/ai-config.yml` - User-specific
- `EAC_AI_PROVIDER` environment variable - Session-specific

**Example configuration**:

```yaml
provider: anthropic
api_key: sk-ant-...
model: claude-3-5-sonnet-20241022
```

### When to Run init

**First time setup**:

```bash
# Clone repository
git clone https://github.com/your-org/project
cd project

# Initialize EAC
r2r eac init
```

**Switching providers**:

```bash
# Reconfigure for different provider
r2r eac init

# Override with environment variable
export EAC_AI_PROVIDER=openai
export EAC_AI_API_KEY=sk-...
```

**CI/CD setup**:

```yaml
# GitHub Actions example
- name: Configure EAC
  env:
    EAC_AI_PROVIDER: anthropic
    EAC_AI_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: r2r eac init --non-interactive
```

## Help Commands

### help Command

Display command help information:

```bash
# General help
r2r eac help

# Command-specific help
r2r eac help build
r2r eac help test suite

# Category help
r2r eac help show
```

### show help Command

Display help in formatted table:

```bash
# All commands
r2r eac show help

# Filtered by category
r2r eac show help | grep "show"
```

### Help Flag

Available on all commands:

```bash
# Using --help flag
r2r eac build --help
r2r eac test suite --help

# Short form
r2r eac build -h
```

## Extension Metadata

### extension-meta Command

Output metadata for r2r CLI integration:

```bash
# Get extension metadata
r2r eac extension-meta

# Output:
# {
#   "name": "eac",
#   "version": "1.0.0",
#   "commands": [...],
#   "description": "Engineering Automation Commands"
# }
```

**Purpose**: Enables the `r2r` CLI to discover and integrate EAC commands as an extension.

**Usage in r2r**:

```bash
# r2r discovers eac via extension-meta
r2r eac show modules

# Without r2r
eac show modules
```

## Common Workflows

### New Project Setup

```bash
# 1. Clone repository
git clone https://github.com/your-org/project
cd project

# 2. Initialize AI provider
r2r eac init

# 3. Validate repository
r2r eac validate

# 4. Build all modules
r2r eac build --all

# 5. Run tests
r2r eac test --all
```

### Development Cycle

```bash
# 1. Make code changes
# ... edit files ...

# 2. Build changed modules
r2r eac build src-auth

# 3. Run tests
r2r eac test src-auth

# 4. Commit
r2r eac work commit --all
```

### CI/CD Pipeline

```bash
# 1. Validate repository structure
r2r eac validate

# 2. Build changed modules
CHANGED=$(r2r eac get changed-modules-ci | jq -r '.changed_modules[]')
for module in $CHANGED; do
  r2r eac build $module
done

# 3. Run tests
r2r eac test $CHANGED

# 4. Validate artifacts
for module in $CHANGED; do
  r2r eac validate artifacts $module
done

# 5. Deploy if successful
# ... deployment steps ...
```

## Configuration

### Build Configuration

Configure build behavior in module contracts:

```yaml
# .eac/contracts/modules/src-auth.yml
moniker: src-auth
type: go-library
path: go/src/auth
build:
  output: bin/lib/auth.a
  flags: ["-ldflags", "-s -w"]
  env:
    CGO_ENABLED: "0"
```

### AI Configuration

Configure AI provider:

```yaml
# .eac/ai-config.yml
provider: anthropic
api_key: ${ANTHROPIC_API_KEY}
model: claude-3-5-sonnet-20241022
max_tokens: 4096
temperature: 0.7
```

### Environment Variables

```bash
# AI Configuration
export EAC_AI_PROVIDER=anthropic
export EAC_AI_API_KEY=sk-ant-...
export EAC_AI_MODEL=claude-3-5-sonnet-20241022

# Build Configuration
export EAC_BUILD_PARALLEL=true
export EAC_BUILD_CLEAN=false

# General Configuration
export EAC_DEBUG=true
export EAC_VERBOSE=true
```

## Best Practices

### Build Practices

1. **Build incrementally during development**

   ```bash
   r2r eac build src-auth  # Not --all
   ```

2. **Clean builds for releases**

   ```bash
   r2r eac build --all --clean
   ```

3. **Validate after building**

   ```bash
   r2r eac build r2r-cli && r2r eac validate artifacts r2r-cli
   ```

4. **Build before testing**

   ```bash
   r2r eac build src-auth && r2r eac test src-auth
   ```

### Configuration Practices

1. **Use project-specific configuration**

   ```bash
   # Create .eac/ai-config.yml in repository
   # Don't commit API keys
   ```

2. **Use environment variables in CI**

   ```yaml
   env:
     EAC_AI_API_KEY: ${{ secrets.API_KEY }}
   ```

3. **Document required configuration**

   ```markdown
   # Setup
   1. Run `r2r eac init`
   2. Configure API key
   3. Run `r2r eac validate`
   ```

### Help Practices

1. **Always check help first**

   ```bash
   r2r eac build --help
   ```

2. **Use show help for browsing**

   ```bash
   r2r eac show help | less
   ```

3. **Check examples in help output**

   ```bash
   r2r eac help create commit-message
   ```

## Command Details

### init

Initialize AI provider configuration:

```bash
# Interactive setup
r2r eac init

# Non-interactive (CI)
r2r eac init --non-interactive \
  --provider anthropic \
  --api-key $ANTHROPIC_API_KEY
```

**See**: [init command reference](../other/init.md)

### build

Build one or more modules:

```bash
# Single module
r2r eac build src-auth

# Multiple modules
r2r eac build src-auth src-api

# All modules
r2r eac build --all

# With options
r2r eac build src-auth --clean --verbose
```

**See**: [build command reference](../other/build.md)

### help

Display help information:

```bash
# General help
r2r eac help

# Command help
r2r eac help <command>

# Subcommand help
r2r eac help <command> <subcommand>
```

**See**: [help command reference](../other/help.md)

### extension-meta

Output extension metadata for r2r CLI:

```bash
r2r eac extension-meta

# Used internally by r2r CLI
# Not typically called directly
```

**See**: [extension-meta command reference](../other/extension-meta.md)

## See Also

- [validate](../categories/validate.md) - Validate repository structure
- [show modules](../show/modules.md) - List all modules
- [test](../categories/test.md) - Run tests
- [work](../categories/work.md) - Workspace management
- [get modules](../get/modules.md) - Module metadata
- [Build Guide](../../../how-to-guides/eac/building/overview.md)
- [Configuration Guide](../../../how-to-guides/eac/configuration/overview.md)

{{ diataxis_footer() }}
