<!-- EDITOR
# Editor: how-to-guides/commands/areas/books-commands.md

## Soul

Command reference for books with auto-detecting build, show books, and validate books commands supporting PDF generation and staging directory workflows.

## Sections

1. Quick Start
2. Command Reference
3. Output Structure
4. Workflows
5. Integration with Existing Commands
6. MCP Tool Mapping
7. Error Handling
8. Best Practices
9. Troubleshooting
10. Summary
11. Next Steps
-->

# Books Commands

**Problem**: You need to build documentation with aggregated content, view configured books, and validate book configuration.

**Solution**: Use the book-related commands to build, list, and validate books.

## Quick Start

```bash
# Build a book (same as module build - auto-detects books.yml)
r2r eac build docs

# Build with PDF generation
r2r eac build docs --pdf

# Build with both PDF themes
r2r eac build docs --pdf-theme=all

# List configured books
r2r eac show books

# Validate books.yml
r2r eac validate books
```

## Command Reference

### build (with book)

When a module has a corresponding book configuration in `books.yml`, the build command automatically runs book preprocessing before MkDocs.

```bash
r2r eac build <module> [options]

# Options (same as mkdocs-site builds):
--pdf                  # Generate PDF output
--pdf-theme <theme>    # PDF theme: dark, light, or all
--accept-warnings      # Continue build despite warnings

# Examples:
r2r eac build docs                    # HTML only
r2r eac build docs --pdf              # HTML + PDF (dark theme)
r2r eac build docs --pdf-theme=all    # HTML + both PDF themes
```

**What Happens**:

1. Detects `books.yml` configuration for module
2. Creates staging directory: `out/staging/books/<moniker>/`
3. Processes copy sources (copies static files)
4. Processes command sources (executes EAC commands, captures output)
5. Generates navigation for generated sections
6. Invokes MkDocs build on staging directory
7. Outputs to `out/build/<moniker>/site/`

**Build Output**:

```text
Building module: docs (type: mkdocs-site)
Book configuration found, preprocessing...
Processing source 1/8: copy (docs/**/*.md)
Processing source 2/8: copy (docs/**/.nav.yml)
Processing source 3/8: copy (docs/assets/**)
Processing source 4/8: command (show files)
Processing source 5/8: command (show modules)
Processing source 6/8: command (show dependencies)
Processing source 7/8: command (show moduletypes)
Processing source 8/8: command (show tests)
Generating navigation...
Invoking MkDocs builder...
INFO    -  Building documentation to directory: out/build/docs/site
INFO    -  Documentation built in 4.23 seconds
```

### show books

List all configured books and their details.

```bash
r2r eac show books

# Output:
| # | Name | Description              | Sources | Module  |
|---|------|--------------------------|---------|---------|
| 1 | docs | Ready-to-Release Docs    | 8       | docs    |
```

**Columns**:

| Column      | Description                        |
| ----------- | ---------------------------------- |
| Name        | Book name (matches module moniker) |
| Description | Human-readable description         |
| Sources     | Number of sources (copy + command) |
| Module      | Associated module moniker          |

### validate books

Validate the `books.yml` configuration file.

```bash
r2r eac validate books

# Output (success):
Validating books.yml...
✓ Book 'docs' (module: docs)
  ├─ 3 copy sources (patterns valid)
  ├─ 5 command sources (commands valid)
  └─ 1 generated nav section
All books valid.

# Output (errors):
Validating books.yml...
✗ Book 'docs' (module: docs)
  ├─ ERROR: Copy source pattern 'invalid/[' is invalid
  └─ ERROR: Command 'show nonexistent' not found
Validation failed with 2 errors.
```

**Validations Performed**:

| Check          | Description                                       |
| -------------- | ------------------------------------------------- |
| Module exists  | Book name must match a module moniker             |
| Module type    | Module must be type `mkdocs-site`                 |
| Copy patterns  | Glob patterns must be syntactically valid         |
| Commands exist | Referenced commands must be valid EAC commands    |
| Nav references | `insert_into` must reference existing directories |
| Schema         | Configuration must match JSON schema              |

## Output Structure

### Build Outputs

```text
out/
├── staging/
│   └── books/
│       └── docs/                    # Staging directory (temporary)
│           ├── mkdocs.yml
│           ├── .nav.yml
│           ├── index.md
│           ├── tutorials/
│           ├── reference/
│           │   ├── .nav.yml         # Modified with generated section
│           │   └── generated/       # Generated content
│           │       ├── .nav.yml
│           │       ├── files.md
│           │       ├── modules.md
│           │       └── dependencies.md
│           └── assets/
│
└── build/
    └── docs/                        # Final build output
        ├── build.log
        ├── site/                    # HTML site
        │   ├── index.html
        │   ├── tutorials/
        │   ├── reference/
        │   │   └── generated/
        │   │       ├── files/
        │   │       ├── modules/
        │   │       └── dependencies/
        │   └── assets/
        └── pdf/                     # PDF outputs (if --pdf)
            ├── docs-dark.pdf
            └── docs-light.pdf
```

## Workflows

### Local Development

```bash
# Build and serve locally
r2r eac build docs
r2r eac serve docs

# Or build with live reload (if supported)
r2r eac serve docs --live-reload
```

### Pre-commit Validation

```bash
# Validate books before commit
r2r eac validate books

# Full validation pipeline
r2r eac validate && r2r eac build docs
```

### CI/CD Pipeline

```yaml
# .github/workflows/ci-docs.yaml
jobs:
  build:
    steps:
      - uses: actions/checkout@v4

      - name: Validate books
        run: go run ./go/eac/commands validate books

      - name: Build documentation
        run: go run ./go/eac/commands build docs

      - name: Build PDF
        run: go run ./go/eac/commands build docs --pdf-theme=all

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: documentation
          path: out/build/docs/
```

### Release Workflow

```yaml
# .github/workflows/release-docs-book.yaml
jobs:
  release:
    steps:
      - name: Build PDF book
        run: go run ./go/eac/commands build docs --pdf-theme=all

      - name: Create release
        run: |
          gh release create book/$(date +%Y.%m.%d) \
            out/build/docs/pdf/docs-dark.pdf \
            out/build/docs/pdf/docs-light.pdf \
            --title "Documentation Book $(date +%Y.%m.%d)"
```

## Integration with Existing Commands

Books work seamlessly with existing EAC commands:

### With validate

```bash
# Validate all (includes books)
r2r eac validate

# Validate books specifically
r2r eac validate books
```

### With build

```bash
# Build module (auto-detects book config)
r2r eac build docs

# Build all modules (books processed automatically)
r2r eac build
```

### With show

```bash
# Show configured books
r2r eac show books

# Show modules (docs module still appears)
r2r eac show modules
```

## MCP Tool Mapping

Books commands are available via MCP:

| MCP Tool                        | CLI Command      |
| ------------------------------- | ---------------- |
| `mcp__commands__build`          | `build docs`     |
| `mcp__commands__show-books`     | `show books`     |
| `mcp__commands__validate-books` | `validate books` |

Example MCP usage:

```bash
# Via MCP
mcp__commands__build --args "docs --pdf"
mcp__commands__show-books
mcp__commands__validate-books
```

## Error Handling

### Command Execution Failures

If a command source fails during build:

```text
Processing source 4/8: command (show files)
ERROR: Command 'show files' failed with exit code 1
       Output: error: could not read module contracts
Build failed.
```

**Resolution**: Fix the underlying command issue and retry.

### Missing Dependencies

If referenced files don't exist:

```text
Processing source 1/8: copy (docs/**/*.md)
WARNING: Pattern 'docs/**/*.md' matched 0 files

Processing source 2/8: copy (docs/**/.nav.yml)
ERROR: No .nav.yml files found
Build failed.
```

**Resolution**: Ensure source files exist and patterns are correct.

### Navigation Conflicts

If nav insertion fails:

```text
Generating navigation...
ERROR: Cannot insert into 'reference/.nav.yml': file not found
       Ensure copy sources include .nav.yml files
Build failed.
```

**Resolution**: Add `.nav.yml` files to copy sources.

## Best Practices

### Always Validate Before Build

```bash
# Validate first
r2r eac validate books

# Then build
r2r eac build docs
```

### Use Verbose Output for Debugging

```bash
# Build with verbose logging
r2r eac build docs --verbose
```

### Keep Staging for Debugging

```bash
# Examine staging directory after failed build
ls -la out/staging/books/docs/
```

### Test Commands Independently

```bash
# Test command output before adding to books.yml
r2r eac show files
r2r eac show modules
```

## Troubleshooting

| Problem                  | Solution                                    |
| ------------------------ | ------------------------------------------- |
| "Book not found"         | Ensure book name matches module moniker     |
| "Module not mkdocs-site" | Book can only build `mkdocs-site` modules   |
| "Command failed"         | Run command independently to debug          |
| "No files matched"       | Check glob pattern syntax                   |
| "Nav file not found"     | Ensure `.nav.yml` files are in copy sources |
| "PDF generation failed"  | Check Docker is running, image exists       |
| "Build timeout"          | Complex commands may need longer timeout    |

## Summary

| Command            | Purpose                             |
| ------------------ | ----------------------------------- |
| `build docs`       | Build book (auto-detects books.yml) |
| `build docs --pdf` | Build with PDF generation           |
| `show books`       | List configured books               |
| `validate books`   | Validate books.yml                  |

Books integrate transparently with existing EAC workflows—the same `build` command works for both plain MkDocs modules and books.

## Next Steps

- [Books Overview](books-overview.md) - Concept and architecture
- [Books Configuration](books-configuration.md) - Detailed configuration reference
