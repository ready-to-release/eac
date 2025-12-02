# Commit Command

**Problem**: Writing high-quality, semantic commit messages is time-consuming and requires consistency across team members, especially for complex multi-module changes.

**Solution**: Use `commit` to generate AI-powered commit messages that analyze your staged changes and follow project conventions automatically.

## Key Benefits

- Multi-phase AI generation with validation
- Parallel processing for multi-module commits
- Automatic contract validation
- Standard commit types (feat, fix, refactor, docs, chore, test, perf, style)
- Retry logic with auto-cleanup of formatting issues
- Context-aware module analysis

## Quick Start

```bash
# Stage your changes
git add src/auth/login.go src/auth/login_test.go

# Generate commit message
r2r eac create commit-message

# Review and create commit automatically
r2r eac create commit-message --commit
```

## Command Reference

### create commit-message

Generate AI-powered commit messages from staged changes.

```bash
r2r eac create commit-message [options]

# Options:
--commit, -c           # Automatically create the commit (no manual edit)
--debug, -d            # Save intermediate outputs to out/ directory

# Examples:
r2r eac create commit-message                    # Generate message, open editor
r2r eac create commit-message --commit           # Generate and commit automatically
r2r eac create commit-message --debug            # Debug AI generation process
r2r eac create commit-message --commit --debug   # Auto-commit with debug output
```

**What it does:**

1. **Context Analysis**: Analyzes git status, diff, and recent commits
2. **Summary Generation**: Creates concise change summary
3. **Module Sections**: Processes each affected module in parallel
4. **Assembly**: Combines sections into structured commit message
5. **Validation**: Validates against project contracts
6. **Auto-cleanup**: Fixes common formatting issues (up to 5 retries)
7. **Commit**: Opens editor or commits directly with `--commit`

**Generated format:**

```text
<type>(<module>): <short description>

<detailed description of changes>

<bullet points for key changes>

Files modified:
- path/to/file1.go
- path/to/file2.go
```

### Soft Reset (git native)

To undo the last commit while preserving all changes, use git directly:

```bash
git reset --soft HEAD~1

# What happens:
# 1. Undoes the last commit
# 2. Preserves all changes in staging area
# 3. Allows you to recommit with a new message
```

**Use when:**

- You want to improve the commit message
- You need to add more changes to the last commit
- You committed too early and want to stage additional files

## Commit Types

The AI uses standard semantic commit types:

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New feature or capability | `feat(src-auth): implement JWT authentication` |
| `fix` | Bug fix | `fix(src-api): resolve rate limiting race condition` |
| `refactor` | Code restructuring (no behavior change) | `refactor(eac-core): extract validation logic` |
| `docs` | Documentation only | `docs(specs): add authentication examples` |
| `chore` | Maintenance tasks | `chore(deps): update dependencies` |
| `test` | Test changes only | `test(src-auth): add integration tests` |
| `perf` | Performance improvements | `perf(src-cache): optimize lookup algorithm` |
| `style` | Code formatting (no logic change) | `style(eac-core): apply gofmt` |

## AI Generation Process

### Phase 1: Context Collection

```text
Collecting context...
- Git status (staged, unstaged, untracked files)
- Git diff (detailed changes)
- Recent commits (for style consistency)
- Module ownership (which modules are affected)
```

### Phase 2: Summary Generation

```text
Generating summary...
- Analyzes overall change purpose
- Identifies change type (feat, fix, refactor, etc.)
- Determines primary affected module
```

### Phase 3: Module Processing (Parallel)

```text
Processing modules in parallel...
- src-auth: Analyzing authentication changes...
- eac-core: Analyzing core library changes...
- src-tests: Analyzing test changes...
```

### Phase 4: Assembly

```text
Assembling commit message...
- Combines module sections
- Structures with headers and bullets
- Adds file list
- Formats according to conventions
```

### Phase 5: Validation

```text
Validating commit message...
✓ Follows semantic commit format
✓ Module exists in contracts
✓ Proper structure and formatting
✓ No forbidden patterns
```

## Example Outputs

### Single Module Feature

```text
feat(src-auth): implement JWT authentication system

Add secure authentication using JWT tokens for API access. The system
includes token generation, validation, and refresh mechanisms with
configurable expiration times.

Key changes:
- Token generation with HS256 signing
- Middleware for request authentication
- Token refresh endpoint
- Comprehensive test coverage

Files modified:
- src/auth/jwt.go
- src/auth/jwt_test.go
- src/auth/middleware.go
```

### Multi-Module Refactoring

```text
refactor(multi-module): reorganize validation logic

Extract validation logic from individual modules into a shared validation
package to reduce duplication and improve consistency across the codebase.

eac-core changes:
- Create new validation package with common validators
- Add field validation helpers
- Implement error aggregation

src-auth changes:
- Replace local validators with shared validators
- Update tests to use new validation package

src-api changes:
- Migrate request validation to shared validators
- Remove duplicate validation code

Files modified:
- go/eac/core/validation/validators.go (new)
- go/eac/core/validation/validators_test.go (new)
- src/auth/validators.go (removed)
- src/auth/login.go
- src/api/handlers.go
```

### Bug Fix

```text
fix(src-cache): resolve race condition in cache invalidation

Fix concurrent map access race condition that occurred during cache
invalidation under high load. Add proper mutex locking to protect
shared state.

Changes:
- Add sync.RWMutex to cache struct
- Lock during invalidation operations
- Add race detector tests

Files modified:
- src/cache/cache.go
- src/cache/cache_test.go
```

## Debug Mode

Use `--debug` to inspect the AI generation process:

```bash
r2r eac create commit-message --debug
```

Creates debug files in `out/`:

```text
out/
├── commit-message/
│   ├── 01-context.md              # Git status, diff, recent commits
│   ├── 02-summary-prompt.md       # AI prompt for summary
│   ├── 03-summary-response.md     # AI-generated summary
│   ├── 04-module-prompts/
│   │   ├── src-auth.md            # Prompt for src-auth module
│   │   └── eac-core.md            # Prompt for eac-core module
│   ├── 05-module-responses/
│   │   ├── src-auth.md            # AI response for src-auth
│   │   └── eac-core.md            # AI response for eac-core
│   ├── 06-assembly-prompt.md      # Prompt to combine sections
│   ├── 07-assembly-response.md    # Final assembled message
│   ├── 08-validation-result.json  # Validation details
│   └── 09-final-message.txt       # Cleaned commit message
```

**Use debug mode when:**

- AI generates unexpected messages
- Validation fails repeatedly
- Understanding module analysis
- Troubleshooting formatting issues
- Customizing AI prompts

## Typical Workflows

### Standard Workflow

```bash
# 1. Make changes
vim src/auth/login.go

# 2. Stage changes
git add src/auth/

# 3. Generate commit message
r2r eac create commit-message

# 4. Review in editor, save, and commit
```

### Fast Workflow (Auto-commit)

```bash
# 1. Make changes and stage
git add .

# 2. Generate and commit in one step
r2r eac create commit-message --commit

# No editor opens - message is auto-committed
```

### Iterative Workflow

```bash
# 1. Generate initial commit
r2r eac create commit-message --commit

# 2. Realize you need to add more changes
# commit reset was removed - use git reset --soft HEAD~1

# 3. Stage additional changes
git add src/auth/logout.go

# 4. Generate new commit with all changes
r2r eac create commit-message --commit
```

### Multi-Module Development

```bash
# Work on multiple modules
vim src/auth/jwt.go
vim go/eac/core/validation.go
vim src/api/handlers.go

# Stage all changes
git add go/eac/auth/ go/eac/core/ go/eac/api/

# AI processes modules in parallel
r2r eac create commit-message --debug

# Review parallel processing in out/commit-message/
```

## Integration with Work Command

### Standard Work Integration

```bash
# Create workspace
r2r eac work create feature/authentication

# Develop feature
vim src/auth/jwt.go

# Use commit command directly
git add src/auth/
r2r eac create commit-message --commit

# Or use work commit (which calls commit message internally)
r2r eac work commit --all
```

### Difference: commit vs work commit

| Command | Use Case | Working Directory |
|---------|----------|-------------------|
| `r2r eac create commit-message` | Direct commit message generation | Any git repo |
| `r2r eac work commit` | Workspace-aware commits with auto-staging | Inside work workspace |

Both use the same AI engine and validation.

## Best Practices

### Staging Strategy

```bash
# ✅ Good: Stage related changes together
git add src/auth/jwt.go src/auth/jwt_test.go
r2r eac create commit-message --commit

# ❌ Avoid: Mixing unrelated changes
git add src/auth/jwt.go src/api/unrelated.go src/docs/random.md
r2r eac create commit-message --commit
```

### Commit Frequency

```bash
# ✅ Good: Frequent, focused commits
git add src/auth/jwt.go
r2r eac create commit-message --commit

git add src/auth/middleware.go
r2r eac create commit-message --commit

# ❌ Avoid: Massive commits with many unrelated changes
git add .
r2r eac create commit-message --commit
```

### Message Quality

The AI generates better messages when you:

- Stage related changes together
- Make focused changes to specific modules
- Include tests with implementation
- Have clear, descriptive file names
- Follow consistent coding patterns

## Validation and Retry Logic

### Automatic Retry

If validation fails, the AI automatically retries (up to 5 attempts):

```text
Attempt 1: ✗ Invalid format - missing module in parentheses
Attempt 2: ✗ Invalid format - incorrect bullet structure
Attempt 3: ✓ Validation passed
```

### Common Auto-fixes

- Missing module identifier
- Incorrect bullet formatting
- Extra whitespace
- Invalid commit type
- Missing file list

### Manual Intervention

If all retries fail:

```text
Error: Failed to generate valid commit message after 5 attempts.

Please check:
1. Staged changes are valid
2. Module contracts are up to date
3. Debug output in out/commit-message/

Try running with --debug for detailed diagnostics.
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| No staged changes | Run `git add <files>` first |
| AI API error | Check API key configuration with `r2r eac init` |
| Invalid module | Verify module exists in contracts: `r2r eac show modules` |
| Validation fails | Use `--debug` to inspect outputs, check contract rules |
| Message too generic | Stage more focused changes, smaller commits |
| Wrong commit type | AI analyzes git diff - ensure changes match intent |
| Parallel processing slow | Normal for multi-module changes; AI processes in parallel |
| Reset fails | Ensure you have at least one commit: `git log` |

## Advanced Usage

### Custom Editor

```bash
# Use specific editor for commit message
export GIT_EDITOR="vim"
r2r eac create commit-message

export GIT_EDITOR="code --wait"
r2r eac create commit-message
```

### Pre-commit Hook Integration

```bash
# .git/hooks/prepare-commit-msg
#!/bin/bash

# Only generate message if commit message is empty
if [ -z "$2" ]; then
  r2r eac create commit-message --commit
fi
```

### CI/CD Validation

```bash
# Validate commit message format in CI
git log -1 --pretty=%B > last-commit.txt

# Parse and validate (implement custom validator)
if ! validate-commit-format last-commit.txt; then
  echo "Invalid commit message format"
  exit 1
fi
```

### Batch Commits

```bash
# Commit multiple logical changes separately
for dir in go/eac/auth go/eac/api go/eac/core; do
  if git diff --cached --name-only | grep -q "^$dir/"; then
    git reset  # Unstage all
    git add $dir/
    r2r eac create commit-message --commit
  fi
done
```

## Configuration

### AI Provider Setup

```bash
# Configure AI provider (first time)
r2r eac init

# Select provider: openai, anthropic, azure, etc.
# Enter API key when prompted
```

### Custom Prompts

Customize AI behavior by editing prompts in `.r2r/eac/ai/commit/`:

```text
.r2r/eac/ai/commit/
├── context-prompt.md       # How to analyze git context
├── summary-prompt.md       # How to generate summary
├── module-prompt.md        # How to analyze each module
└── assembly-prompt.md      # How to combine sections
```

After editing, generate commits normally - custom prompts are used automatically.

## Performance

### Parallel Processing

Multi-module commits process modules in parallel:

```text
Processing 3 modules...
[src-auth] Started
[src-api]  Started
[eac-core] Started
[src-auth] Complete (2.3s)
[eac-core] Complete (2.5s)
[src-api]  Complete (2.8s)
Total: 2.8s (vs 7.6s sequential)
```

### Caching

Context analysis results are cached during retry attempts to avoid re-processing.

## Examples from Real Usage

### Documentation Update

```text
docs(specs): add authentication specification examples

Add comprehensive examples for authentication specifications including
JWT, OAuth2, and session-based authentication patterns.

Changes:
- Add JWT authentication example spec
- Add OAuth2 flow specification
- Add session management examples
- Update specification guidelines

Files modified:
- specs/src-auth/jwt-authentication.feature (new)
- specs/src-auth/oauth2-flow.feature (new)
- docs/specifications.md
```

### Performance Optimization

```text
perf(src-cache): optimize cache lookup with bloom filter

Reduce cache miss overhead by adding bloom filter for quick existence
checks before expensive cache operations. Improves performance by 40%
for cache-miss scenarios.

Implementation:
- Add bloom filter with configurable false positive rate
- Check bloom filter before cache lookup
- Benchmark shows 40% improvement for misses
- No impact on cache-hit performance

Files modified:
- src/cache/cache.go
- src/cache/bloom.go (new)
- src/cache/cache_test.go
- src/cache/benchmark_test.go
```

### Breaking Change

```text
refactor(src-api): change response format to JSON:API spec

BREAKING CHANGE: Update all API responses to follow JSON:API specification.
Clients must update to handle new response structure.

Changes:
- Implement JSON:API response formatter
- Update all handlers to use new format
- Add migration guide for clients
- Update API documentation

Files modified:
- src/api/response.go
- src/api/handlers.go
- src/api/handlers_test.go
- docs/api-migration.md (new)
```

## Summary

1. **Stage changes**: `git add <files>`
2. **Generate message**: `r2r eac create commit-message`
3. **Auto-commit**: Add `--commit` flag
4. **Debug issues**: Add `--debug` flag
5. **Reset if needed**: `# commit reset was removed - use git reset --soft HEAD~1`

The commit command leverages AI to analyze your changes across multiple modules in parallel, generate semantic commit messages following project conventions, and validate against contracts - all while providing full transparency through debug mode.
