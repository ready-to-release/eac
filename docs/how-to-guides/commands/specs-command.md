# Specs Commands (Create & Validate)

**Problem**: Writing comprehensive Gherkin specifications is time-consuming and requires BDD expertise.

**Solution**: Use `create spec` to generate and `validate specs` to validate Gherkin specifications from natural language descriptions.

## Key Benefits

- AI-powered specification generation from natural language
- Automatic validation against quality contracts
- Proper specification structure (Feature, Rule, Scenario)
- Organized by module in `specs/` directory
- Ensures team consistency and quality standards

## Quick Start

```bash
# Generate specification from description
r2r eac create spec "User authentication with JWT tokens"

# Validate existing specifications
r2r eac validate specs

# Validate specific file
r2r eac validate specs specs/src-auth/authentication.feature
```

## Typical Workflow

### Creating Specifications

```bash
# Simple feature
r2r eac create spec "Calculate shipping costs based on weight and distance"

# Specify target module
r2r eac create spec "User login validation" --module src-auth

# Debug AI generation
r2r eac create spec "Shopping cart checkout" --debug

# Custom output path
r2r eac create spec "Payment processing" --output specs/custom/payment.feature

# Overwrite existing specification
r2r eac create spec "Updated user authentication flow" --module src-auth --force
```

### Validation Workflow

```bash
# Validate all specifications
r2r eac validate specs

# Output shows:
# ✓ specs/src-auth/authentication.feature - PASS
# ✗ specs/src-cart/checkout.feature - FAIL
#   - Missing Rule blocks
#   - Scenario lacks Given/When/Then structure
#   - Invalid tag format

# Auto-fix correctable issues (tag formatting, naming)
r2r eac validate specs --fix

# Skip tag validation if needed
r2r eac validate specs --no-check-tags

# Fix remaining issues manually and revalidate
r2r eac validate specs specs/src-cart/checkout.feature
```

## Command Reference

### create spec

Generate Gherkin specification from natural language.

```bash
r2r eac create spec <description> [options]

# Options:
--module, -m <name>      # Target module (e.g., src-commands)
--output, -o <path>      # Custom output path
--debug, -d              # Save intermediate outputs to out/logs/specs/
--template <path>        # Custom template file
--prompt <path>          # Custom system prompt
--force, -f              # Overwrite existing specification files

# Examples:
r2r eac create spec "User registration with email verification"
r2r eac create spec "API rate limiting" --module src-api
r2r eac create spec "Password reset flow" --output specs/security/password-reset.feature
r2r eac create spec "Feature description" --debug
r2r eac create spec "Update existing spec" --force
```

**What it does:**

1. Analyzes natural language description
2. Loads specification contract and rules
3. Generates Gherkin with proper structure
4. Validates output against quality standards
5. Auto-retries if validation fails
6. Saves to `specs/<module>/<feature-name>.feature`

### validate specs

Validate Gherkin specifications against contracts.

```bash
r2r eac validate specs [path] [options]

# Options:
--format <type>           # Output format: text (default) or json
--fix                     # Auto-fix correctable tag issues and naming problems
--check-tags              # Enable tag validation (default: true)
--no-check-tags           # Disable tag validation

# Examples:
r2r eac validate specs                           # Validate all .feature files
r2r eac validate specs specs/src-auth/           # Validate directory
r2r eac validate specs specs/auth.feature        # Validate single file
r2r eac validate specs --format json             # Machine-readable output
r2r eac validate specs --fix                     # Auto-fix tag and naming issues
r2r eac validate specs --no-check-tags           # Skip tag validation
```

**Validation checks:**

- Gherkin syntax correctness
- Feature/Rule/Scenario hierarchy
- Tag formatting (@feature, @rule, @scenario)
- Step structure (Given/When/Then)
- Content quality standards

**Auto-fix capabilities** (with `--fix` flag):

- Corrects tag formatting issues (e.g., @test → @scenario)
- Fixes naming convention problems
- Updates deprecated tag patterns
- Preserves file content and structure while fixing metadata

## Generated Specification Format

### Example Output

```gherkin
@feature
Feature: User Authentication
  As a registered user
  I want to log in with my credentials
  So that I can access my account

  @rule
  Rule: Valid credentials grant access

    @scenario @positive
    Scenario: Successful login with valid credentials
      Given the user exists with email "user@example.com" and password "SecurePass123"
      When the user submits login form with email "user@example.com" and password "SecurePass123"
      Then the system authenticates the user
      And the system creates a session token
      And the user is redirected to the dashboard

  @rule
  Rule: Invalid credentials deny access

    @scenario @negative
    Scenario: Login fails with incorrect password
      Given the user exists with email "user@example.com"
      When the user submits login form with email "user@example.com" and incorrect password
      Then the system rejects the authentication
      And the user sees an error message "Invalid credentials"
      And no session token is created
```

### Structure

- **Feature**: Top-level business capability
- **Rule**: Business rule or constraint
- **Scenario**: Concrete example of the rule
- **Tags**: Categorization (@feature, @rule, @scenario, @positive, @negative)
- **Steps**: Given (context), When (action), Then (outcome)

## Module Organization

Specifications are organized by module:

```text
specs/
├── src-auth/
│   ├── authentication.feature
│   └── authorization.feature
├── src-api/
│   ├── rate-limiting.feature
│   └── versioning.feature
└── src-core/
    └── validation.feature
```

Module detection:

- **Explicit**: `--module src-auth`
- **Inferred**: AI analyzes description to determine module
- **Default**: `specs/` root if unclear

## Debug Mode

Use `--debug` to inspect AI generation process:

```bash
r2r eac create spec "Feature description" --debug
```

Creates debug files in `out/logs/specs/`:

```text
out/logs/specs/
├── debug-full-prompt.md           # Complete AI prompt
├── debug-raw-ai-response.md       # Unfiltered AI output
├── debug-cleaned-output.feature   # After anti-corruption layer
└── debug-validation-result.json   # Validation details
```

**Use debug mode when:**

- AI generates unexpected output
- Validation fails repeatedly
- Customizing prompts or templates
- Understanding AI reasoning

## Customization

### Custom Prompts

Override default AI behavior:

```bash
# Create custom prompt
cat > custom-prompt.md << 'EOF'
Generate a Gherkin specification following these rules:
- Use business language
- Focus on user value
- Include edge cases
EOF

# Use custom prompt
r2r eac create spec "Feature" --prompt custom-prompt.md
```

### Custom Templates

Use project-specific templates:

```bash
r2r eac create spec "Feature" --template templates/my-spec-template.md
```

### AI Config Customization

Edit AI configs in `.r2r/eac/ai/specifications/`:

```bash
# Modify system prompt
nano .r2r/eac/ai/specifications/specification.md

# Changes apply to all future spec generation
r2r eac create spec "New feature"
```

## Validation Output

### Text Format (Default)

```text
Validating specifications...

✓ specs/src-auth/authentication.feature
  Feature: User Authentication (5 scenarios)

✗ specs/src-cart/checkout.feature
  Errors:
    - Line 12: Missing Rule block before Scenario
    - Line 25: Scenario missing When step
    - Line 30: Invalid tag format: use @scenario not @test

Summary:
  Total: 15 files
  Passed: 14
  Failed: 1
```

### JSON Format

```bash
r2r eac validate specs --format json > validation.json
```

```json
{
  "summary": {
    "total": 15,
    "passed": 14,
    "failed": 1
  },
  "results": [
    {
      "file": "specs/src-auth/authentication.feature",
      "status": "pass",
      "scenarios": 5
    },
    {
      "file": "specs/src-cart/checkout.feature",
      "status": "fail",
      "errors": [
        {
          "line": 12,
          "message": "Missing Rule block before Scenario"
        }
      ]
    }
  ]
}
```

## Integration Patterns

### TDD Workflow

```bash
# 1. Write specification first
r2r eac create spec "Calculate tax based on location and amount"

# 2. Implement step definitions
# Create tests/steps/tax_steps.go

# 3. Run tests (failing)
r2r eac test suite tax

# 4. Implement feature code
# Edit src/tax/calculator.go

# 5. Run tests (passing)
r2r eac test suite tax

# 6. Commit with AI-generated message
r2r eac work commit --all
```

### Documentation Workflow

```bash
# Generate specs for existing features
r2r eac create spec "Existing payment processing logic"

# Validate against current implementation
r2r eac validate specs

# Use specs as living documentation
r2r eac serve docs  # View in MkDocs
```

### Team Review Workflow

```bash
# Developer creates spec
r2r eac create spec "New feature concept"

# Validate before commit
r2r eac validate specs

# Commit for team review
git add specs/
r2r eac work commit -m "docs: add specification for new feature"

# Team reviews spec in PR before implementation
r2r eac work pr
```

## Best Practices

- **Spec-first development**: Write specifications before code
- **Validate often**: Run `validate specs` before commits
- **Use descriptive names**: Clear feature and scenario descriptions
- **One feature per file**: Keep specifications focused
- **Commit specifications**: Track spec changes with code
- **Review specs in PRs**: Get team alignment early
- **Update specs with code**: Keep specifications current

## Troubleshooting

| Problem                          | Solution                                                                      |
| -------------------------------- | ----------------------------------------------------------------------------- |
| AI generates invalid Gherkin     | Use `--debug` to inspect output in `out/logs/specs/`, check AI provider setup |
| Module not inferred correctly    | Use `--module src-modulename` explicitly                                      |
| Validation fails repeatedly      | Check `.r2r/eac/ai/specifications/` for AI config requirements                |
| API key error                    | Run `r2r eac init --ai <provider>` first                                      |
| Output path issues               | Use `--output` to specify exact path                                          |
| Complex feature generates poorly | Break into smaller features, be more specific in description                  |
| File already exists              | Use `--force` flag to overwrite existing specifications                       |
| Tag validation errors            | Use `--fix` to auto-correct tag issues, or `--no-check-tags` to skip          |

## Advanced Usage

### Batch Generation

```bash
# Generate multiple specs
for desc in "User login" "User logout" "Password reset"; do
  r2r eac create spec "$desc" --module src-auth
done

# Validate all
r2r eac validate specs
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Validate specifications
  run: |
    r2r eac validate specs --format json > validation.json
    if [ $? -ne 0 ]; then
      cat validation.json
      exit 1
    fi
```

### Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
if git diff --cached --name-only | grep -q '\.feature$'; then
  echo "Validating Gherkin specifications..."
  r2r eac validate specs || exit 1
fi
```

## Summary

1. **Create specs**: `r2r eac create spec "<description>"`
2. **Validate**: `r2r eac validate specs`
3. **Debug** (if needed): Add `--debug` flag
4. **Customize** (optional): Edit `.r2r/eac/ai/specifications/`
5. **Commit**: `git add specs/` and commit

Specifications drive BDD/TDD workflows and serve as living documentation for your project.
