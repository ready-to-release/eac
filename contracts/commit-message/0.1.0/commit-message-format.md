---
name: commit-message-format
description: Enforce the contracts/commit-message/0.1.0/structure.yml output format. Use when generating commit messages or any output that must follow the semantic commit message structure.
---

# Commit Message Format Skill

You MUST output commit messages following `contracts/commit-message/0.1.0/structure.yml`.

## Output Contract

### Structure Requirements

````markdown
# <module|multi-module>: <type>: <summary>

<2-4 sentences describing all changes, wrapped at 72 chars>

## <module-name>

<module>: <type>: <description>

<Detailed changes wrapped at 72 chars>

//OPTIONAL SOURCE CODE CHANGE
```<language>
<Code extract showing WHAT changed>
```

---

## <another-module>

<module>: <type>: <description>

<... same as <module-name section>>

---

````

### Critical Rules

**Top-level heading** (Line 1):

- Format: `# <module|multi-module>: <type>: <summary>`
- Max 72 characters
- Use `multi-module` when changes span multiple modules
- Examples:
  - `# multi-module: feat: add commit message validation`
  - `# src-cli: fix: resolve command parsing error`

**Top-level body** (After heading, no section header):

- 2-6 sentences describing ALL changes
- Every line MUST be ≤72 characters
- Wrap naturally at word boundaries
- NO section header like "## Summary"

**Module sections**:

- Header: `## <module-name>` (plain name, no colons)
- Subject line: `<module>: <type>: <description>` (≤72 chars)
- Body: Detailed explanation (every line ≤72 chars)
- Code extract: Shows WHAT changed
- YAML paths: Glob patterns for affected files

**Semantic types**:

- `feat` - New feature
- `fix` - Bug fix
- `refactor` - Code restructuring
- `docs` - Documentation
- `chore` - Maintenance
- `test` - Tests
- `perf` - Performance
- `style` - Formatting

### Anti-Corruption Layer

**ABSOLUTELY FORBIDDEN**:

- ❌ "Based on"
- ❌ "Here is"
- ❌ "Here's"
- ❌ "Let me"
- ❌ "I will/I'll/I've"
- ❌ "The title"
- ❌ "After reviewing"
- ❌ Emojis
- ❌ Wrapper text
- ❌ Meta-commentary
- ❌ Conversational filler

**Output ONLY**:
✅ The raw commit message content
✅ Start with `#` (heading)
✅ End with final `---`

### Line Length Enforcement

HARD LIMIT: 72 characters for first line of output with the section 1 markdown header.

### Validation

The contract defines these error codes:

- `MD041` - First line must be top-level heading with format
- `MD022` - Headers should be surrounded by blank lines
- `MD026` - No trailing punctuation in headings

### Example

See. [Example](./example.md)

### Rules

See. [structure](./structure.md)

## Reminder

When this skill is active:

1. Output ONLY the commit message
2. NO greetings, NO explanations
3. Follow the structure EXACTLY
4. NO meta-commentary
