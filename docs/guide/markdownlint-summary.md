# Markdownlint Documentation Summary

> **Source:** [markdownlint v0.38.0 documentation](https://github.com/DavidAnson/markdownlint/tree/v0.38.0/doc) > **Generated:** 2025-10-25

## Overview

Markdownlint is a comprehensive markdown linting tool that enforces style consistency and best practices across markdown documents. It includes 56 built-in rules covering headings, whitespace, lists, links, code blocks, tables, and accessibility.

---

## Core Concepts

### Rule Structure

Each rule contains:

- **ID**: MD### (e.g., MD001, MD013)
- **Name**: Descriptive identifier (e.g., `heading-increment`, `line-length`)
- **Tags**: Categories for grouping (e.g., `headings`, `whitespace`, `accessibility`)
- **Aliases**: Alternative names for easier reference
- **Configuration**: Rule-specific parameters with defaults
- **Fixable**: Whether auto-fix is supported

### Configuration System

- **Global Default**: All rules enabled by default (`default: true`)
- **Extends**: Support for configuration inheritance
- **Per-Rule Config**: Individual rules can be enabled/disabled or configured
- **Multiple Formats**: JSON, JSONC, YAML, and JS configuration files supported

---

## Complete Rule Index (56 Rules)

### Heading Rules

| Rule  | Name                             | Description                            | Key Config                          |
| ----- | -------------------------------- | -------------------------------------- | ----------------------------------- |
| MD001 | heading-increment                | Heading levels increment by one        | -                                   |
| MD003 | heading-style                    | Consistent heading style               | `style: "consistent"`               |
| MD018 | no-missing-space-atx             | Space after hash required              | -                                   |
| MD019 | no-multiple-space-atx            | No multiple spaces after hash          | -                                   |
| MD020 | no-missing-space-closed-atx      | Space inside closed atx headings       | -                                   |
| MD021 | no-multiple-space-closed-atx     | No multiple spaces in closed atx       | -                                   |
| MD022 | blanks-around-headings           | Blank lines around headings            | `lines_above: 1`, `lines_below: 1`  |
| MD023 | heading-start-left               | Headings start at line beginning       | -                                   |
| MD024 | no-duplicate-heading             | No duplicate heading content           | `siblings_only: false`              |
| MD025 | single-title/single-h1           | Only one top-level heading             | `level: 1`                          |
| MD026 | no-trailing-punctuation          | No trailing punctuation in headings    | `punctuation: ".,;:!。，；：！"`    |
| MD036 | no-emphasis-as-heading           | Don't use emphasis instead of headings | `punctuation: ".,;:!?。，；：！？"` |
| MD041 | first-line-heading/first-line-h1 | First line should be top-level heading | `level: 1`, `allow_preamble: false` |
| MD043 | required-headings                | Enforce required heading structure     | `headings: []`, `match_case: false` |

### Whitespace Rules

| Rule  | Name                         | Description                         | Key Config                               |
| ----- | ---------------------------- | ----------------------------------- | ---------------------------------------- |
| MD009 | no-trailing-spaces           | No trailing spaces                  | `br_spaces: 2`, `strict: false`          |
| MD010 | no-hard-tabs                 | No hard tabs                        | `code_blocks: true`, `spaces_per_tab: 1` |
| MD012 | no-multiple-blanks           | Max consecutive blank lines         | `maximum: 1`                             |
| MD027 | no-multiple-space-blockquote | No multiple spaces after blockquote | -                                        |
| MD028 | no-blanks-blockquote         | No blank lines in blockquote        | -                                        |
| MD030 | list-marker-space            | Spaces after list markers           | `ul_single: 1`, `ul_multi: 1`            |
| MD037 | no-space-in-emphasis         | No spaces inside emphasis markers   | -                                        |
| MD038 | no-space-in-code             | No spaces inside code spans         | -                                        |
| MD039 | no-space-in-links            | No spaces inside link text          | -                                        |
| MD047 | single-trailing-newline      | Files end with single newline       | -                                        |

### List Rules

| Rule  | Name                | Description                     | Key Config                     |
| ----- | ------------------- | ------------------------------- | ------------------------------ |
| MD004 | ul-style            | Consistent unordered list style | `style: "consistent"`          |
| MD005 | list-indent         | Consistent list indentation     | -                              |
| MD007 | ul-indent           | Unordered list indentation      | `indent: 2`, `start_indent: 2` |
| MD029 | ol-prefix           | Ordered list item prefix        | `style: "one_or_ordered"`      |
| MD032 | blanks-around-lists | Lists surrounded by blank lines | -                              |

### Link & Image Rules

| Rule  | Name                             | Description                          | Key Config                                                 |
| ----- | -------------------------------- | ------------------------------------ | ---------------------------------------------------------- |
| MD011 | no-reversed-links                | No reversed link syntax              | -                                                          |
| MD034 | no-bare-urls                     | No bare URLs                         | -                                                          |
| MD042 | no-empty-links                   | No empty links                       | -                                                          |
| MD045 | no-alt-text                      | Images require alt text              | -                                                          |
| MD051 | link-fragments                   | Link fragments must be valid         | `ignore_case: false`                                       |
| MD052 | reference-links-images           | Reference labels must be defined     | `ignored_labels: ["x"]`                                    |
| MD053 | link-image-reference-definitions | Reference definitions must be needed | `ignored_definitions: ["//"]`                              |
| MD054 | link-image-style                 | Consistent link/image style          | All styles enabled by default                              |
| MD059 | link-text                        | Link text should be descriptive      | `prohibited_texts: ["click here", "here", "link", "more"]` |

### Code Block Rules

| Rule  | Name                 | Description                             | Key Config                                      |
| ----- | -------------------- | --------------------------------------- | ----------------------------------------------- |
| MD014 | commands-show-output | Dollar signs in commands without output | -                                               |
| MD031 | blanks-around-fences | Fenced code blocks surrounded by blanks | `list_items: true`                              |
| MD040 | fenced-code-language | Fenced code blocks need language        | `allowed_languages: []`, `language_only: false` |
| MD046 | code-block-style     | Consistent code block style             | `style: "consistent"`                           |
| MD048 | code-fence-style     | Consistent code fence style             | `style: "consistent"`                           |

### Table Rules

| Rule  | Name                 | Description                      | Key Config            |
| ----- | -------------------- | -------------------------------- | --------------------- |
| MD055 | table-pipe-style     | Consistent table pipe style      | `style: "consistent"` |
| MD056 | table-column-count   | Consistent table column count    | -                     |
| MD058 | blanks-around-tables | Tables surrounded by blank lines | -                     |

### Emphasis & Styling Rules

| Rule  | Name           | Description               | Key Config            |
| ----- | -------------- | ------------------------- | --------------------- |
| MD049 | emphasis-style | Consistent emphasis style | `style: "consistent"` |
| MD050 | strong-style   | Consistent strong style   | `style: "consistent"` |

### Other Rules

| Rule  | Name           | Description                      | Key Config                                         |
| ----- | -------------- | -------------------------------- | -------------------------------------------------- |
| MD013 | line-length    | Line length limit                | `line_length: 80`, `strict: false`, `stern: false` |
| MD033 | no-inline-html | No inline HTML                   | `allowed_elements: []`                             |
| MD035 | hr-style       | Consistent horizontal rule style | `style: "consistent"`                              |
| MD044 | proper-names   | Proper name capitalization       | `names: []`, `code_blocks: true`                   |

---

## Key Configuration Parameters

### MD013 - Line Length (Most Configurable)

```yaml
MD013:
  line_length: 80              # Base line limit
  heading_line_length: 80      # Separate limit for headings
  code_block_line_length: 80   # Separate limit for code blocks
  code_blocks: true            # Check code blocks
  tables: true                 # Check tables
  headings: true               # Check headings
  strict: false                # Report all violations
  stern: false                 # Warn on fixable violations
```

**Exception Handling:**

- Lines without spaces beyond limit are exempted
- Link/image reference definitions always exempted
- Standalone links/images exempted

### MD043 - Required Headings (Structure Enforcement)

```yaml
MD043:
  headings: []        # Required heading sequence
  match_case: false   # Case-sensitive matching
```

**Wildcard Support:**

- `"*"` - Zero or more unspecified headings
- `"+"` - One or more unspecified headings
- `"?"` - Exactly one unspecified heading

### MD054 - Link/Image Styles

```yaml
MD054:
  autolink: true              # <https://example.com>
  inline: true                # [text](url)
  full: true                  # [text][ref]
  collapsed: true             # [ref][]
  shortcut: true              # [ref]
  url_inline: true            # [url](url) - convert to autolink
```

---

## Custom Rules

### Creating Custom Rules

Custom rules extend markdownlint functionality through the `options.customRules` parameter.

**Required Components:**

- **Names**: Array of identifiers for the rule
- **Description**: Purpose explanation
- **Tags**: Category groupings
- **Parser**: Data source (`"micromark"`, `"markdown-it"`, or `"none"`)
- **Function**: Implementation receiving `params` and `onError` callback

### Parser Options

- **Micromark** (preferred): Structured token-based parsing
- **Markdown-it**: Alternative parser with different token structure
- **None**: Direct text processing

### Error Reporting

`onError` callback accepts:

- `lineNumber`: Line of violation
- `detail`: Additional context
- `context`: Surrounding text
- `range`: Character position
- `fixInfo`: Auto-fix instructions

### Async Support

Custom rules can return Promises for async operations (not supported in synchronous contexts).

### Helper Package

`markdownlint-rule-helpers` provides shared utilities for rule development.

### Simple Alternative

`markdownlint-rule-search-replace` plugin enables text-replacement rules without coding.

---

## Integration & Tools

### Prettier Integration

- **Compatibility**: Minimal conflicts with default settings
- **Configuration**: Use `prettier.json` extension to disable overlapping rules
- **Tab Width Adjustments**: When using `--tab-width 4`:
  - Set `ul_multi: 3` and `ul_single: 3`
  - Set `ul-indent: 4`

### Validation

Two JSON schemas available:

1. **Standard**: `markdownlint-config-schema.json` - Allows custom rules
2. **Strict**: `markdownlint-config-schema-strict.json` - Built-in rules only

**Validation Methods:**

```json
// Add to .markdownlint.json
"$schema": "https://raw.githubusercontent.com/DavidAnson/markdownlint/main/schema/markdownlint-config-schema.json"
```

```bash
# Command-line validation
npx ajv-cli validate -s ./markdownlint/schema/markdownlint-config-schema.json -d "**/.markdownlint.{json,yaml}" --strict=false
```

### Release Process

Staggered release schedule across packages:

1. **markdownlint** (core library)
2. **markdownlint-cli2** (CLI tool)
3. **markdownlint-cli2-action** (GitHub Action)
4. **vscode-markdownlint** (VS Code extension)
5. **markdownlint-cli** (alternative CLI)

Schedule allows flexibility based on release content and feature scope.

---

## Best Practices

### Accessibility Rules

- **MD001**: Maintain heading hierarchy for screen readers
- **MD045**: Require alt text for images
- **MD059**: Enforce descriptive link text

### Code Quality

- **MD013**: Enforce line length for editor compatibility
- **MD040**: Require language specs for syntax highlighting
- **MD046**: Maintain consistent code block style

### Document Structure

- **MD041**: Start files with top-level heading
- **MD043**: Enforce organizational heading structure
- **MD022**: Use blank lines around headings for readability

### Style Consistency

- **MD003/004**: Consistent heading and list styles
- **MD046/048/049/050**: Consistent code, emphasis, and strong styles
- **MD054/055**: Consistent link, image, and table styles

---

## Configuration Examples

### Minimal Config (Disable Specific Rules)

```yaml
# Disable line-length and inline HTML rules
MD013: false
MD033: false
```

### Relaxed Line Length

```yaml
MD013:
  line_length: 120
  code_block_line_length: 120
  heading_line_length: 120
  tables: false
```

### Strict Document Structure

```yaml
MD041:
  level: 1
  allow_preamble: false

MD043:
  headings:
    - "# Title"
    - "## Overview"
    - "## Requirements"
    - "*"
    - "## Conclusion"
  match_case: true
```

### Custom HTML Elements Allowed

```yaml
MD033:
  allowed_elements:
    - br
    - div
    - span
    - img
```

---

## File Formats Supported

- `.markdownlint.json` - JSON configuration
- `.markdownlint.jsonc` - JSON with comments
- `.markdownlint.yaml` / `.markdownlint.yml` - YAML configuration
- JavaScript configuration files (via CLI tools)

---

## Summary Statistics

- **Total Rules**: 56 (MD001-MD059, with some numbers skipped)
- **Fixable Rules**: Many rules support auto-fix
- **Configuration Parameters**: 100+ across all rules
- **Rule Categories**: 7 main categories (headings, whitespace, lists, links, code, tables, accessibility)
- **Default Behavior**: All rules enabled unless explicitly disabled
