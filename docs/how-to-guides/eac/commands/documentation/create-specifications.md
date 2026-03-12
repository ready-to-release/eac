# Create Specifications

## What You'll Accomplish

Generate Gherkin BDD specifications from natural language requirements using AI.

## Prerequisites

### Required Knowledge

**New to specifications?** Learn these concepts first:

- [BDD Fundamentals](../../../../explanation/specifications/concepts/bdd-fundamentals.md) - Understand Gherkin syntax and Given/When/Then pattern

### Required Setup

- AI provider configured
- Clear requirements or user stories
- Specifications directory exists

## How It Works

The command uses AI to generate specifications in valid Gherkin format:

- **Structured Generation**: AI generates specifications following Gherkin syntax rules
- **Automatic Validation**: Output is validated against Gherkin format requirements
- **Intelligent Retry**: If validation fails, the AI receives error feedback and regenerates improved output
- **Quality Assurance**: Generated specs are syntactically correct and ready to use

This ensures high-quality specifications every time.

## Steps

### 1. Generate Spec from Description

```bash
eac create spec "User can login with email and password"
```

**What happens**: AI generates a Gherkin feature file with scenarios, validates the syntax, and retries automatically if needed

### 2. Review Generated Spec

```bash
cat specs/<module>/<feature-name>/specification.feature
```

**What happens**: View generated Gherkin scenarios with proper tags and structure

**Note**: The path depends on your project's organization (e.g., `specs/auth/user-login/specification.feature`)

### 3. Validate Spec Quality

```bash
eac validate specs
```

**What happens**: Checks spec follows quality standards

### 4. Implement Step Definitions

Create step definitions to match generated steps:

```go
// steps/auth_steps.go
func (s *AuthSteps) UserIsOnLoginPage() error {
    // Implementation
}

func (s *AuthSteps) UserHasAccountWithEmail(email string) error {
    // Implementation
}
```

**Note**: Organize step definitions in a way that matches your project structure and testing framework (e.g., godog, cucumber-js)

## Example Scenario

Creating login specification:

```bash
# Generate spec
eac create spec "User can login with email and password to access dashboard"

# Output:
# Generating specification...
# ✓ Created specs/auth/user-login/specification.feature

# View generated spec
cat specs/auth/user-login/specification.feature

# @deps:go
# Feature: auth_user-login
#
#   As a user
#   I want to login with email and password
#   So that I can access my dashboard
#
#   @L2 @ov
#   Scenario: Successful login with valid credentials
#     Given user is on login page
#     And user has account with email "user@example.com"
#     When user enters email "user@example.com"
#     And user enters password "SecurePass123"
#     And user clicks login button
#     Then user should be redirected to dashboard
#     And user should see welcome message
#
#   @L2 @ov
#   Scenario: Failed login with invalid password
#     Given user is on login page
#     When user enters email "user@example.com"
#     And user enters password "wrongpassword"
#     And user clicks login button
#     Then user should see error "Invalid credentials"
#     And user should remain on login page

# Validate
eac validate specs
# ✓ specs/auth/user-login/specification.feature is valid
```

**Note**: Basic specification structure includes:

- **Feature naming**: Use descriptive names (e.g., `auth_user-login`)
- **Minimum tags**: Add `@deps:` for dependencies and test level tags like `@L2 @ov` on scenarios
- **Organize by module**: Group related features in subdirectories (e.g., `specs/auth/`, `specs/api/`)
- Each feature gets its own subdirectory with a `specification.feature` file

For advanced tagging (environments, OSCAL controls, Rules), see your project's tagging guidelines.

**Directory organization**:

- `specs/<module>/` - Organize by functional area (e.g., `auth`, `api`, `database`)
- `specs/<module>/<feature>/specification.feature` - Each feature in its own subdirectory
- Examples: `specs/auth/user-login/`, `specs/api/rest-endpoint/`, `specs/ui/dashboard/`

## Spec Components

Generated specifications include:

- **Feature** - High-level capability description
- **User Story** - As/I want/So that format
- **Tags** - Metadata for test classification (e.g., `@deps:go`, `@L2`, `@ov`)
- **Scenarios** - Specific test cases
- **Given/When/Then** - BDD steps describing behavior
- **Examples** - Data-driven scenarios (optional)

**Tagging**: Add tags based on your project's needs. Common tags include dependencies (`@deps:`), test levels (`@L0`-`@L4`), and verification types (`@ov`, `@iv`). Check your project's documentation for specific tagging requirements

## Common Issues

| Problem           | Solution                           |
| ----------------- | ---------------------------------- |
| Spec too generic  | Provide more detailed requirements |
| Missing scenarios | Add edge cases manually            |
| Undefined steps   | Implement step definitions         |

## Next Steps

- [Validate Specifications](../build-test-validate/validate-specifications.md) → Check quality

## Related Commands

- [`create spec`](../../../../reference/eac/commands/create/spec.md) - Generate specification
