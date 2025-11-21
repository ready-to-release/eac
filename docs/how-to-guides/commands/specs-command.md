# Specs Command

**Problem**: Writing comprehensive Gherkin specifications is time-consuming and requires BDD expertise.

**Solution**: Use `specs` to generate and validate Gherkin specifications from natural language descriptions.

## Key Benefits

- AI-powered specification generation from natural language
- Automatic validation against quality contracts
- Proper specification structure (Feature, Rule, Scenario)
- Organized by module in `specs/` directory
- Ensures team consistency and quality standards

## Quick Start

```bash
# Generate specification from description
r2r eac specs create "User authentication with JWT tokens"

# Validate existing specifications
r2r eac specs validate

# Validate specific file
r2r eac specs validate specs/src-auth/authentication.feature
```

## Typical Workflow

### Creating Specifications

```bash
# Simple feature
r2r eac specs create "Calculate shipping costs based on weight and distance"

# Specify target module
r2r eac specs create "User login validation" --module src-auth

# Debug AI generation
r2r eac specs create "Shopping cart checkout" --debug

# Custom output path
r2r eac specs create "Payment processing" --output specs/custom/payment.feature
```

### Validation Workflow

```bash
# Validate all specifications
r2r eac specs validate

# Output shows:
# ✓ specs/src-auth/authentication.feature - PASS
# ✗ specs/src-cart/checkout.feature - FAIL
#   - Missing Rule blocks
#   - Scenario lacks Given/When/Then structure

# Fix issues and revalidate
r2r eac specs validate specs/src-cart/checkout.feature
```

## Command Reference

### specs create

Generate Gherkin specification from natural language.

```bash
r2r eac specs create <description> [options]

# Options:
--module, -m <name>      # Target module (e.g., src-commands)
--output, -o <path>      # Custom output path
--debug, -d              # Save intermediate outputs to out/
--template <path>        # Custom template file
--prompt <path>          # Custom system prompt

# Examples:
r2r eac specs create "User registration with email verification"
r2r eac specs create "API rate limiting" --module src-api
r2r eac specs create "Password reset flow" --output specs/security/password-reset.feature
r2r eac specs create "Feature description" --debug
```

**What it does:**
1. Analyzes natural language description
2. Loads specification contract and rules
3. Generates Gherkin with proper structure
4. Validates output against quality standards
5. Auto-retries if validation fails
6. Saves to `specs/<module>/<feature-name>.feature`

### specs validate

Validate Gherkin specifications against contracts.

```bash
r2r eac specs validate [path] [options]

# Options:
--format <type>    # Output format: text (default) or json

# Examples:
r2r eac specs validate                           # Validate all .feature files
r2r eac specs validate specs/src-auth/           # Validate directory
r2r eac specs validate specs/auth.feature        # Validate single file
r2r eac specs validate --format json             # Machine-readable output
```

**Validation checks:**
- Gherkin syntax correctness
- Feature/Rule/Scenario hierarchy
- Tag formatting (@feature, @rule, @scenario)
- Step structure (Given/When/Then)
- Content quality standards

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

```
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
r2r eac specs create "Feature description" --debug
```

Creates debug files in `out/`:

```
out/
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
r2r eac specs create "Feature" --prompt custom-prompt.md
```

### Custom Templates

Use project-specific templates:

```bash
r2r eac specs create "Feature" --template templates/my-spec-template.md
```

### Contract Customization

Edit contracts in `.r2r/contracts/ai/specs-create/`:

```bash
# Modify system prompt
nano .r2r/contracts/ai/specs-create/system-prompt.md

# Changes apply to all future spec generation
r2r eac specs create "New feature"
```

## Validation Output

### Text Format (Default)

```
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
r2r eac specs validate --format json > validation.json
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
r2r eac specs create "Calculate tax based on location and amount"

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
r2r eac specs create "Existing payment processing logic"

# Validate against current implementation
r2r eac specs validate

# Use specs as living documentation
r2r eac docs serve  # View in MkDocs
```

### Team Review Workflow

```bash
# Developer creates spec
r2r eac specs create "New feature concept"

# Validate before commit
r2r eac specs validate

# Commit for team review
git add specs/
r2r eac work commit -m "docs: add specification for new feature"

# Team reviews spec in PR before implementation
r2r eac work pr
```

## Best Practices

- **Spec-first development**: Write specifications before code
- **Validate often**: Run `specs validate` before commits
- **Use descriptive names**: Clear feature and scenario descriptions
- **One feature per file**: Keep specifications focused
- **Commit specifications**: Track spec changes with code
- **Review specs in PRs**: Get team alignment early
- **Update specs with code**: Keep specifications current

## Troubleshooting

| Problem | Solution |
|---------|----------|
| AI generates invalid Gherkin | Use `--debug` to inspect output, check AI provider setup |
| Module not inferred correctly | Use `--module src-modulename` explicitly |
| Validation fails repeatedly | Check `.r2r/contracts/ai/specs-create/` for contract requirements |
| API key error | Run `r2r eac init --ai <provider>` first |
| Output path issues | Use `--output` to specify exact path |
| Complex feature generates poorly | Break into smaller features, be more specific in description |

## Advanced Usage

### Batch Generation

```bash
# Generate multiple specs
for desc in "User login" "User logout" "Password reset"; do
  r2r eac specs create "$desc" --module src-auth
done

# Validate all
r2r eac specs validate
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Validate specifications
  run: |
    r2r eac specs validate --format json > validation.json
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
  r2r eac specs validate || exit 1
fi
```

## Summary

1. **Create specs**: `r2r eac specs create "<description>"`
2. **Validate**: `r2r eac specs validate`
3. **Debug** (if needed): Add `--debug` flag
4. **Customize** (optional): Edit `.r2r/contracts/ai/specs-create/`
5. **Commit**: `git add specs/` and commit

Specifications drive BDD/TDD workflows and serve as living documentation for your project.
