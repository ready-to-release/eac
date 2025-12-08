<!-- EDITOR
# Editor: reference/commands/commit-reference.md

## Soul

Complete technical reference for the commit command including all flags, commit types, validation rules, debug mode, pre-commit hooks, and advanced configuration.

## Sections

1. Command Syntax
2. Options and Flags
3. Commit Types
4. Generated Format
5. Debug Mode
6. Validation and Retry Logic
7. Integration with Work Command
8. Advanced Usage
9. Configuration
10. Performance
11. Example Outputs

## Related Files

- [How-to Guide](../../how-to-guides/commands/commit-command.md) - Quick start and common workflows
- [AI Process Explanation](../../explanation/commands/commit-ai-process.md) - How the AI generation works

-->

# Commit Command Reference

Complete technical reference for the `commit` command.

For practical usage and workflows, see the [How-to Guide](../../how-to-guides/commands/commit-command.md).

To understand the AI generation process, see [AI Process Explanation](../../explanation/commands/commit-ai-process.md).

## Command Syntax

```bash
r2r eac create commit-message [options]
```

## Options and Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--commit` | `-c` | Automatically create the commit without opening editor | `false` |
| `--debug` | `-d` | Save intermediate outputs to `out/` directory for inspection | `false` |

### Examples

```bash
# Generate message and open editor
r2r eac create commit-message

# Generate and commit automatically
r2r eac create commit-message --commit

# Debug AI generation process
r2r eac create commit-message --debug

# Auto-commit with debug output
r2r eac create commit-message --commit --debug
```

## Commit Types

The AI uses standard semantic commit types:

| Type | Description | Use Case | Example |
|------|-------------|----------|---------|
| `feat` | New feature or capability | Adding new functionality | `feat(src-auth): implement JWT authentication` |
| `fix` | Bug fix | Fixing broken behavior | `fix(src-api): resolve rate limiting race condition` |
| `refactor` | Code restructuring (no behavior change) | Improving code structure | `refactor(eac-core): extract validation logic` |
| `docs` | Documentation only | Documentation changes | `docs(specs): add authentication examples` |
| `chore` | Maintenance tasks | Dependency updates, tooling | `chore(deps): update dependencies` |
| `test` | Test changes only | Adding or modifying tests | `test(src-auth): add integration tests` |
| `perf` | Performance improvements | Optimizing performance | `perf(src-cache): optimize lookup algorithm` |
| `style` | Code formatting (no logic change) | Formatting, whitespace | `style(eac-core): apply gofmt` |

### Multi-Module Commits

For changes affecting multiple modules, use `multi-module` as the scope:

```text
refactor(multi-module): reorganize validation logic
```

## Generated Format

### Standard Commit Structure

```text
<type>(<module>): <short description>

<detailed description of changes>

<bullet points for key changes>

Files modified:
- path/to/file1.go
- path/to/file2.go
```

### Single Module Example

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

### Multi-Module Example

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

## Debug Mode

Enable debug mode with the `--debug` flag to inspect the AI generation process:

```bash
r2r eac create commit-message --debug
```

### Debug Output Structure

Debug files are saved to `out/commit-message/`:

```text
out/
├── commit-message/
│   ├── 01-context.md              # Git status, diff, recent commits
│   ├── 02-summary-prompt.md       # AI prompt for summary generation
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

### Debug File Contents

#### 01-context.md
Contains the raw git context:
- `git status` output
- `git diff --cached` output
- Recent commit history
- Module ownership information

#### 02-summary-prompt.md
The prompt sent to the AI to generate the overall summary, including:
- Context about the changes
- Instructions for summary format
- Examples of good summaries

#### 03-summary-response.md
The AI's response containing:
- Identified commit type
- Primary affected module
- Overall change description

#### 04-module-prompts/
Individual prompts for each affected module, containing:
- Module-specific changes
- Module context
- Instructions for detailed analysis

#### 05-module-responses/
AI-generated detailed analysis for each module

#### 06-assembly-prompt.md
Prompt to combine all sections into final commit message

#### 07-assembly-response.md
The assembled commit message before validation

#### 08-validation-result.json
Validation results including:
- Pass/fail status
- Validation errors
- Retry attempts
- Contract checks

#### 09-final-message.txt
The final commit message after all cleanup and validation

### When to Use Debug Mode

Use `--debug` when:

- AI generates unexpected messages
- Validation fails repeatedly
- Understanding module analysis
- Troubleshooting formatting issues
- Customizing AI prompts
- Investigating performance issues
- Learning how the AI works

## Validation and Retry Logic

### Automatic Retry

The command automatically retries up to 5 times if validation fails:

```text
Attempt 1: ✗ Invalid format - missing module in parentheses
Attempt 2: ✗ Invalid format - incorrect bullet structure
Attempt 3: ✓ Validation passed
```

### Common Auto-fixes

The retry logic automatically corrects:

- Missing module identifier in header
- Incorrect bullet formatting (uses `-` instead of `*` or `•`)
- Extra whitespace in headers or body
- Invalid commit type (suggests closest valid type)
- Missing or malformed file list

### Validation Rules

Commit messages are validated against:

1. **Format**: Must follow `<type>(<module>): <description>` pattern
2. **Commit type**: Must be one of: feat, fix, refactor, docs, chore, test, perf, style
3. **Module**: Must exist in project contracts or be "multi-module"
4. **Structure**: Must have description, details, and file list
5. **File list**: Must start with "Files modified:" and use bullet points

### Manual Intervention

If all 5 retry attempts fail:

```text
Error: Failed to generate valid commit message after 5 attempts.

Please check:
1. Staged changes are valid
2. Module contracts are up to date
3. Debug output in out/commit-message/

Try running with --debug for detailed diagnostics.
```

**Solutions:**

1. Check `out/commit-message/08-validation-result.json` for specific errors
2. Review module contracts: `r2r eac show modules`
3. Verify staged changes: `git diff --cached`
4. Check if changes match a clear pattern (feat, fix, etc.)

## Integration with Work Command

### Command Comparison

| Command | Use Case | Working Directory | Auto-staging |
|---------|----------|-------------------|--------------|
| `r2r eac create commit-message` | Direct commit message generation | Any git repo | No (manual `git add`) |
| `r2r eac work commit` | Workspace-aware commits | Inside work workspace | Yes (with `--all`) |

### Work Command Example

```bash
# Create workspace
r2r eac work create feature/authentication

# Develop feature
vim src/auth/jwt.go

# Use commit command directly (manual staging)
git add src/auth/
r2r eac create commit-message --commit

# Or use work commit (auto-staging)
r2r eac work commit --all
```

Both commands use the same AI engine and validation logic.

## Advanced Usage

### Custom Editor

Configure your preferred editor for commit message editing:

```bash
# Vim
export GIT_EDITOR="vim"
r2r eac create commit-message

# VS Code (wait for window to close)
export GIT_EDITOR="code --wait"
r2r eac create commit-message

# Nano
export GIT_EDITOR="nano"
r2r eac create commit-message
```

### Pre-commit Hook Integration

Automatically generate commit messages when committing:

```bash
# .git/hooks/prepare-commit-msg
#!/bin/bash

# Only generate message if commit message is empty
if [ -z "$2" ]; then
  r2r eac create commit-message --commit
fi
```

Make the hook executable:

```bash
chmod +x .git/hooks/prepare-commit-msg
```

### CI/CD Validation

Validate commit message format in continuous integration:

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

Commit multiple logical changes separately:

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

### Soft Reset

To undo the last commit while preserving changes:

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

## Configuration

### AI Provider Setup

Configure your AI provider on first use:

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

**Example customization:**

Edit `.r2r/eac/ai/commit/summary-prompt.md` to change how the AI generates summaries:

```markdown
# Custom Summary Prompt

Analyze the following git changes and generate a commit summary.

Focus on:
- Business value of the change
- User-facing impact
- Technical approach

Format: <type>(<module>): <description>
```

## Performance

### Parallel Processing

Multi-module commits process modules in parallel for speed:

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

**Performance factors:**

- Number of modules: More modules = more parallel processing benefit
- AI provider latency: Varies by provider and region
- Change complexity: Larger diffs take longer to analyze
- Retry attempts: Validation failures add overhead

### Caching

Context analysis results are cached during retry attempts to avoid re-processing:

- Git status (cached between retries)
- Git diff (cached between retries)
- Recent commits (cached between retries)
- Module ownership (cached between retries)

Cache is cleared after successful commit or command termination.

## Example Outputs

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

## See Also

- [How-to Guide](../../how-to-guides/commands/commit-command.md) - Quick start and common workflows
- [AI Process Explanation](../../explanation/commands/commit-ai-process.md) - How the AI generation works
- [Workspace Commands](../../how-to-guides/commands/areas/workspace-commands.md) - Workspace-aware development
