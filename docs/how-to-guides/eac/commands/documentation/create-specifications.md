# Create Specifications

{{ page_breadcrumb() }}

## What You'll Accomplish

Generate Gherkin BDD specifications from natural language requirements using AI.

## Prerequisites

- AI provider configured
- Clear requirements or user stories
- Specifications directory exists

## Steps

### 1. Generate Spec from Description

```bash
r2r eac create spec "User can login with email and password"
```

**What happens**: AI generates Gherkin feature file with scenarios

### 2. Review Generated Spec

```bash
cat features/auth/login.feature
```

**What happens**: View generated Gherkin scenarios

### 3. Validate Spec Quality

```bash
r2r eac validate specs
```

**What happens**: Checks spec follows quality standards

### 4. Implement Step Definitions

Create step definitions to match generated steps:

```go
// steps/auth/login_steps.go
func (s *AuthSteps) UserEntersEmail(email string) error {
    // Implementation
}
```

## Example Scenario

Creating login specification:

```bash
# Generate spec
r2r eac create spec "User can login with email and password to access dashboard"

# Output:
# Generating specification...
# ✓ Created features/auth/login.feature

# View generated spec
cat features/auth/login.feature

# Feature: User Login
#   As a user
#   I want to login with email and password
#   So that I can access my dashboard
#
#   Scenario: Successful login with valid credentials
#     Given user is on login page
#     And user has account with email "user@example.com"
#     When user enters email "user@example.com"
#     And user enters password "SecurePass123"
#     And user clicks login button
#     Then user should be redirected to dashboard
#     And user should see welcome message
#
#   Scenario: Failed login with invalid password
#     Given user is on login page
#     When user enters email "user@example.com"
#     And user enters password "wrongpassword"
#     And user clicks login button
#     Then user should see error "Invalid credentials"
#     And user should remain on login page

# Validate
r2r eac validate specs
# ✓ features/auth/login.feature is valid
```

## Spec Components

Generated specifications include:

- **Feature** - High-level capability
- **User Story** - As/I want/So that format
- **Scenarios** - Specific test cases
- **Given/When/Then** - BDD steps
- **Examples** - Data-driven scenarios

## Common Issues

| Problem | Solution |
|---------|----------|
| Spec too generic | Provide more detailed requirements |
| Missing scenarios | Add edge cases manually |
| Undefined steps | Implement step definitions |

## Next Steps

- [Validate Specifications](../build-test-validate/validate-specifications.md) → Check quality

## Related Commands

- [`create spec`](../../../../reference/commands/create/spec.md) - Generate specification
- [`validate specs`](../../../../reference/commands/validate/specs.md) - Validate quality

{{ diataxis_footer() }}
