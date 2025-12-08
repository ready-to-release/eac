<!-- EDITOR
# Editor: how-to-guides/commands/areas/books-configuration.md

## Soul

Configuration reference for books.yml structure with copy, command, and inline source types plus navigation synthesis for generated content sections.

## Sections

1. Configuration File
2. Book Definition
3. Source Types
4. Copy Sources
5. Command Sources
6. Inline Sources
7. Navigation Configuration
8. Complete Example
9. Staging Directory
10. Validation
11. Schema Reference
12. Best Practices
13. Troubleshooting
14. Next Steps
-->

# Books Configuration

**Problem**: You need to define how a book aggregates static content with dynamically-generated content.

**Solution**: Configure books in `.r2r/eac/books.yml` with copy sources for static files and command sources for generated content.

## Configuration File

Books are configured in `.r2r/eac/books.yml`:

```yaml
# .r2r/eac/books.yml
books:
  - name: docs                              # Module moniker
    description: "Ready-to-Release Docs"    # Human description
    sources:
      - type: copy
        from: "docs/**/*.md"
        to: "./"
      - type: command
        command: "show modules"
        target: "reference/generated/modules.md"
```

## Book Definition

Each book in the `books` array has these fields:

| Field           | Required | Description                                                         |
| --------------- | -------- | ------------------------------------------------------------------- |
| `name`          | Yes      | Module moniker this book builds (must match a `mkdocs-site` module) |
| `description`   | No       | Human-readable description                                          |
| `sources`       | Yes      | Array of content sources (copy or command)                          |
| `generated_nav` | No       | Navigation configuration for generated sections                     |

## Source Types

Books support three source types:

| Type      | Purpose                                    | Output                   |
| --------- | ------------------------------------------ | ------------------------ |
| `copy`    | Copy static files from repository          | Files in staging         |
| `command` | Execute command, create new file           | New file in staging      |
| `inline`  | Execute command, insert into existing file | Modified file in staging |

### Copy Sources

Copy static files from the repository to the staging directory.

```yaml
- type: copy
  from: "docs/**/*.md"          # Glob pattern (relative to repo root)
  to: "./"                      # Target directory (relative to staging)
  exclude:                      # Optional exclusion patterns
    - "**/draft-*.md"
    - "**/WIP/**"
```

| Field     | Required | Description                   |
| --------- | -------- | ----------------------------- |
| `type`    | Yes      | Must be `copy`                |
| `from`    | Yes      | Glob pattern for source files |
| `to`      | Yes      | Target directory in staging   |
| `exclude` | No       | Patterns to exclude from copy |

**Path Behavior**:

- Source paths are relative to repository root
- Target paths are relative to staging directory
- Directory structure from glob match is preserved

**Example**:

```yaml
- type: copy
  from: "docs/tutorials/**/*.md"
  to: "tutorials/"
```

Copies `docs/tutorials/getting-started.md` → `staging/tutorials/getting-started.md`

### Command Sources

Execute an EAC command and write its markdown output to a file.

```yaml
- type: command
  command: "show modules"                   # EAC command to execute
  target: "reference/generated/modules.md"  # Output file path
  frontmatter:                              # Optional YAML frontmatter
    title: "Modules"
    description: "Repository module listing"
    hide:
      - navigation
  order: 1                                  # Sort order in navigation
```

| Field         | Required | Description                                      |
| ------------- | -------- | ------------------------------------------------ |
| `type`        | Yes      | Must be `command`                                |
| `command`     | Yes      | EAC command to execute (e.g., `show files`)      |
| `target`      | Yes      | Output file path in staging                      |
| `frontmatter` | No       | YAML frontmatter to prepend to output            |
| `order`       | No       | Sort order within generated section (default: 0) |

**Available Commands**:

Any EAC command that outputs markdown can be used:

| Command             | Output                          |
| ------------------- | ------------------------------- |
| `show files`        | Repository file inventory table |
| `show modules`      | Module contracts table          |
| `show dependencies` | Module dependency graph         |
| `show moduletypes`  | Module type definitions         |
| `show environments` | Environment contracts           |
| `show tests`        | Test suite listings             |

**Generated File Format**:

```markdown
---
title: "Modules"
description: "Repository module listing"
---

<!-- Generated by: show modules -->
<!-- Generated at: 2024-01-15T10:30:00Z -->

| # | Moniker | Type | Root Path |
|---|---------|------|-----------|
| 1 | docs | mkdocs-site | docs |
| 2 | eac-commands | go-commands | go/eac/commands |
...
```

### Inline Sources

Execute EAC commands and insert their output at marker positions in existing files.

```yaml
- type: inline
  target: "reference/overview.md"       # File to modify (in staging)
  marker_pattern: "..."                 # Optional: override default regex
  inserts:
    - marker: "show-modules"            # Marker ID to find in file
      command: "show modules"           # Command to execute
    - marker: "show-dependencies"
      command: "show dependencies"
```

| Field               | Required | Description                                           |
| ------------------- | -------- | ----------------------------------------------------- |
| `type`              | Yes      | Must be `inline`                                      |
| `target`            | Yes      | Path to file in staging (must exist after copy phase) |
| `marker_pattern`    | No       | Regex with capture group for marker ID                |
| `inserts`           | Yes      | Array of marker-to-command mappings                   |
| `inserts[].marker`  | Yes      | Marker ID to match (captured from regex)              |
| `inserts[].command` | Yes      | EAC command to execute                                |

**Default Marker Pattern**:

```regex
<!--\s*book:insert\s+([a-zA-Z0-9_-]+)\s*-->
```

**Use Case**:

Place markers in your markdown where dynamic content should appear:

```markdown
# Repository Overview

This repository contains the following modules:

<!-- book:insert show-modules -->

## Dependencies

The dependency relationships are:

<!-- book:insert show-dependencies -->
```

**Processing**:

1. Inline commands execute during Step 2 (with command sources)
2. Output is stored in memory (not written to new files)
3. During Step 5, markers are found and replaced with command output
4. Output is wrapped with generation markers for traceability

**Output Format**:

Markers are replaced with wrapped output:

```markdown
<!-- book:generated show-modules -->
| # | Moniker | Type |
|---|---------|------|
| 1 | docs | mkdocs-site |
<!-- /book:generated -->
```

**Custom Marker Patterns**:

Override the default pattern for different marker styles:

```yaml
# Jinja-style: {{ insert modules }}
marker_pattern: "\\{\\{\\s*insert\\s+([a-zA-Z0-9_-]+)\\s*\\}\\}"

# Shortcode-style: [book:modules]
marker_pattern: "\\[book:([a-zA-Z0-9_-]+)\\]"
```

**Validation**:

- Target file must exist in staging (after copy phase)
- Marker pattern must be valid regex with capture group
- Warning if marker ID not found in file
- Commands must be valid EAC commands

## Navigation Configuration

### Generated Nav Sections

Configure how generated content appears in navigation:

```yaml
generated_nav:
  - section: "reference/generated"    # Path to generated section
    title: "Generated Reference"      # Title in nav
    insert_into: "reference"          # Parent nav to insert into
    position: "after:index.md"        # Where to insert
```

| Field         | Required | Description                                          |
| ------------- | -------- | ---------------------------------------------------- |
| `section`     | Yes      | Path to the generated section directory              |
| `title`       | No       | Title for the section in navigation                  |
| `insert_into` | Yes      | Parent directory whose `.nav.yml` to modify          |
| `position`    | No       | `first`, `last`, or `after:<item>` (default: `last`) |

### How Nav Generation Works

1. **Copy Phase**: `.nav.yml` files from source docs are copied to staging
2. **Generate Phase**: New `.nav.yml` created for generated sections
3. **Insert Phase**: Generated sections inserted into parent navs

**Example**:

Given:

```yaml
generated_nav:
  - section: "reference/generated"
    title: "Generated"
    insert_into: "reference"
    position: "after:index.md"
```

Before (source `reference/.nav.yml`):

```yaml
title: Reference
nav:
  - index.md
  - decision-records
  - specifications
```

After (staging `reference/.nav.yml`):

```yaml
title: Reference
nav:
  - index.md
  - generated        # ← Inserted
  - decision-records
  - specifications
```

Generated (`reference/generated/.nav.yml`):

```yaml
title: Generated
nav:
  - modules.md
  - files.md
  - dependencies.md
```

## Complete Example

```yaml
# .r2r/eac/books.yml
books:
  - name: docs
    description: "Ready-to-Release Documentation"

    sources:
      # Static content from docs/
      - type: copy
        from: "docs/**/*.md"
        to: "./"

      - type: copy
        from: "docs/**/.nav.yml"
        to: "./"

      - type: copy
        from: "docs/assets/**"
        to: "assets/"

      # Generated reference content
      - type: command
        command: "show files"
        target: "reference/generated/files.md"
        frontmatter:
          title: "Repository Files"
          description: "Complete file inventory with module ownership"
        order: 1

      - type: command
        command: "show modules"
        target: "reference/generated/modules.md"
        frontmatter:
          title: "Modules"
          description: "All modules in the repository"
        order: 2

      - type: command
        command: "show dependencies"
        target: "reference/generated/dependencies.md"
        frontmatter:
          title: "Dependencies"
          description: "Module dependency relationships"
        order: 3

      - type: command
        command: "show moduletypes"
        target: "reference/generated/module-types.md"
        frontmatter:
          title: "Module Types"
          description: "Available module type definitions"
        order: 4

      - type: command
        command: "show tests"
        target: "reference/generated/tests.md"
        frontmatter:
          title: "Test Suites"
          description: "All test suites in the repository"
        order: 5

      # Inline content inserted at markers in existing files
      - type: inline
        target: "index.md"
        inserts:
          - marker: "modules-summary"
            command: "show modules"

      - type: inline
        target: "reference/overview.md"
        inserts:
          - marker: "module-types"
            command: "show moduletypes"
          - marker: "dependencies"
            command: "show dependencies"

    # Navigation for generated sections
    generated_nav:
      - section: "reference/generated"
        title: "Generated Reference"
        insert_into: "reference"
        position: "after:index.md"
```

## Staging Directory

During build, content is staged before MkDocs runs:

```text
out/staging/books/docs/
├── .nav.yml                    # Copied from docs/
├── index.md                    # Modified: marker "modules-summary" replaced
├── tutorials/
│   ├── .nav.yml
│   └── *.md
├── reference/
│   ├── .nav.yml                # Modified: includes "generated"
│   ├── index.md
│   ├── overview.md             # Modified: markers replaced with content
│   ├── generated/              # Created by book builder
│   │   ├── .nav.yml            # Generated
│   │   ├── files.md            # From: show files
│   │   ├── modules.md          # From: show modules
│   │   ├── dependencies.md     # From: show dependencies
│   │   ├── module-types.md     # From: show moduletypes
│   │   └── tests.md            # From: show tests
│   └── decision-records/
├── explanation/
└── assets/
```

## Validation

Validate your books.yml configuration:

```bash
r2r eac validate books

Validating books.yml...
✓ Book 'docs' (module: docs)
  ├─ 3 copy sources (patterns valid)
  ├─ 5 command sources (commands valid)
  └─ 1 generated nav section
All books valid.
```

## Schema Reference

Books configuration is validated against `contracts/eac-core/0.1.0/books.schema.json`.

Key validations:

- `name` must match a module moniker of type `mkdocs-site`
- `sources` must have at least one entry
- `command` must be a valid EAC command
- `from` patterns must be valid globs
- `position` must be `first`, `last`, or `after:<item>`

## Best Practices

### Source Ordering

Process copy sources before command sources to ensure static files exist when commands reference them:

```yaml
sources:
  # 1. Copy static content first
  - type: copy
    from: "docs/**/*.md"
    to: "./"

  # 2. Then generate dynamic content
  - type: command
    command: "show modules"
    target: "reference/generated/modules.md"
```

### Explicit Nav Copies

Always explicitly copy `.nav.yml` files:

```yaml
- type: copy
  from: "docs/**/.nav.yml"
  to: "./"
```

### Frontmatter for SEO

Use frontmatter on generated files for better navigation and SEO:

```yaml
- type: command
  command: "show modules"
  target: "reference/generated/modules.md"
  frontmatter:
    title: "Modules"              # Page title
    description: "Module listing" # Meta description
    hide:
      - toc                       # Hide table of contents
```

### Consistent Ordering

Use `order` field to maintain consistent navigation order:

```yaml
- type: command
  command: "show files"
  target: "reference/generated/files.md"
  order: 1

- type: command
  command: "show modules"
  target: "reference/generated/modules.md"
  order: 2
```

## Troubleshooting

| Problem                      | Solution                                                |
| ---------------------------- | ------------------------------------------------------- |
| Book not found               | Ensure `name` matches a module moniker in `modules.yml` |
| Module type mismatch         | Book name must reference a `mkdocs-site` module         |
| Copy pattern matches nothing | Check glob pattern syntax, ensure files exist           |
| Command not found            | Verify command exists with `r2r eac help <command>`     |
| Nav not updated              | Ensure `insert_into` points to existing directory       |
| Generated files missing      | Check command execution logs in build output            |

## Next Steps

- [Books Overview](books-overview.md) - Concept and architecture
- [Books Commands](books-commands.md) - Command reference for working with books
