# Generate Module Section

Generate a module section for a multi-module commit message.

Your task is to generate structured JSON data. The command will automatically format it into a module section.

## JSON Output Structure

Generate a JSON object matching this schema:

```json
{
  "module": "module-name",
  "type": "refactor",
  "description": "brief description of change",
  "body": "2-4 sentences explaining what changed in this module"
}
```

### JSON Field Requirements

**module** (required): Module name (e.g., `eac-commands`, `contracts`)
- Lowercase with hyphens
- Length: 1-30 characters

**type** (required): One of: `feat`, `fix`, `refactor`, `docs`, `chore`, `test`, `perf`, `style`, `ci`, `build`

**description** (required): Brief description of changes in this module
- Max 72 characters when combined with `module: type: `
- No trailing period
- Lowercase first letter

**body** (required): 2-4 sentences explaining what changed
- Each line wrapped at 72 characters
- Provide specific details about this module's changes

### JSON Generation Rules

- Generate ONLY valid JSON
- No markdown code fences (no ```json)
- No explanations or commentary before/after the JSON
- Just pure JSON starting with `{` and ending with `}`
- All string fields must use double quotes
- Use proper JSON escaping for special characters (especially newlines: `\n`)

### Example JSON Output

```json
{
  "module": "eac-commands",
  "type": "refactor",
  "description": "simplify commit message generation",
  "body": "Removed template variable embedding from generation logic and updated\nprompts to use direct format instructions. Simplified assembly code to\nuse blank line separators instead of dashes between module sections."
}
```

## Output Format (for your reference)

The JSON will be automatically converted to module section format:

```text
module
------------
module: type: description

body
```

Example final output:
```text
eac-commands
------------
eac-commands: refactor: simplify commit message generation

Removed template variable embedding from generation logic and updated
prompts to use direct format instructions. Simplified assembly code to
use blank line separators instead of dashes between module sections.
```

## Output Rules

- Generate ONLY valid JSON
- No markdown headers or formatting within JSON strings
- No conversational text before or after JSON

Generate JSON now based on the context below:
