# Common Flags and Options

## Overview

EAC commands share common flags and patterns that work consistently across the command set. Understanding these global options helps you use commands more effectively.

## Global Flags

### Help Flag

**Available on**: All commands

```bash
--help, -h          Display help for the command
```

**Usage**:

```bash
# Get help for any command
eac build --help
eac test --help
eac get modules --help

# Short form
eac build -h
```

**Output**: Command-specific help including:

- Command purpose and description
- Syntax and usage examples
- Available flags and options
- Common use cases

---

### Debug Flag

**Available on**: AI-powered commands (`create` category)

```bash
--debug, -d         Save intermediate AI outputs for inspection
```

**Usage**:

```bash
# Debug commit message generation
eac create commit-message --debug

# Debug specification generation
eac create spec "User can login" --debug

# Debug PR description generation
eac create pr --debug
```

**Output location**: `out/` directory with timestamped subdirectories

**What it saves**:

- AI prompts sent
- AI responses received
- Intermediate processing steps
- Validation results
- Final outputs

**When to use**:

- AI generates unexpected output
- Debugging AI behavior
- Understanding the generation process
- Customizing AI prompts

---

### Skip Cache Flag

**Available on**: `build`, `test`, `lint`, `scan`, `get ci-dispatch`

```bash
--skip-cache <spec>    Bypass specific cache types
```

The `--skip-cache` flag uses a 2D taxonomy (Level x Type) to target specific caches.

**Spec format**: `level:type`, `level`, `type`, or `all`

| Spec             | Effect                                              |
|------------------|-----------------------------------------------------|
| `local:state`    | Ignore UoW manifests (force rebuild)                |
| `local:asset`    | Re-render all diagrams (Mermaid, Structurizr)       |
| `local:layer`    | Docker build with `--no-cache`                      |
| `local:registry` | Force `docker pull` (ignore local image cache)      |
| `remote:ci`      | Force CI dispatch for all modules                   |
| `remote:layer`   | Skip remote BuildKit cache                          |
| `local`          | Skip all local caches                               |
| `remote`         | Skip all remote caches                              |
| `all`            | Skip all caches                                     |

**Usage**:

```bash
# Force rebuild ignoring local state
eac build --skip-cache=local:state

# Re-render all diagrams
eac build --skip-cache=local:asset

# Force CI dispatch for all modules
eac get ci-dispatch --skip-cache=remote:ci

# Combine multiple specs
eac build --skip-cache=local:state,local:asset

# Skip everything
eac build --skip-cache=all
```

See [Cache System](../../architecture/cache-system.md) for the full 2D taxonomy reference.

---

## Common Flag Patterns

### Module Selection

Many commands accept module monikers as arguments:

```bash
# Single module
eac build eac-commands
eac test src-auth

# Multiple modules
eac build eac-commands src-auth src-api
eac test eac-commands eac-core
```

**Rules**:

- Space-separated list
- Use exact module monikers (see `show modules`)
- Order doesn't matter (dependencies are resolved automatically)

---

### Output Formatting

#### JSON Output (get commands)

All `get` commands output JSON by default:

```bash
eac get modules
# Output: {"modules": [...]}
```

**Processing with jq**:

```bash
# Extract specific fields
eac get modules | jq -r '.modules[].moniker'

# Filter results
eac get modules | jq '.modules[] | select(.type == "go-library")'

# Count results
eac get modules | jq '.modules | length'
```

#### Formatted Output (show commands)

`show` commands output human-readable formats:

- Tables
- Lists
- Formatted text
- Colorized output (when terminal supports it)

---

### Flags for Specific Commands

#### build command

```bash
--parallel, -p      Build modules in parallel (default: true)
--force, -f         Force rebuild even if up-to-date
```

**Usage**:

```bash
# Sequential build
eac build --parallel=false eac-commands

# Force rebuild
eac build --force eac-commands
```

#### test command

```bash
--parallel, -p      Run tests in parallel (default: true)
--verbose, -v       Verbose test output
--short             Run only short tests
--coverage          Generate coverage report
```

**Usage**:

```bash
# Sequential tests
eac test --parallel=false src-auth

# With coverage
eac test --coverage src-auth

# Short tests only (for pre-commit)
eac test --short src-auth
```

#### test suite command

```bash
--parallel, -p      Run tests in parallel (default: true)
--tags              Run tests with specific tags
--stop-on-failure   Stop on first failure
```

**Usage**:

```bash
# Run specific tags
eac test suite acceptance --tags @auth,@login

# Stop on first failure
eac test suite integration --stop-on-failure
```

#### scan commands

```bash
--severity          Filter by severity (HIGH,CRITICAL,MEDIUM,LOW)
--format            Output format (table, json, sarif)
--output            Write results to file
```

**Usage**:

```bash
# Only high and critical vulnerabilities
eac scan --scanner vuln --severity HIGH,CRITICAL

# JSON output for CI
eac scan --scanner secrets --format json --output secrets-report.json

# SARIF format for GitHub integration
eac scan --scanner sast --format sarif --output sast-results.sarif
```

#### validate commands

```bash
--fix               Auto-fix issues where possible
--strict            Strict validation (fail on warnings)
```

**Usage**:

```bash
# Auto-fix markdown issues
eac validate markdown --fix docs/

# Strict validation
eac validate specs --strict
```

#### create commit-message

```bash
--commit, -c        Automatically create the commit (skip editor)
--debug, -d         Save intermediate outputs
```

**Usage**:

```bash
# Generate and commit automatically
eac create commit-message --commit

# Debug AI generation
eac create commit-message --debug

# Both
eac create commit-message --commit --debug
```

#### work commands

```bash
# work commit
--all, -a           Stage all changes before committing
--message, -m       Specify commit message (skip AI generation)

# work create
--from              Create from branch other than main
```

**Usage**:

```bash
# Commit all changes with AI message
eac work commit --all

# Commit with manual message
eac work commit --all --message "fix: resolve auth bug"

# Create workspace from release branch
eac work create hotfix/security-patch --from release/v1.2.0
```

#### release commands

```bash
# release changelog
--from              Start from specific commit
--to                End at specific commit

# release check-ci
--branch            Branch to check CI status for
--workflow          Workflow name to check
```

**Usage**:

```bash
# Generate changelog for specific range
eac release changelog --from v1.0.0 --to HEAD

# Check CI for specific branch
eac release check-ci --branch release/v1.2.0
```

#### pipeline commands

```bash
# pipeline run
--stage             Run specific stage only
--parallel          Run stages in parallel where possible

# pipeline wait
--timeout           Timeout in minutes (default: 60)
--interval          Check interval in seconds (default: 30)
```

**Usage**:

```bash
# Run specific stage
eac pipeline run --stage build clie

# Wait with custom timeout
eac pipeline wait --timeout 120 --interval 60
```

---

## Flag Value Formats

### Boolean Flags

Boolean flags can be specified in multiple ways:

```bash
# Enable
--parallel
--parallel=true
--parallel true

# Disable
--no-parallel
--parallel=false
--parallel false
```

### String Flags

String values can be quoted or unquoted:

```bash
# Unquoted (if no spaces)
--message "fix auth bug"
--branch feature/new-auth

# Quoted (if spaces)
--message "fix: resolve authentication bug"
```

### List Flags

Lists can be comma-separated or space-separated:

```bash
# Comma-separated
--severity HIGH,CRITICAL
--tags @auth,@login

# Space-separated
--severity HIGH CRITICAL
--tags @auth @login
```

---

## Environment Variables

Some commands respect environment variables for configuration:

### AI Provider Configuration

```bash
ANTHROPIC_API_KEY    # Anthropic API key (for create commands)
OPENAI_API_KEY       # OpenAI API key (for create commands)
GOOGLE_API_KEY       # Google API key (for create commands)
```

**Usage**:

```bash
# Set temporarily
export ANTHROPIC_API_KEY="sk-ant-api03-..."
eac create commit-message

# Set permanently (add to ~/.bashrc or ~/.zshrc)
echo 'export ANTHROPIC_API_KEY="sk-ant-api03-..."' >> ~/.bashrc
```

### GitHub Integration

```bash
GITHUB_TOKEN         # GitHub personal access token (for pipeline commands)
```

**Usage**:

```bash
# Required for pipeline ci commands
export GITHUB_TOKEN="ghp_..."
eac pipeline ci dispatch-and-wait
```

### Editor Configuration

```bash
EDITOR               # Default editor for commit messages
GIT_EDITOR           # Git-specific editor (overrides EDITOR)
```

**Usage**:

```bash
# Set default editor
export EDITOR="vim"
eac create commit-message

# Set git-specific editor
export GIT_EDITOR="code --wait"
eac create commit-message
```

---

## Exit Codes

All commands follow standard exit code conventions:

| Exit Code | Meaning           | When Used                      |
| --------- | ----------------- | ------------------------------ |
| 0         | Success           | Command completed successfully |
| 1         | General error     | Command failed                 |
| 2         | Usage error       | Invalid arguments or flags     |
| 3         | Validation failed | Quality gates not met          |

**Usage in scripts**:

```bash
#!/bin/bash
set -e  # Exit on any error

# Build will exit with non-zero on failure
eac build src-auth

# Test will exit with non-zero on failure
eac test src-auth

# Validate will exit with non-zero on failure
eac validate specs

# If we reach here, all commands succeeded
echo "All checks passed!"
```

**Check exit codes**:

```bash
# Check if command succeeded
eac test src-auth
if [ $? -eq 0 ]; then
  echo "Tests passed"
else
  echo "Tests failed"
  exit 1
fi

# Or use directly in conditionals
if eac validate specs; then
  echo "Specs valid"
else
  echo "Specs invalid"
fi
```

---

## Standard Input/Output/Error

### Standard Output (stdout)

- Command results
- JSON output from `get` commands
- Formatted tables from `show` commands

**Usage**:

```bash
# Redirect stdout to file
eac get modules > modules.json
eac show modules > modules.txt

# Pipe to other commands
eac get modules | jq '.modules[].moniker'
```

### Standard Error (stderr)

- Error messages
- Warning messages
- Progress indicators
- Debug output

**Usage**:

```bash
# Redirect stderr to file
eac build src-auth 2> build-errors.txt

# Redirect both stdout and stderr
eac test src-auth > test-output.txt 2>&1

# Suppress stderr
eac build src-auth 2>/dev/null
```

### Standard Input (stdin)

Some commands accept input from stdin:

```bash
# Pipe file list to validation
find docs/ -name "*.md" | eac validate markdown

# Pipe commits to changelog generation
git log --oneline | eac release changelog
```

---

## Configuration Files

### Repository Configuration

`.eac/eac-config.yml` - EAC configuration:

```yaml
provider: claude-api
api_key_env: ANTHROPIC_API_KEY
model: claude-sonnet-4-5
max_tokens: 4096
temperature: 0.7
```

**Set with**: `eac init`

**See also**: [Init Command](../init/init.md)

---

## Best Practices

### Flag Usage

1. **Use long form in scripts** for clarity:

   ```bash
   # Good: Clear intent
   eac test --coverage --verbose src-auth

   # Avoid: Unclear in scripts
   eac test -cv src-auth
   ```

2. **Use short form interactively** for speed:

   ```bash
   # Interactive use
   eac build -f src-auth
   eac test -v src-auth
   ```

3. **Quote values with spaces**:

   ```bash
   # Good
   eac work commit --message "fix: resolve auth bug"

   # Bad (will fail)
   eac work commit --message fix: resolve auth bug
   ```

4. **Check command help** when unsure:

   ```bash
   eac <command> --help
   ```

### Error Handling

1. **Always check exit codes** in scripts:

   ```bash
   if ! eac test src-auth; then
     echo "Tests failed"
     exit 1
   fi
   ```

2. **Capture stderr for debugging**:

   ```bash
   eac build src-auth 2>&1 | tee build.log
   ```

3. **Use set -e** to fail fast:

   ```bash
   #!/bin/bash
   set -e
   eac validate specs
   eac test src-auth
   eac build src-auth
   ```

---

## See Also

- [Command Taxonomy](./command-taxonomy.md) - How commands are organized
- [Naming Conventions](./naming-conventions.md) - Command naming patterns
- [Output Formats](./output-formats.md) - Understanding command output
- [Help Command](../help/help.md) - Using the help system
- [Init Command](../init/init.md) - Initialize AI configuration
