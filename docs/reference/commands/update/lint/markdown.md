# Markdown Linting

Markdown files are linted using [markdownlint-cli2](https://github.com/DavidAnson/markdownlint-cli2), the same engine used by the VS Code markdownlint extension.

## Configuration

Two configuration files define the linting rules and ignores:

| File                      | Purpose                         |
| ------------------------- | ------------------------------- |
| `.markdownlint.yml`       | Lint rules (226 rule configs)   |
| `.markdownlint-cli2.yaml` | Extends rules + defines ignores |

### Config Structure

```text
workspace/
├── .markdownlint.yml          # Rules (source of truth)
├── .markdownlint-cli2.yaml    # Ignores + extends rules
└── docs/
    └── *.md                   # Linted files
```

### Why Two Files?

The VS Code extension and CLI both use `markdownlint-cli2` internally. Having separate files for rules and ignores provides:

- **`.markdownlint.yml`** - Detailed, human-readable rule configuration
- **`.markdownlint-cli2.yaml`** - Machine-focused: extends rules, defines ignores

Both files are read automatically by the CLI and VS Code extension.

## Ignored Paths

Default ignores in `.markdownlint-cli2.yaml`:

```yaml
ignores:
  - "node_modules/**"
  - "**/node_modules/**"
  - "out/**"
  - ".git/**"
  - ".vscode/**"
  - ".cache/**"
  - "docs/assets/cache/**"
```

## Usage

```bash
# Lint markdown in a module
r2r eac update lint repository

# Lint with auto-fix
r2r eac update lint repository --fix

# All static/docs modules
r2r eac update lint
```

## Auto-Fix Support

The `--fix` flag automatically fixes many issues:

- Trailing spaces
- Missing blank lines around lists/headings
- Inconsistent list markers
- Line length (where possible)

Issues that cannot be auto-fixed are reported for manual correction.

## Table Formatting

MD060 (table-column-style) requires aligned tables but **cannot be auto-fixed** by markdownlint-cli2. Use [Prettier](https://prettier.io/) to format tables:

```bash
# Format all markdown tables
npx prettier --write "docs/**/*.md"

# Check without modifying
npx prettier --check "docs/**/*.md"
```

Install Prettier globally:

```bash
npm install -g prettier
```

Configuration is in `.prettierrc`.

## Enabled Rules

Key rules from `.markdownlint.yml`:

| Rule  | Description                                |
| ----- | ------------------------------------------ |
| MD001 | Heading levels increment by one            |
| MD009 | Trailing spaces                            |
| MD012 | Multiple consecutive blank lines           |
| MD013 | Line length (2000 chars, 100 for headings) |
| MD022 | Headings surrounded by blank lines         |
| MD025 | Single top-level heading                   |
| MD032 | Lists surrounded by blank lines            |
| MD040 | Fenced code blocks have language           |
| MD047 | Files end with newline                     |

See `.markdownlint.yml` for complete rule configuration.

## IDE Integration

### VS Code

Install the [markdownlint extension](https://marketplace.visualstudio.com/items?itemName=DavidAnson.vscode-markdownlint):

```bash
code --install-extension DavidAnson.vscode-markdownlint
```

The extension automatically discovers `.markdownlint-cli2.yaml` and `.markdownlint.yml` from the workspace root.

### Recommended Settings

```json
// .vscode/settings.json
{
  "[markdown]": {
    "editor.wordWrap": "on"
  }
}
```

The extension is already included in `.vscode/extensions.json` recommendations.

## Output Example

```text
=== Linting Markdown files ===
Running: markdownlint-cli2 [**/*.md]
markdownlint-cli2 v0.20.0 (markdownlint v0.40.0)
Finding: **/*.md !node_modules/** !out/** !.git/**
Linting: 487 file(s)
Summary: 3 error(s)

docs/README.md:10 MD040/fenced-code-language Fenced code blocks should have a language specified
docs/guide.md:45 MD032/blanks-around-lists Lists should be surrounded by blank lines
docs/api.md:100:14 MD060/table-column-style Table column style [Table pipe does not align]

✗ Markdown lint issues found
```

## System Requirements

| Tool              | Minimum Version | Install Command                    |
| ----------------- | --------------- | ---------------------------------- |
| Node.js           | 18+             | (see nodejs.org)                   |
| markdownlint-cli2 | any             | `npm install -g markdownlint-cli2` |
| Prettier          | any             | `npm install -g prettier`          |

Verify installation:

```bash
npm list -g markdownlint-cli2 --depth=0
# markdownlint-cli2@0.20.0
```

## Troubleshooting

### "markdownlint-cli2 not found"

Install globally:

```bash
npm install -g markdownlint-cli2
```

On Windows, ensure Node.js bin directory is in PATH.

### Ignores not working

Ensure `.markdownlint-cli2.yaml` exists in workspace root with correct format:

```yaml
config:
  extends: .markdownlint.yml

ignores:
  - "node_modules/**"
```

### Too many errors

Run with `--fix` first to auto-correct formatting issues:

```bash
r2r eac update lint --fix
```

Then address remaining semantic issues manually.

## See Also

- [update lint](./lint.md) - Main lint command
- [update lint (Go)](./go.md) - Go linting
- [markdownlint rules](https://github.com/DavidAnson/markdownlint/blob/main/doc/Rules.md)
- [markdownlint-cli2 docs](https://github.com/DavidAnson/markdownlint-cli2)
