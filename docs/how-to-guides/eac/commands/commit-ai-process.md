# AI Commit Message Generation Process

{{ page_breadcrumb() }}

This document explains how the AI-powered commit message generation works under the hood.

For practical usage, see the [How-to Guide](./commit-command.md).

For technical details, see the [Command Reference](../../../reference/commands/commit-reference.md).

## Overview

The commit command uses a multi-phase AI approach to generate high-quality, semantic commit messages:

1. **Context Collection**: Gather all relevant information about the changes
2. **Summary Generation**: Analyze overall purpose and identify commit type
3. **Module Processing**: Process each affected module in parallel
4. **Assembly**: Combine all sections into structured message
5. **Validation**: Validate against contracts and retry if needed

This approach ensures:

- Accurate understanding of changes across multiple modules
- Consistent formatting following project conventions
- Semantic commit types (feat, fix, refactor, etc.)
- Proper module attribution
- Contract compliance

## Five-Phase Generation Process

### Phase 1: Context Collection

**What happens:**

The system collects comprehensive context about your changes:

```text
Collecting context...
- Git status (staged, unstaged, untracked files)
- Git diff (detailed changes)
- Recent commits (for style consistency)
- Module ownership (which modules are affected)
```

**Why this matters:**

- **Git status**: Identifies which files are changed and their state
- **Git diff**: Provides line-by-line change details for analysis
- **Recent commits**: Ensures consistency with existing commit message style
- **Module ownership**: Maps changed files to their owning modules

**Example context:**

```text
Staged files:
- src/auth/jwt.go (modified)
- src/auth/jwt_test.go (new)
- src/auth/middleware.go (modified)

Diff summary:
- Added JWT token generation (150 lines)
- Added token validation middleware (80 lines)
- Added comprehensive tests (200 lines)

Recent commit style:
feat(src-auth): implement OAuth2 flow
fix(src-api): resolve timeout handling

Module mapping:
- src/auth/* → src-auth module
```

### Phase 2: Summary Generation

**What happens:**

The AI analyzes the collected context to understand the overall change:

```text
Generating summary...
- Analyzes overall change purpose
- Identifies change type (feat, fix, refactor, etc.)
- Determines primary affected module
```

**How it works:**

The AI receives a prompt containing:

1. Git diff with all changes
2. List of affected files
3. Recent commit examples for style
4. Instructions to identify:
   - Commit type (feat, fix, refactor, docs, chore, test, perf, style)
   - Primary module affected
   - High-level description

**Example analysis:**

```text
Input: Changes to src/auth/jwt.go, src/auth/jwt_test.go, src/auth/middleware.go
Analysis:
- Type: feat (new functionality, not fixing existing)
- Module: src-auth (all changes in auth module)
- Purpose: Implement JWT authentication system
```

**Decision logic:**

| Change Pattern | Commit Type |
|----------------|-------------|
| New files, new functions, new features | `feat` |
| Fixing broken behavior | `fix` |
| Restructuring without behavior change | `refactor` |
| Only documentation changes | `docs` |
| Only test changes | `test` |
| Performance improvements | `perf` |
| Formatting only | `style` |
| Dependencies, tooling | `chore` |

### Phase 3: Module Processing (Parallel)

**What happens:**

Each affected module is analyzed in parallel:

```text
Processing modules in parallel...
- src-auth: Analyzing authentication changes...
- eac-core: Analyzing core library changes...
- src-tests: Analyzing test changes...
```

**How parallel processing works:**

For a multi-module change:

```text
Changes detected:
- go/eac/auth/jwt.go
- go/eac/core/validation.go
- go/eac/api/handlers.go

Parallel analysis:
Thread 1: Analyze go/eac/auth/* → Generate auth section
Thread 2: Analyze go/eac/core/* → Generate core section
Thread 3: Analyze go/eac/api/* → Generate api section

Wait for all threads to complete...
```

**Per-module analysis:**

Each module receives a focused prompt:

```text
Analyze changes in the src-auth module:

Files changed:
- src/auth/jwt.go: +150 lines, -0 lines
- src/auth/jwt_test.go: +200 lines (new file)

Diff:
[module-specific diff only]

Task: Describe what changed in this module specifically
```

**Example module responses:**

**src-auth module:**
```text
Implement JWT authentication system with token generation, validation,
and refresh mechanisms. Add comprehensive test coverage including unit
and integration tests.

Key changes:
- Token generation with HS256 signing
- Middleware for request authentication
- Token refresh endpoint
- Comprehensive test coverage
```

**eac-core module:**
```text
Add validation helpers for JWT token structure and claims validation.

Changes:
- Token validation utilities
- Claims verification helpers
```

### Phase 4: Assembly

**What happens:**

The AI combines all sections into a coherent, structured commit message:

```text
Assembling commit message...
- Combines module sections
- Structures with headers and bullets
- Adds file list
- Formats according to conventions
```

**Assembly process:**

The AI receives:

1. Overall summary from Phase 2
2. All module-specific sections from Phase 3
3. Complete file list
4. Formatting instructions

**Example assembly:**

```text
Inputs:
- Summary: "feat(multi-module): implement JWT authentication"
- Module sections:
  - src-auth: [detailed description]
  - eac-core: [detailed description]
- Files: jwt.go, jwt_test.go, validation.go

Assembly prompt:
"Combine these sections into a structured commit message following
the format: header, description, module sections, file list"

Output:
feat(multi-module): implement JWT authentication system

Add JWT authentication across authentication and core modules...

src-auth changes:
- Token generation with HS256 signing
...

eac-core changes:
- Token validation utilities
...

Files modified:
- src/auth/jwt.go
- src/auth/jwt_test.go
- go/eac/core/validation.go
```

### Phase 5: Validation

**What happens:**

The commit message is validated against project contracts:

```text
Validating commit message...
✓ Follows semantic commit format
✓ Module exists in contracts
✓ Proper structure and formatting
✓ No forbidden patterns
```

**Validation checks:**

1. **Format validation**
   ```text
   Pattern: ^(feat|fix|refactor|docs|chore|test|perf|style)\([^)]+\): .+
   ✓ feat(src-auth): implement JWT authentication
   ✗ added jwt auth (invalid format)
   ```

2. **Module validation**
   ```text
   Check: Module exists in contracts or is "multi-module"
   ✓ src-auth (exists in contracts)
   ✓ multi-module (special case)
   ✗ src-invalid (not in contracts)
   ```

3. **Structure validation**
   ```text
   Required sections:
   ✓ Header line
   ✓ Description paragraph
   ✓ Detailed changes (bullets)
   ✓ Files modified list
   ```

4. **Content validation**
   ```text
   Forbidden patterns:
   ✗ Generic descriptions like "updated files"
   ✗ Missing file list
   ✗ Inconsistent formatting
   ```

**Retry logic:**

If validation fails, the AI retries up to 5 times with feedback:

```text
Attempt 1: Generate initial message
Validation: ✗ Missing module identifier
Feedback: "Add module in format: type(module): description"

Attempt 2: Generate with feedback
Validation: ✗ Incorrect bullet formatting
Feedback: "Use '-' for bullets, not '*'"

Attempt 3: Generate with all feedback
Validation: ✓ Success
```

**Auto-fixes between retries:**

- Add missing module identifier
- Fix bullet point formatting
- Correct commit type spelling
- Add missing file list
- Fix whitespace issues

## Context Analysis Details

### Git Status Parsing

The system parses `git status` to understand file states:

```text
git status --porcelain
M  src/auth/jwt.go          → Modified (staged)
A  src/auth/jwt_test.go     → Added (staged)
?? src/auth/temp.go         → Untracked (ignored)
```

**Why it matters:**

- Only staged files are included in analysis
- Untracked files are ignored
- File state (added, modified, deleted) influences description

### Git Diff Analysis

The system analyzes `git diff --cached` for detailed changes:

```diff
diff --git a/src/auth/jwt.go b/src/auth/jwt.go
index 0000000..1234567 100644
--- a/src/auth/jwt.go
+++ b/src/auth/jwt.go
@@ -1,0 +1,50 @@
+package auth
+
+func GenerateToken(userID string) (string, error) {
+    // Token generation logic
+}
```

**What the AI looks for:**

- **New functions**: Indicates new features
- **Modified functions**: Suggests fixes or refactoring
- **Deleted code**: Shows removal or cleanup
- **Test files**: Indicates quality commitment
- **Import changes**: May suggest dependency updates

### Recent Commit Analysis

The system reads recent commits for style consistency:

```bash
git log -10 --oneline
a1b2c3d feat(src-auth): implement OAuth2 flow
e4f5g6h fix(src-api): resolve timeout handling
i7j8k9l refactor(eac-core): extract validation logic
```

**Style elements extracted:**

- Commit type usage patterns
- Module naming conventions
- Description formatting
- Level of detail
- Bullet point style

## Module Processing Strategy

### Module Detection

The system maps files to modules using contract definitions:

```text
Contract: src-auth owns src/auth/**
File: src/auth/jwt.go → Module: src-auth

Contract: eac-core owns go/eac/core/**
File: go/eac/core/validation.go → Module: eac-core
```

### Single vs. Multi-Module

**Single module change:**

```text
Files: src/auth/jwt.go, src/auth/jwt_test.go
Modules: src-auth
Result: feat(src-auth): ...
```

**Multi-module change:**

```text
Files: src/auth/jwt.go, go/eac/core/validation.go, src/api/handlers.go
Modules: src-auth, eac-core, src-api
Result: feat(multi-module): ...

Body sections:
- src-auth changes: ...
- eac-core changes: ...
- src-api changes: ...
```

### Parallel Execution

**Why parallel processing?**

For multi-module commits, sequential processing would be slow:

```text
Sequential (slow):
1. Process src-auth: 3 seconds
2. Process eac-core: 3 seconds
3. Process src-api: 3 seconds
Total: 9 seconds

Parallel (fast):
1. Process all modules simultaneously
Total: 3 seconds (max of all modules)
```

**How it works:**

```text
Main thread:
├─ Spawn worker for src-auth module
├─ Spawn worker for eac-core module
├─ Spawn worker for src-api module
└─ Wait for all workers to complete

Each worker:
1. Receive module-specific diff
2. Call AI with module context
3. Return module description
```

## Validation and Retry Logic

### Why Validation Matters

AI responses can be inconsistent, so validation ensures:

- Messages follow project conventions
- Module names are valid
- Format is parseable by git tools
- Content is meaningful, not generic

### Validation Workflow

```text
1. Generate commit message
   ↓
2. Run validation checks
   ↓
3. Pass? → Success, use message
   ↓
4. Fail? → Collect error feedback
   ↓
5. Retry with feedback (max 5 attempts)
   ↓
6. All retries failed? → Error to user
```

### Feedback Loop

**Example validation cycle:**

```text
Attempt 1:
Generated: "implement jwt authentication"
Error: Missing commit type and module
Feedback: "Use format: type(module): description"

Attempt 2:
Generated: "feat: implement jwt authentication"
Error: Missing module identifier
Feedback: "Add module in parentheses: feat(module): description"

Attempt 3:
Generated: "feat(src-auth): implement jwt authentication"
Success: ✓
```

### Common Retry Patterns

| Issue | Auto-fix Strategy |
|-------|-------------------|
| Missing module | Extract from file paths and add to header |
| Wrong bullet format | Convert `*` or `•` to `-` |
| Invalid commit type | Suggest closest valid type (e.g., "feature" → "feat") |
| Missing file list | Generate from staged files |
| Extra whitespace | Trim and normalize spacing |

## Understanding the Output

### Reading Debug Files

When you use `--debug`, you can see exactly how the AI works:

**01-context.md** - What the AI sees:
```text
Status:
M  src/auth/jwt.go
A  src/auth/jwt_test.go

Diff:
+func GenerateToken(userID string) (string, error) {
+    // Implementation
+}

Recent commits:
feat(src-auth): implement OAuth2 flow
```

**02-summary-prompt.md** - What we ask:
```text
Analyze these changes and identify:
1. Commit type (feat, fix, refactor, etc.)
2. Primary module
3. Overall purpose
```

**03-summary-response.md** - What AI responds:
```text
Type: feat
Module: src-auth
Purpose: Implement JWT authentication system
```

**04-module-prompts/** - Module-specific questions:
```text
src-auth.md:
"Describe changes in src-auth module specifically"
```

**05-module-responses/** - Module-specific answers:
```text
src-auth.md:
"Implement JWT token generation with HS256 signing..."
```

**06-assembly-prompt.md** - How to combine:
```text
Combine these sections into a commit message:
- Header: feat(src-auth): implement JWT authentication
- Sections: [module descriptions]
- Files: [file list]
```

**07-assembly-response.md** - Final assembled message

**08-validation-result.json** - Validation details:
```json
{
  "valid": true,
  "errors": [],
  "warnings": [],
  "attempt": 1
}
```

**09-final-message.txt** - What gets committed

### Interpreting AI Decisions

**Why did the AI choose "feat" vs "fix"?**

```text
feat: New functionality added (new functions, new files)
fix: Broken behavior corrected (bug fixes)

Indicators for "feat":
- New files created
- New exported functions
- Adding capabilities

Indicators for "fix":
- Fixing broken tests
- Correcting incorrect behavior
- Addressing reported issues
```

**Why "multi-module" instead of single module?**

```text
Single module: All changes in one module
Multi-module: Changes span 2+ modules

Example single:
- src/auth/jwt.go
- src/auth/jwt_test.go
→ feat(src-auth): ...

Example multi:
- src/auth/jwt.go
- go/eac/core/validation.go
→ feat(multi-module): ...
```

**Why did parallel processing take X seconds?**

```text
Parallel time = max(module_1_time, module_2_time, module_3_time)

Example:
- Module 1: 2.5s
- Module 2: 3.2s ← slowest
- Module 3: 1.8s
Total: 3.2s (not 2.5 + 3.2 + 1.8 = 7.5s)
```

## Summary

The AI commit message generation uses a sophisticated 5-phase process:

1. **Context Collection**: Gather comprehensive information about changes
2. **Summary Generation**: Understand overall purpose and commit type
3. **Module Processing**: Analyze each module in parallel for detailed descriptions
4. **Assembly**: Combine all sections into structured message
5. **Validation**: Ensure quality with automatic retry and feedback

This approach provides:

- High-quality, semantic commit messages
- Consistent formatting across the team
- Proper module attribution
- Contract compliance
- Full transparency via debug mode

For practical usage, see the [How-to Guide](./commit-command.md).

For technical details, see the [Command Reference](../../../reference/commands/commit-reference.md).

{{ diataxis_footer() }}
