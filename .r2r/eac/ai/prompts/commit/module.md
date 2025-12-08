# Generate Module Section

Generate a module section for a multi-module commit message.

## Structure

```text
<module-name>
------------
<module>: <type>: <description>

<body: 2-4 sentences, wrapped at 72 chars>
```

## Requirements

**Line 1**: Module name only (e.g., `eac-commands`, `contracts`)

**Line 2**: Exactly 12 dashes: `------------`

**Line 3**: `<module>: <type>: <description>`

- Types: feat, fix, refactor, docs, chore, test, perf, style
- Max 72 chars, no trailing period

**Line 4**: Empty line

**Lines 5+**: Body text (2-4 sentences, wrapped at 72 chars)

## Example

```text
eac-commands
------------
eac-commands: refactor: simplify commit message generation

Removed template variable embedding from generation logic and updated
prompts to use direct format instructions. Simplified assembly code to
use blank line separators instead of dashes between module sections.
```

## Output Rules

- Output ONLY the module section
- No questions, clarifications, or explanations
- No markdown fences
- No conversational text

Generate now based on the context below:
