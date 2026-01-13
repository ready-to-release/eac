# Generate Top-Level Commit Message

You are an expert in writing clear, professional commit messages.

Generate a commit message summarizing all changes across the repository.

Your task is to generate structured JSON data. The command will automatically format it into a conventional commit message.

## JSON Output Structure

Generate a JSON object matching this schema:

```json
{
  "type": "feat",
  "scope": "module-name",
  "subject": "brief description of change",
  "body": "2-4 sentences explaining what and why.\n\nProvide context and rationale.",
  "breaking": false
}
```

### JSON Field Requirements

**type** (required): One of: `feat`, `fix`, `refactor`, `docs`, `chore`, `test`, `perf`, `style`, `ci`, `build`

**scope** (required): For single-module: the module name. For multi-module: use `multi-module`
- Pattern: lowercase with hyphens (e.g., `eac-commands`, `eac-core`)
- Length: 1-20 characters
- Must match actual module names from the context

**subject** (required): Brief description of the change
- Max 72 characters total when combined with `type(scope): `
- No trailing period
- Lowercase first letter
- Imperative mood (e.g., "add feature" not "adds feature")

**body** (optional but recommended): Detailed explanation
- First sentence becomes the Auditor-Summary (auto-extracted)
- 2-4 sentences explaining what changed and why
- Provide context and rationale
- Can use newlines (`\n\n`) to separate paragraphs

**breaking** (optional): Set to `true` if this is a breaking change, otherwise `false` or omit

### JSON Generation Rules

- Generate ONLY valid JSON
- No markdown code fences (no ```json)
- No explanations or commentary before/after the JSON
- Just pure JSON starting with `{` and ending with `}`
- All string fields must use double quotes
- Use proper JSON escaping for special characters

### Example JSON Output

```json
{
  "type": "feat",
  "scope": "api",
  "subject": "add user authentication endpoint",
  "body": "Implements JWT-based authentication for API access. This commit adds a new authentication endpoint that validates user credentials and returns JWT tokens. Includes rate limiting and input validation for security.",
  "breaking": false
}
```

## Output Format (for your reference)

The JSON will be automatically converted to conventional commit format:

```text
type(scope): subject

Auditor-Summary: [First sentence from body]

[Full body text]

Changes: [From git diff stats]
```

Example final output:
```text
feat(api): add user authentication endpoint

Auditor-Summary: Implements JWT-based authentication for API access.

Implements JWT-based authentication for API access. This commit adds
a new authentication endpoint that validates user credentials and
returns JWT tokens. Includes rate limiting and input validation for
security.

Changes: 12 files, +450 insertions, -20 deletions
```

## CRITICAL: What NOT to include

The following are FORBIDDEN in your JSON output:

- NO file paths or file lists in the JSON
- NO markdown formatting within string values
- NO code blocks or examples in the JSON
- NO nested objects beyond the schema shown

Generate JSON now based on the context below:
