# Output Formats

{{ page_breadcrumb() }}

## Overview

EAC commands produce output in different formats optimized for their use case. Understanding these formats helps you choose the right command and process its output effectively.

## Format Categories

### 1. JSON Output (Machine-Readable)

**Used by**: `get` commands

**Characteristics**:

- Structured, parseable data
- Always valid JSON
- Designed for automation
- Cacheable and transformable
- Language-agnostic

**Example**:

```bash
$ r2r eac get modules
{
  "modules": [
    {
      "moniker": "eac-commands",
      "type": "go-commands",
      "path": "go/eac/commands",
      "dependencies": ["eac-core"],
      "files": 45
    },
    {
      "moniker": "eac-core",
      "type": "go-library",
      "path": "go/eac/core",
      "dependencies": [],
      "files": 32
    }
  ]
}
```

**When to use**:

- CI/CD pipelines
- Build scripts
- Integration with other tools
- Data processing
- Caching results

**Processing with jq**:

```bash
# Extract specific fields
r2r eac get modules | jq -r '.modules[].moniker'

# Filter by type
r2r eac get modules | jq '.modules[] | select(.type == "go-library")'

# Count results
r2r eac get modules | jq '.modules | length'

# Transform structure
r2r eac get modules | jq '{names: [.modules[].moniker]}'
```

---

### 2. Formatted Tables (Human-Readable)

**Used by**: `show` commands

**Characteristics**:

- Aligned columns
- Headers and borders
- Colorized (when supported)
- Sortable and filterable (internally)
- Easy to scan visually

**Example**:

```bash
$ r2r eac show modules
┌───────────────┬─────────────┬────────────────────┬──────┐
│ Moniker       │ Type        │ Path               │ Files│
├───────────────┼─────────────┼────────────────────┼──────┤
│ eac-commands  │ go-commands │ go/eac/commands    │   45 │
│ eac-core      │ go-library  │ go/eac/core        │   32 │
│ src-auth      │ go-library  │ go/src/auth        │   18 │
└───────────────┴─────────────┴────────────────────┴──────┘
```

**When to use**:

- Interactive terminal use
- Quick exploration
- Status checking
- Reading output directly

**Features**:

- **Alignment**: Columns auto-align for readability
- **Colors**: Status indicators (green ✓, red ✗)
- **Sorting**: Pre-sorted by relevance
- **Truncation**: Long values truncated with ellipsis

---

### 3. Plain Text (Unstructured)

**Used by**: Various commands for messages and logs

**Characteristics**:

- Free-form text
- Human-readable
- May include formatting (bold, colors)
- Not easily parseable

**Example**:

```bash
$ r2r eac build src-auth
Building module: src-auth
✓ Dependencies resolved
✓ Compilation successful
✓ Tests passed
✓ Artifacts generated

Build completed in 12.3s
```

**When to use**:

- Progress messages
- Status updates
- Error messages
- User feedback

---

### 4. Reports (Formatted Documents)

**Used by**: `show *-summary` commands

**Characteristics**:

- Multi-section formatted output
- Markdown-compatible
- GitHub Actions friendly
- Includes summaries and details

**Example**:

```bash
$ r2r eac show test-summary src-auth acceptance
# Test Summary: src-auth (acceptance)

## Results
✓ Passed: 45
✗ Failed: 2
⊘ Skipped: 1
Total: 48

## Failed Tests
- Login with invalid password (@auth)
- Session timeout handling (@session)

## Performance
- Fastest: login_success (0.12s)
- Slowest: oauth_flow (2.34s)
- Average: 0.45s
```

**When to use**:

- CI/CD summaries
- Test reports
- Build reports
- Documentation generation

---

## Format Comparison

### get vs show Output

Most information commands come in pairs: `get` (JSON) and `show` (formatted).

#### Example: Module Information

**get modules** (JSON):

```bash
$ r2r eac get modules | jq '.modules[] | {moniker, type}'
{
  "moniker": "eac-commands",
  "type": "go-commands"
}
{
  "moniker": "eac-core",
  "type": "go-library"
}
```

**show modules** (Table):

```bash
$ r2r eac show modules
┌───────────────┬─────────────┬────────────────────┬──────┐
│ Moniker       │ Type        │ Path               │ Files│
├───────────────┼─────────────┼────────────────────┼──────┤
│ eac-commands  │ go-commands │ go/eac/commands    │   45 │
│ eac-core      │ go-library  │ go/eac/core        │   32 │
└───────────────┴─────────────┴────────────────────┴──────┘
```

**Same data, different formats** for different use cases.

---

## JSON Schema Patterns

### Standard Wrapper

Most JSON commands wrap results in a top-level object:

```json
{
  "modules": [...],
  "dependencies": {...},
  "files": [...]
}
```

### Common Fields

#### Metadata

```json
{
  "timestamp": "2024-12-09T14:30:00Z",
  "version": "1.0.0",
  "repository": "/path/to/repo"
}
```

#### Counts

```json
{
  "total": 10,
  "filtered": 8,
  "results": [...]
}
```

#### Status

```json
{
  "status": "success",
  "message": "Operation completed",
  "data": {...}
}
```

### Examples by Command

#### get modules

```json
{
  "modules": [
    {
      "moniker": "string",
      "type": "string",
      "path": "string",
      "dependencies": ["string"],
      "files": number
    }
  ]
}
```

#### get dependencies

```json
{
  "dependencies": {
    "module-name": ["dependency1", "dependency2"]
  }
}
```

#### get execution order

```json
{
  "execution_order": ["module1", "module2", "module3"]
}
```

#### get changed-modules

```json
{
  "changed_modules": ["module1", "module2"]
}
```

#### get config

```json
{
  "repository": {
    "root": "string",
    "name": "string"
  },
  "ai": {
    "provider": "string",
    "model": "string"
  },
  "build": {
    "parallel": boolean
  }
}
```

---

## Processing JSON Output

### Using jq

jq is the standard tool for processing JSON:

#### Extract Fields

```bash
# Get all module names
r2r eac get modules | jq -r '.modules[].moniker'

# Get specific field
r2r eac get config | jq -r '.ai.provider'
```

#### Filter Results

```bash
# Filter by type
r2r eac get modules | jq '.modules[] | select(.type == "go-library")'

# Filter by condition
r2r eac get modules | jq '.modules[] | select(.files > 30)'
```

#### Transform Structure

```bash
# Create array of names
r2r eac get modules | jq '[.modules[].moniker]'

# Create lookup table
r2r eac get dependencies | jq 'to_entries | map({(.key): .value})'
```

#### Combine Commands

```bash
# Get modules and count by type
r2r eac get modules | jq '.modules | group_by(.type) | map({type: .[0].type, count: length})'
```

### Using grep

For simple text extraction:

```bash
# Extract monikers (fragile, prefer jq)
r2r eac get modules | grep -Po '"moniker":\s*"\K[^"]*'

# Count occurrences
r2r eac get modules | grep -c '"moniker"'
```

### Using Python

For complex processing:

```python
import json
import subprocess

# Get module data
result = subprocess.run(
    ['r2r', 'eac', 'get', 'modules'],
    capture_output=True,
    text=True
)

data = json.loads(result.stdout)

# Process
for module in data['modules']:
    print(f"{module['moniker']}: {module['files']} files")
```

---

## Table Output Features

### Alignment

Tables auto-align columns based on content:

```text
┌─────────────┬───────┐     # Left-aligned text
│ Module      │ Files │     # Right-aligned numbers
├─────────────┼───────┤
│ eac-core    │    32 │
│ eac-commands│    45 │
└─────────────┴───────┘
```

### Truncation

Long values are truncated with ellipsis:

```text
┌──────────────────────────────────┐
│ Path                              │
├──────────────────────────────────┤
│ go/eac/commands/impl/very-lo...  │
└──────────────────────────────────┘
```

### Colors and Symbols

Status indicators use colors and symbols:

```text
Status          Symbol  Color
──────────────────────────────
Success         ✓       Green
Failure         ✗       Red
Warning         ⚠       Yellow
In Progress     ⋯       Blue
Skipped         ⊘       Gray
```

### Sorting

Tables are pre-sorted by relevance:

- Alphabetically (for lists)
- By dependency order (for modules)
- By timestamp (for logs)
- By importance (for issues)

---

## Report Formats

### Test Summary Reports

Format: Markdown-compatible

```markdown
# Test Summary: module-name (suite-name)

## Results
✓ Passed: 45
✗ Failed: 2
⊘ Skipped: 1
──────────────
Total: 48 (95.8% pass rate)

## Failed Tests
1. test_auth_invalid_password
   - Expected: 401 Unauthorized
   - Got: 200 OK
   - File: tests/auth_test.go:42

2. test_session_timeout
   - Error: Connection timeout
   - File: tests/session_test.go:78

## Performance
- Fastest: test_login_success (0.12s)
- Slowest: test_oauth_flow (2.34s)
- Average: 0.45s
- Total time: 21.6s

## Coverage
- Line coverage: 87.3%
- Branch coverage: 82.1%
```

### Build Summary Reports

Format: Markdown-compatible

```markdown
# Build Summary: module-name

## Status
✓ Build successful

## Timing
- Dependencies: 2.3s
- Compilation: 8.1s
- Tests: 12.3s
- Artifacts: 1.2s
──────────────
Total: 23.9s

## Artifacts Generated
- bin/eac-commands (15.2 MB)
- bin/eac-commands.sha256

## Dependencies
Built in order:
1. eac-core (5.1s)
2. eac-commands (18.8s)
```

---

## Output Redirection

### Saving JSON

```bash
# Save to file
r2r eac get modules > modules.json

# Append to file
r2r eac get modules >> all-data.json

# Save with errors
r2r eac get modules > modules.json 2>&1
```

### Saving Tables

```bash
# Save table output
r2r eac show modules > modules.txt

# Save without colors (for logs)
r2r eac show modules 2>&1 | tee modules.log
```

### Piping

```bash
# Pipe to jq
r2r eac get modules | jq '.modules[].moniker'

# Pipe to grep
r2r eac show modules | grep "go-library"

# Pipe to less
r2r eac show modules | less

# Pipe to wc
r2r eac get modules | jq '.modules | length'
```

---

## Error Output

### Error Format

Errors are written to **stderr** with this format:

```text
Error: <error-message>

<details>
<context>
<suggestion>
```

**Example**:

```bash
$ r2r eac build invalid-module
Error: Module 'invalid-module' not found

Available modules:
  - eac-commands
  - eac-core
  - src-auth

Suggestion: Check module name with: r2r eac show modules
```

### Error Categories

#### Validation Errors

```text
Error: Validation failed

go/eac/commands/impl/work.go:
  Line 42: undefined variable 'foo'
  Line 58: missing return statement

Run: r2r eac validate contracts
```

#### Configuration Errors

```text
Error: AI provider not configured

No configuration found at: .r2r/eac/eac-config.yml

Run: r2r eac init
```

#### Build Errors

```text
Error: Build failed for module 'eac-commands'

Compilation errors:
  go/eac/commands/main.go:15: undefined: fmt.Printl

Fix errors and run: r2r eac build eac-commands
```

---

## Best Practices

### Choosing Output Format

1. **Use `get` for automation**:

   ```bash
   # CI/CD script
   CHANGED=$(r2r eac get changed-modules | jq -r '.changed_modules[]')
   ```

2. **Use `show` for exploration**:

   ```bash
   # Interactive use
   r2r eac show modules
   r2r eac show dependencies
   ```

3. **Pipe JSON through jq**:

   ```bash
   # Always validate and format JSON
   r2r eac get modules | jq '.'
   ```

4. **Save JSON for caching**:

   ```bash
   # Cache expensive operations
   r2r eac get files > files.json
   jq '.files[] | select(.module == "src-auth")' files.json
   ```

### Processing Output

1. **Check exit codes first**:

   ```bash
   if r2r eac get modules > modules.json; then
     jq '.modules[].moniker' modules.json
   else
     echo "Failed to get modules"
     exit 1
   fi
   ```

2. **Handle empty results**:

   ```bash
   MODULES=$(r2r eac get modules | jq -r '.modules[]')
   if [ -z "$MODULES" ]; then
     echo "No modules found"
     exit 1
   fi
   ```

3. **Quote JSON values**:

   ```bash
   # Always use -r flag and quote
   MODULE=$(r2r eac get modules | jq -r '.modules[0].moniker')
   echo "Building $MODULE"
   ```

---

## See Also

- [Command Taxonomy](./command-taxonomy.md) - How commands are organized
- [Naming Conventions](./naming-conventions.md) - Command naming patterns
- [Common Flags](./common-flags.md) - Global options
- [Get Commands Category](../categories/get.md) - All JSON commands
- [Show Commands Category](../categories/show.md) - All formatted commands

{{ diataxis_footer() }}
