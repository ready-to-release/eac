# Books: Documentation Aggregation

{{ page_breadcrumb() }}

**Problem**: Documentation sites need both static content (hand-written markdown) and dynamic content (generated from repository state like module listings, file inventories, dependency graphs).

**Solution**: The **book** system aggregates static markdown from `docs/` with dynamically-generated content from EAC commands, producing a unified MkDocs site.

## What is a Book?

A book is a **preprocessing layer** for `mkdocs-site` modules that:

1. **Copies** static markdown files from source directories
2. **Executes** EAC commands to generate new files
3. **Synthesizes** navigation from both sources
4. **Inserts** inline command output into existing files
5. **Builds** the final site using MkDocs

```text
┌─────────────────────────────────────────────────────────────────┐
│                        Book Build Flow                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Static Content       New File Commands     Inline Commands     │
│  ┌────────────┐       ┌────────────┐       ┌─────────────┐      │
│  │ docs/*.md  │       │ show files │       │ show modules│      │
│  │ docs/.nav  │       │ show deps  │       │ (insert at  │      │
│  │ docs/assets│       │ → new file │       │  line 10)   │      │
│  └─────┬──────┘       └─────┬──────┘       └─────┬───────┘      │
│        │                    │                    │              │
│        ▼                    ▼                    ▼              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              Book Preprocessor (5 Steps)                 │   │
│  │                                                          │   │
│  │  Step 1: Copy static files to staging                    │   │
│  │  Step 2: Execute commands (new files + inline outputs)   │   │
│  │  Step 3: Generate .nav.yml for generated sections        │   │
│  │  Step 4: Insert generated sections into parent navs      │   │
│  │  Step 5: Insert inline command outputs at line numbers   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                           │                                     │
│                           ▼                                     │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              MkDocs Build (Docker)                       │   │
│  │  - HTML site generation                                  │   │
│  │  - Optional PDF generation                               │   │
│  └──────────────────────────────────────────────────────────┘   │
│                           │                                     │
│                           ▼                                     │
│                    ┌──────────────┐                             │
│                    │   Output     │                             │
│                    │ site/ + pdf/ │                             │
│                    └──────────────┘                             │
└─────────────────────────────────────────────────────────────────┘
```

## Key Concepts

### Source Types

Books define **sources** that specify where content comes from:

| Type      | Description                         | Example                                         |
| --------- | ----------------------------------- | ----------------------------------------------- |
| `copy`    | Static files copied from repository | `docs/**/*.md`                                  |
| `command` | EAC command output → new file       | `show files` → `generated/files.md`             |
| `inline`  | EAC command output → replace marker | `show modules` → `<!-- book:insert modules -->` |

### Transparent Activation

Books integrate seamlessly with existing workflows:

- **No new module type**: The `docs` module remains type `mkdocs-site`
- **Auto-detection**: When `books.yml` exists, book preprocessing runs automatically
- **Same commands**: `build docs` works unchanged—it detects and uses books.yml

### Navigation Synthesis

Books automatically handle navigation for generated content:

1. **Copy** `.nav.yml` files from source docs
2. **Generate** `.nav.yml` for command-generated sections
3. **Insert** generated sections into parent navigation

## When to Use Books

### Use Books When

- Your documentation includes **generated reference content** (file listings, module tables, dependency graphs)
- You want **single-source-of-truth** documentation that reflects actual repository state
- You need to **combine** hand-written guides with auto-generated reference material

### Don't Use Books When

- Your documentation is **entirely static** (no generated content)
- You only have a **few manually-maintained reference pages**
- The overhead of command execution isn't worth the automation benefit

## Architecture Integration

Books fit into the EAC ecosystem without disruption:

```text
┌──────────────────────────────────────────────────────────────────┐
│                    EAC Module System                             │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  modules.yml          module-types.yml         handlers.yml      │
│  ┌─────────────┐       ┌──────────────┐         ┌─────────────┐  │
│  │ docs:       │       │ mkdocs-site  │         │ mkdocs:     │  │
│  │   type:     │─────▶│   caps:      │────────▶│   handler   │  │
│  │   mkdocs-   │       │   [doc,      │         │             │  │
│  │   site      │       │    container]│         └──────┬──────┘  │
│  └─────────────┘       └──────────────┘                │         │
│                                                        │         │
│  books.yml                                             │         │
│  ┌─────────────┐                                       │         │
│  │ books:      │───────────────────────────────────────┘         │
│  │   - name:   │  "if books.yml exists,                          │
│  │     docs    │   run book preprocessing"                       │
│  │     sources:│                                                 │
│  │     - copy  │                                                 │
│  │     - cmd   │                                                 │
│  └─────────────┘                                                 │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### What Stays the Same

| Component          | Status                                  |
| ------------------ | --------------------------------------- |
| `modules.yml`      | No changes—`docs` remains `mkdocs-site` |
| `module-types.yml` | No changes—no new types needed          |
| `handlers.yml`     | No changes—dispatch rules unchanged     |
| CI workflows       | No changes—`build docs` works as before |

### What's New

| Component                         | Purpose                      |
| --------------------------------- | ---------------------------- |
| `.r2r/eac/books.yml`              | Book definitions and sources |
| `contracts/.../books.schema.json` | JSON schema for validation   |
| `show books` command              | List configured books        |
| `validate books` command          | Validate books.yml           |

## Quick Example

Given this `books.yml`:

```yaml
books:
  - name: docs
    sources:
      # Copy static files
      - type: copy
        from: "docs/**/*.md"
        to: "./"

      # Generate new file from command
      - type: command
        command: "show modules"
        target: "reference/generated/modules.md"

      # Insert command output at markers in existing file
      - type: inline
        target: "reference/overview.md"
        inserts:
          - marker: "dependencies"
            command: "show dependencies"
```

Where `reference/overview.md` contains:

```markdown
# Overview

## Dependencies

<!-- book:insert dependencies -->
```

Running `build docs` will:

1. Copy all markdown from `docs/` to staging
2. Execute `show modules` → write to `reference/generated/modules.md`
3. Execute `show dependencies` → store output in memory
4. Generate navigation for the new generated section
5. Find `<!-- book:insert dependencies -->` marker and replace with output
6. Build the MkDocs site with static, generated, and inline content

## Next Steps

- [Books Configuration](books-configuration.md) - Detailed configuration reference
- [Books Commands](books-commands.md) - Command reference for working with books

{{ diataxis_footer() }}
