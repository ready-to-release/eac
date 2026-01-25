# Understanding Configuration Files

{{ page_breadcrumb() }}

Learn about the configuration files in `.r2r/eac/` and how they're managed by EAC.

**Prerequisites:** Completed [Quick Start](./quick-start.md)

## What You'll Learn

By the end of this tutorial, you'll understand:

- The difference between R2R CLI and EAC configuration
- What configuration files exist in `.r2r/`
- Which files are created by commands vs. exist as system defaults
- How to customize configurations when needed
- Which files to commit to git

## Configuration Overview

R2R uses two layers of configuration:

| Layer         | File/Directory     | Purpose               | Created By     |
| ------------- | ------------------ | --------------------- | -------------- |
| **Framework** | `.r2r/r2r-cli.yml` | Extension management  | `r2r init`     |
| **Extension** | `.r2r/eac/`        | EAC-specific settings | `r2r eac init` |

### R2R CLI vs EAC Configuration

Understanding the two configuration layers:

**`.r2r/r2r-cli.yml`** (Framework Configuration)

- Created by: `r2r init`
- Purpose: Manages which extensions are available
- Scope: Framework-level (applies to all extensions)
- Example content: Extension registry, Docker images

**`.r2r/eac/`** (EAC Extension Configuration)

- Created by: `r2r eac init`
- Purpose: Configures how the EAC extension behaves
- Scope: Extension-specific (only affects EAC)
- Example content: AI provider settings, module configuration

See [CLI vs Extensions](../../reference/eac/architecture/cli-integration.md) for detailed architecture explanation.

## R2R CLI Configuration

### The `.r2r/r2r-cli.yml` File

This file is created by `r2r init` and manages extension installation:

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/ext-eac:latest'
    description: 'Everything-as-Code automation'
```

**Purpose:**

- Defines which extensions are available
- Specifies Docker images for each extension
- Configures registry and pull policies

**When to edit:**

- Adding new extensions (or use `r2r install <extension>`)
- Changing extension versions
- Configuring local development images

**Should you commit it?** ✅ Yes - team needs to know which extensions are used

For complete reference, see [R2R CLI Configuration Guide](../../reference/r2r/commands/configuration.md).

## EAC Extension Configuration

### The `.r2r/eac/` Directory

After running `r2r eac init`, you'll have a `.r2r/eac/` directory in your repository. This directory contains configuration files that control how the EAC extension works.

## File Categories

EAC configuration files fall into three categories:

### 1. User-Specific Files (Created by Init)

These files are created when you run `r2r eac init` and contain your team or personal settings:

| File                       | Purpose                                      | Commit?                     |
| -------------------------- | -------------------------------------------- | --------------------------- |
| `ai-provider.yml`          | AI provider configuration (API keys, models) | Optional (team config)      |
| `ai-provider.personal.yml` | Personal AI overrides                        | ❌ Never (personal secrets) |

**Example: `ai-provider.yml`**

```yaml
ai:
  provider: claude-api
  model: claude-3-haiku-20240307
  endpoint: https://api.anthropic.com/v1
  api_key: ${ANTHROPIC_API_KEY}

git:
  token: ${GIT_TOKEN}
```

**When to edit:**

- Change AI model
- Switch AI providers
- Update API endpoints

---

### 2. System Defaults (Automatic Fallback)

These files ship with EAC and provide default configurations. You **don't need to create them** - EAC automatically uses them.

| File                      | Purpose                                                                         | Need to Copy?               |
| ------------------------- | ------------------------------------------------------------------------------- | --------------------------- |
| `ai-config.yml`           | AI type definitions (specs, commit-message)                                     | ❌ No (uses system default) |
| `module-types.yml`        | Module type definitions                                                         | ❌ No (uses system default) |
| `system-dependencies.yml` | System dependency definitions                                                   | ❌ No (uses system default) |
| `security-tools.yml`      | Security tool configurations                                                    | ❌ No (uses system default) |
| `logging.yml`             | Logging configuration                                                           | ❌ No (uses system default) |
| `environments.yml`        | Test environment definitions                                                    | ❌ No (uses system default) |
| `test-suites.yml`         | Test suite definitions (unit, integration, acceptance, production-verification) | ❌ No (uses system default) |
| `testing-tags.yml`        | Test tag definitions                                                            | ❌ No (uses system default) |

**How it works:**

EAC uses a **layered fallback system** (similar to Git config):

1. **User override** (`.r2r/eac/ai-config.yml`) - checked first
2. **System default** (built into EAC) - fallback if not found
3. **Hardcoded default** (in code) - last resort

> **Example: Using system defaults**

```bash
# This command works automatically - no ai-config.yml needed!
r2r eac create spec my-module

# EAC automatically:
# 1. Checks .r2r/eac/ai-config.yml (not found)
# 2. Falls back to built-in system default ✓
# 3. Uses that configuration
```

**When to customize:**

Most users never need to customize these files. The system defaults work for typical use cases.

!!! info "Advanced Customization"
If you need to customize these configurations (rare), you can create your own versions in `.r2r/eac/`. Your versions will take precedence over system defaults. See advanced documentation for details.

---

### 3. Generated Files (Created by Commands)

These files are created by EAC commands and stored in your repository:

| File             | Created By                | Purpose                                    | Commit? |
| ---------------- | ------------------------- | ------------------------------------------ | ------- |
| `repository.yml` | `r2r eac analyze modules` | Repository metadata and discovered modules | ✅ Yes  |
| `books.yml`      | `r2r eac analyze books`   | Architecture patterns found in code        | ✅ Yes  |

**Note on `test-suites.yml`**: This file has **system defaults** (see Section 2 above) providing standard test suites (unit, integration, acceptance, production-verification). You can optionally generate a customized version:

| File              | Created By                         | Purpose                          | Commit?                |
| ----------------- | ---------------------------------- | -------------------------------- | ---------------------- |
| `test-suites.yml` | `r2r eac analyze tests` (optional) | Custom test suite configurations | ✅ Yes (if customized) |

**Example: `repository.yml` (generated)**

```yaml
repository:
  name: eac
  type: monorepo
  root: /var/task

modules:
  - name: core
    path: go/eac/core
    type: lib
    language: go

  - name: commands
    path: go/eac/commands
    type: cli
    language: go
```

**When created:**

```bash
# Run analyze to create repository.yml and books.yml
r2r eac analyze modules

# Optional: Generate custom test-suites.yml (system defaults work automatically)
r2r eac analyze tests
```

**Can you edit them?**

Yes! You can edit generated files after creation. But running the command again with `--force` will overwrite your changes:

```bash
# Regenerate (overwrites your edits!)
r2r eac analyze modules --force
```

---

## Configuration Workflow

### Minimal Setup (Recommended)

For most users, you only need to create user-specific files:

```bash
# 1. Initialize with AI provider
r2r eac init --ai-provider claude-api

# 2. Set API key in environment
export ANTHROPIC_API_KEY=sk-ant-your-key-here

# 3. Run commands (uses system defaults automatically!)
r2r eac analyze modules
r2r eac create spec my-module

# Result: Clean .r2r/eac/ with only user-specific files
```

**Files created:**

```text
.r2r/eac/
├── ai-provider.yml          (your provider config)
└── (all other files use built-in system defaults)
```

---

### Advanced Customization (Optional)

<!-- markdownlint-disable MD046 -->

!!! note "Most Users Don't Need This"

    The minimal setup covers 99% of use cases. Only proceed if you have specific customization requirements.

<!-- markdownlint-disable MD046 -->

If you need custom AI type definitions or other advanced configurations, you can copy the system defaults:

```bash
# Copy system default files to your repository for customization
r2r eac init --copy-templates
```

This copies configuration files like `ai-config.yml`, `module-types.yml`, etc. to `.r2r/eac/`. You can then:

1. Edit the copied files to customize them
2. Commit your customizations to version control
3. EAC will automatically use your versions instead of system defaults

**Example file structure after copying:**

```text
.r2r/eac/
├── ai-provider.yml          (created by init)
├── ai-config.yml            (copied with --copy-templates)
├── module-types.yml         (copied with --copy-templates)
└── (other copied files...)
```

For details on the file formats, see the [Init Command Reference](../../reference/eac/commands/init/init.md#advanced-copying-system-templates).

---

## What to Commit to Git

### Always Commit

✅ Generated files (documents your repository structure)

- `repository.yml`
- `books.yml`

✅ Team configuration (optional, team decides)

- `ai-provider.yml` (if sharing team config)

✅ Custom overrides (if you customized them)

- `ai-config.yml` (only if you created a custom version)
- `module-types.yml` (only if you created a custom version)
- `test-suites.yml` (only if you generated/customized it)
- `testing-tags.yml` (only if you created a custom version)
- etc.

### Never Commit

❌ Personal secrets and preferences

- `*.personal.yml` (personal API keys)
- `*.local.yml` (local overrides)

### Auto-Gitignored

EAC automatically adds these patterns to `.gitignore`:

```gitignore
# EAC local configuration (never commit)
.r2r/*.local.yml
.r2r/eac/*.personal.yml
```

---

## Common Scenarios

### Scenario 1: "I just want it to work"

Use the minimal setup - no customization needed:

```bash
r2r eac init --ai-provider claude-api
export ANTHROPIC_API_KEY=your-key
r2r eac analyze modules
```

**Result:** One config file (`ai-provider.yml`), everything else uses system defaults.

---

### Scenario 2: "I want to change the AI model"

Edit your provider config:

```bash
# Edit ai-provider.yml
vim .r2r/eac/ai-provider.yml

# Change the model:
# model: claude-3-opus-20240229

# Commands now use new model
r2r eac create spec my-module
```

---

### Scenario 3: "I want to customize AI types"

!!! tip "Rare Use Case"
Most users don't need custom AI types. The defaults support all standard workflows.

If you have specific requirements, copy the system defaults first:

```bash
# Copy system template files
r2r eac init --copy-templates

# Edit the file
vim .r2r/eac/ai-config.yml

# Commit your customization
git add .r2r/eac/ai-config.yml
git commit -m "Custom AI configuration"
```

Your custom configuration will override system defaults.

---

### Scenario 4: "I need to reset to defaults"

Delete your custom override to use system default again:

```bash
# Remove custom override
rm .r2r/eac/ai-config.yml

# Now uses system default again
r2r eac create spec my-module
```

---

## File Resolution Priority

When EAC loads configuration files, it checks locations in order:

```text
Priority 1: Your repository (.r2r/eac/ai-config.yml)
         ↓ not found
Priority 2: Built-in system defaults
         ↓ not found
Priority 3: Hardcoded defaults (in code)
```

This means:

- **If you create a file locally, it takes precedence**
- **If not, system defaults are used automatically**
- **You only customize what you need**

---

## Upgrading

### With System Defaults (Recommended)

If you're using system defaults (most users):

**Automatic upgrade:**

- Install new version of r2r CLI
- System defaults automatically updated
- No configuration changes needed

### With Custom Overrides

If you've created custom configuration files:

**Manual review:**

- Install new version of r2r CLI
- Review release notes for configuration changes
- Update your custom files if needed

---

## Summary

| File Type       | When Created   | Customizable  | Commit        |
| --------------- | -------------- | ------------- | ------------- |
| User-Specific   | By `init`      | Yes           | Optional      |
| System Defaults | Built into EAC | Via override  | If overridden |
| Generated       | By commands    | Yes (careful) | Yes           |

**Key Takeaway:** You only need to run `r2r eac init`. Everything else works automatically!

## Next Steps

- **Next tutorial:** [Creating Your First Extension](./creating-your-first-extension.md) - Build a custom r2r extension
- **Learn about specifications:** [BDD Fundamentals](../../explanation/specifications/concepts/bdd-fundamentals.md) - Understand Gherkin and BDD
- **If customizing (advanced):** See reference documentation for configuration file formats

{{ diataxis_footer() }}
