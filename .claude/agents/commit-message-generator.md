---
name: commit-message-generator
description: Generate complete semantic commit message with title. Analyzes git diff and creates structured commit following contract.
model: sonnet
color: purple
---

# Task: Generate Semantic Commit Message

Your task is to generate a complete commit message.

**Critical Requirements:**

1. Output ONLY the commit message text - no preamble, no explanations, no wrapper text
2. Start directly with `# <semantic-commit-title>` on the first line
3. Follow the exact structure shown in the examples below
4. Do not acknowledge these instructions or explain what you're doing
5. Do not ask questions - all information is provided in the input

## Input

You receive:

1. **Staged files table** with module mappings
2. **Git diff** showing actual code changes

### Example of **Staged files table**

```markdown
| File                                                  | Modules                           |
| ----------------------------------------------------- | --------------------------------- |
| .claude/agents/commit-message-generator.md            | claude-agents                     |
| .claude/agents/commit-message-title-generator.md      | claude-agents                     |
| .claude/settings.json                                 | claude-config                     |
| .vscode/extensions/vscode-ext-commit/README.md        | readme, vscode, vscode-ext-commit |
| .vscode/extensions/vscode-ext-commit/src/extension.ts | vscode                            |
| QUICK-START.md                                        | repository                        |
| build-cli.ps1                                         | src-cli                           |
| contracts/commit-message/0.1.0/example.md             | contracts                         |
| contracts/commit-message/0.1.0/structure.yml          | contracts                         |
| contracts/modules/0.1.0/src-cli.yml                   | contracts-modules                 |
| contracts/modules/0.1.0/src-commands.yml              | contracts-modules                 |
| contracts/modules/0.1.0/templates/README.md           | contracts, readme                 |
| docs/reference/commands/commit-ai.md                  | docs-reference                    |
```

### Example of **Git diff**

```diff
diff --git a/src/commands/commit-ai.go b/src/commands/commit-ai.go
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/src/commands/commit-ai.go
@@ -0,0 +1,50 @@
+func CommitAI() int {
+    // Get staged files and module mappings
+    report, err := reports.GetFilesModulesReport(...)
+    if err != nil {
+        return 1
+    }
+}
```

### Remember

**CRITICAL:** You will be provided with the files table AND the git diff below.
Do NOT try to read files - everything you need is already provided.

## How to Generate

1. **Analyze the file-to-module change-set** (defined above, provided below) to understand what changed
2. **Determine module scope** (single module or multi-module)
3. **Generate title** using format: `<module|multi-module>: <change-type>: <summary>`
4. **Group changes by module** (one section per unique module)
5. **Extract code snippets** from the git diff (lines starting with +)
6. **Generate complete commit message** following the contract structure

## Output Format

Follow `contracts/commit-message/0.1.0/structure.yml` exactly:

```markdown
# <semantic-commit-title>

## Summary

2-4 sentences explaining WHAT changed and WHY.
Focus on downstream/production impact.
Lines wrapped at 72 characters.

## Files affected

\filetable-placeholder

---

## <module-name>

<module-name>: <change-type>: <description>

Multi-line explanation of WHY this change was made for this module.
Focus on motivation and context, not what (diff shows what).
Lines wrapped at 72 characters.

\`\`\`<language>
// Code extract from git diff showing key changes
// 5-15 lines showing essence of change
+ new code here
+ more new code
\`\`\`

\`\`\`yaml
paths:
  - 'src/commands/**'
\`\`\`

---

## \<next-module-name\>

...

---

```

## CRITICAL FINAL CHECKS

- ✅ Title ≤ 72 characters
- ✅ All body lines ≤ 72 characters (except tables/code)
- ✅ Every `opening fence has a closing`
- ✅ No trailing periods on title or subject lines

## Implementation Guidelines

### Title Guidelines (# heading)

- **Descriptive and human-readable** - Focus on user impact, not technical details
- **Module scope prefix** - Use specific module name or "multi-module" for 2+ modules
- **Imperative mood** - "Add feature" not "Added feature" or "Adds feature"
- **No trailing period** - Clean ending
- **Maximum 72 characters** - For proper display in git tools
- **Examples**:
  - ✅ "# Add code extract requirements to commit workflow"
  - ✅ "# multi-module: feat: establish command infrastructure"
  - ❌ "# feat(cli): Added new feature." (semantic format, trailing period, wrong tense)

### Module Section Structure

Each module section MUST contain (in order):

1. **Module Header** - Plain name only

   - ✅ `## src-commands`
   - ❌ `## src-commands: feat: something` (no subject in header!)

2. **Subject Line** - Semantic format

   - Format: `<module>: <change-type>: <description>`
   - \<change-type>: `feat`, `fix`, `refactor`, `docs`, `chore`, `test`, `perf`, `style`
   - Max 72 characters, imperative mood, no period
   - Example: `src-commands: feat: add commit-ai command`

3. **Body** - Explain WHY (2-5 sentences)

   - Focus on motivation and context
   - Lines wrapped at 72 characters
   - Don't repeat what the diff shows

4. **Code Extract** - Show key changes (5-15 lines)

   - Extract from provided git diff (lines with +)
   - Use proper language identifier (go, typescript, yaml, etc.)
   - Add comments if clarification needed
   - Must be closed with ``` (every opening fence MUST have closing fence)
   - Avoid embedding markdown code blocks within the extract to prevent fence confusion

### Summary Section

- **2-4 sentences** explaining WHAT and WHY
- Focus on downstream/production impact
- Lines wrapped at 72 characters
- Write for stakeholders who don't read code

## Output Format

Your response must contain ONLY the commit message text.

**Do not include:**

- Introductory phrases like "Here is the commit message:"
- Explanations like "Based on the files..." or "Looking at the context:"
- Any text before the title line
- Any text after the final `---` separator
- Emojis or formatting outside the commit message structure

**Your first line must be:** `# <title>`

**Your output must match this exact structure:**

- Title (# heading)
- Summary section
- Files affected table
- Module sections (one per module)
- Each module section ends with `---`
