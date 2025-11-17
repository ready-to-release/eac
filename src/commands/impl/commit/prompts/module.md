---
description: Generate one module section for commit message
model: claude-3-5-haiku-20241022
temperature: 0.3
max_tokens: 2000
---

# Generate Module Commit Section

**CRITICAL**: Generate ONE module section. Start with `## <module-name>`.

Output ONLY one module section. No explanations. No preamble.

## Output Format (YOU MUST FOLLOW EXACTLY)

Line 1: `## <module-name>`
Line 2: (blank)
Line 3: `<module-name>: <type>: <description>` (MAX 72 CHARS)
Line 4: (blank)
Line 5-N: Body text (1-3 paragraphs, wrapped at 72 chars)

## Algorithm

```
1. Read "Module Name" from input
2. Write EXACTLY: ## <module-name>
3. Write blank line
4. Write EXACTLY: <module-name>: <type>: <description> (max 72 chars)
5. Write blank line
6. Write 1-3 paragraphs describing changes for THIS module
7. STOP (no ---, no other sections)
```

## Section Header (LINE 1 - MANDATORY)

**Format**: `## <module-name>`
- Use EXACT module name from input
- No variations, no additions

## Subject Line (LINE 3 - MANDATORY, MAX 72 CHARS)

**Format**: `<module-name>: <type>: <description>`

- Start with exact module name
- Add colon + type + colon + description
- Max 72 characters TOTAL for the entire line
- No trailing period
- Types: `feat`, `fix`, `refactor`, `docs`, `chore`, `test`, `perf`, `style`

## Body (AFTER BLANK LINE)

- 1-3 paragraphs describing changes for THIS module only
- Focus on WHAT changed and WHY
- Wrap at 72 characters per line
- No code snippets, no code blocks
- STOP when done (do NOT add `---` separator)

## Example

```
## contracts

contracts: feat: add formal commit message contract specification

Added structure.yml defining conventional commit format, semantic
types, line limits, and module-specific summaries. Includes validation
rules to enforce 72-character line wrapping and conventional commit
header format for both top-level and module summaries.
```

---

## Anti-Corruption Layer

**ABSOLUTELY FORBIDDEN** - Do NOT include any of these patterns in your output:

**See**: `contracts/commit-message/0.1.0/anti-corruption.yml` for complete rules.

**Forbidden Prefixes** (lines starting with):
- "Based on"
- "Here is" / "Here's"
- "Let me"
- "I will" / "I'll" / "I've" / "I'm"
- "The title" / "The generated" / "The changes"
- "After reviewing"

**Forbidden Content**:
- Emojis (🚀 ✅ 🤖 🎉 ✨)
- Meta-commentary about the task
- Conversational wrappers
- Explanations of what you're doing

**Output ONLY**:
- ✅ The raw commit message content
- ✅ Start with `##` (module heading)
- ✅ End with body text
