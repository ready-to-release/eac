# TDD: Test-Driven Development

**[<- Back to Testing Overview](./index.md)**

## Table of Contents

- [What is TDD?](#what-is-tdd)
- [TDD in This Project](#tdd-in-this-project)
- [The Red-Green-Refactor Cycle](#the-red-green-refactor-cycle)
- [Workflow](#workflow)
- [Unit Test Structure by Language](#unit-test-structure-by-language)
- [Style Rules](#style-rules)
- [Multi-Language Examples](#multi-language-examples)
- [Test Coverage by Language](#test-coverage-by-language)
- [Validation Checklist](#validation-checklist)
- [Integration with ATDD and BDD](#integration-with-atdd-and-bdd)
- [Related Resources](#related-resources)

---

## What is TDD?

**Test-Driven Development (TDD)** is a software development approach where you write tests **before** writing implementation code. In this project, TDD applies to **unit tests** that verify internal code correctness, distinct from the ATDD/BDD layers which specify requirements and behavior.

### Key Characteristics

| Aspect | Description |
|--------|-------------|
| **Who** | Developers |
| **When** | During implementation |
| **Format** | Unit tests with feature name in comments/attributes |
| **Location** | Test files alongside implementation (language-specific conventions) |
| **Purpose** | Ensure code correctness and support refactoring |

## TDD in This Project

### Separation of Concerns

| Layer | Location | Purpose |
|-------|----------|---------|
| ATDD | `acceptance.spec` (Gauge) | Business requirements and acceptance criteria |
| BDD | `behavior.feature` (Godog) | User-facing behavior scenarios |
| TDD | Test files | Implementation testing and code correctness |

### Language-Specific Test Files

| Language | Test File Pattern | Example |
|----------|-------------------|---------|
| **Go** | `*_test.go` | `config_test.go` |
| **Python** | `test_*.py` or `*_test.py` | `test_config.py` |

### Traceability Model

Unit tests link to feature files via comments:

#### Go

```go
// TestInitProject validates project initialization logic
// Feature: init_project
func TestInitProject(t *testing.T) {
    // Test implementation
}
```

#### Python

```python
# Feature: init_project
def test_init_project_creates_directory():
    """Validates project initialization logic"""
    # Test implementation
```

**Traceability**:

- Feature directory: `requirements/cli/init_project/`
- ATDD spec: `acceptance.spec`
- BDD scenarios: `behavior.feature`
- Unit test comment: `Feature: init_project`
- Search codebase for `Feature: init_project` to find all related tests

## The Red-Green-Refactor Cycle

TDD follows a simple, repeating cycle:

```text
+-----------------------------------------------------------------+
|                  Red-Green-Refactor Cycle                       |
+-----------------------------------------------------------------+

1. RED: Write a Failing Test
   +- Read BDD scenarios from behavior.feature to understand expected behavior
   +- Write test for one small piece of functionality
   +- Run test and verify it fails (RED)
   +- Test should fail for the right reason
      
2. GREEN: Make It Pass
   +- Write minimal code to make test pass
   +- Don't worry about perfection
   +- Run test and verify it passes (GREEN)
   +- Avoid over-engineering
      
3. REFACTOR: Improve the Code
   +- Clean up implementation
   +- Remove duplication
   +- Improve naming and structure
   +- Run tests to ensure they still pass
   +- Keep tests green throughout refactoring
      
4. Repeat
   +- Continue cycle for next piece of functionality
```

## Workflow

### TDD Development Process

```text
+-----------------------------------------------------------------+
|                    TDD Workflow                                 |
+-----------------------------------------------------------------+

1. Read BDD Scenarios
   +- Review behavior.feature in requirements/<module>/<feature>/
   +- Understand expected behavior from Given/When/Then
   +- Note the feature directory name for traceability
   +- Identify internal components needed
      
2. Write Failing Test (RED)
   +- Create or open test file
   +- Add feature comment: // Feature: <feature_name>
   +- Write test function with descriptive name
   +- Exercise the functionality (even if it doesn't exist)
   +- Add assertions for expected behavior
   +- Run test and verify it fails
      
3. Implement Minimal Code (GREEN)
   +- Write just enough code to pass the test
   +- Avoid over-engineering
   +- Focus on making test green
   +- Run test and verify it passes
      
4. Refactor (CLEAN)
   +- Improve code structure
   +- Remove duplication
   +- Enhance readability
   +- Run tests after each change
   +- Keep all tests passing
      
5. Verify Test Coverage
   +- Run coverage tools (language-specific)
   +- Aim for >80% coverage
   +- Identify untested code paths
   +- Add tests for edge cases
      
6. Link to Feature File
   +- Ensure Feature: comment is present
   +- Verify feature name matches file name
   +- Search codebase to verify traceability
      
7. Verify BDD Scenarios
   +- Run acceptance tests (if automated)
   +- Manually verify scenarios pass
   +- Confirm ATDD acceptance criteria met
```

### Prerequisites

Before starting TDD:

- BDD scenarios exist in behavior.feature
- Feature directory name is known (for traceability)
- Test file location identified (language-specific)

### Outputs

After completing TDD:

- Unit tests written with `Feature:` comments
- All tests passing (GREEN)
- Code is clean and refactored
- Test coverage >80%
- BDD scenarios verified
- ATDD acceptance criteria met

---

## Unit Test Structure by Language

### Go

**File Organization**:

```text
src/
+-- config/
|   +-- config.go          # Implementation
|   +-- config_test.go     # Tests
```

**Test Function Format**:

```go
package config

import "testing"

// TestFunctionName describes what is being tested
// Feature: feature_file_name
func TestFunctionName(t *testing.T) {
    // Arrange: Set up test data

    // Act: Execute functionality

    // Assert: Verify outcomes
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

**Running Tests**:

```bash
go test ./...
go test -cover ./...
```

### Python

**File Organization**:

```text
src/
+-- config/
|   +-- __init__.py
|   +-- config.py              # Implementation
tests/
+-- __init__.py
+-- test_config.py             # Tests
```

**Test Function Format** (pytest):

```python
import pytest

# Feature: feature_file_name
def test_function_name_describes_test():
    """Describes what is being tested"""
    # Arrange
    expected = "value"

    # Act
    result = function_under_test()

    # Assert
    assert result == expected
```

**Running Tests**:

```bash
pytest
pytest --cov=src tests/
```

---

## Style Rules

### (check) Do

- **Add feature comment** to every test: `// Feature: feature_name`
- **Use descriptive test names**: Clear purpose
- **Write tests first** (Red-Green-Refactor)
- **Test one thing per test**: Single responsibility
- **Use table/parameterized tests** for multiple cases
- **Test edge cases**: Empty input, nil/null values, errors
- **Use setup/teardown** appropriately
- **Use temp directories** for file operations
- **Verify error messages**: Not just error existence
- **Fail fast**: Stop test when precondition fails

### (X) Don't

- **Forget feature comment**: Always link to feature file
- **Skip Red phase**: Always see test fail first
- **Write tests after code**: TDD means tests first
- **Test implementation details**: Test behavior, not internals
- **Ignore test failures**: Fix or update immediately
- **Use external dependencies**: Mock or stub external services
- **Write flaky tests**: Tests must be deterministic
- **Skip refactor phase**: Clean code is critical

---

## Multi-Language Examples

### Example 1: Simple Function Test

**BDD Scenario** (from `requirements/cli/init_project/behavior.feature`):

```gherkin
Scenario: Initialize creates config file
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And the file should contain valid YAML
```

#### Go Implementation

```go
package config

import (
    "os"
    "path/filepath"
    "testing"
)

// TestCreateConfig validates config file creation
// Feature: init_project
func TestCreateConfig(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "cc.yaml")

    // Act
    err := CreateConfig(configPath)

    // Assert
    if err != nil {
        t.Fatalf("CreateConfig failed: %v", err)
    }

    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        t.Error("Config file was not created")
    }
}
```

#### Python Implementation

```python
import os
import tempfile
import pytest

# Feature: init_project
def test_create_config_creates_file():
    """Validates config file creation"""
    # Arrange
    with tempfile.TemporaryDirectory() as tmpdir:
        config_path = os.path.join(tmpdir, 'cc.yaml')

        # Act
        create_config(config_path)

        # Assert
        assert os.path.exists(config_path)
```

### Example 2: Parameterized/Table-Driven Tests

**BDD Scenario** (from `requirements/cli/handle_config_errors/behavior.feature`):

```gherkin
Scenario: Handle invalid field types
  Given a file has invalid field types
  When I validate the config
  Then specific type errors should be reported
```

#### Go Implementation

```go
// Feature: handle_config_errors
func TestValidateConfig_InvalidTypes(t *testing.T) {
    tests := []struct {
        name      string
        config    Config
        wantError string
    }{
        {
            name:      "version as string",
            config:    Config{Version: "1.0.0"},
            wantError: "Invalid type for 'version'",
        },
        {
            name:      "valid config",
            config:    Config{Name: "test", Version: 1.0},
            wantError: "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateConfig(tt.config)
            // Assertions...
        })
    }
}
```

#### Python Implementation

```python
# Feature: handle_config_errors
@pytest.mark.parametrize("config,expected_error", [
    ({"version": "1.0.0"}, "Invalid type for 'version'"),
    ({"name": "test", "version": 1.0}, None),
])
def test_validate_config_invalid_types(config, expected_error):
    """Validates type checking"""
    # Act
    error = validate_config(config)

    # Assert
    if expected_error:
        assert expected_error in str(error)
    else:
        assert error is None
```

### Example 3: Error Handling

#### Go

```go
// Feature: init_project
func TestInitProject_NonEmptyDirectory(t *testing.T) {
    tmpDir := t.TempDir()
    os.WriteFile(filepath.Join(tmpDir, "existing.txt"), []byte("content"), 0644)

    err := InitProject(tmpDir)

    if err == nil {
        t.Fatal("Expected error for non-empty directory")
    }

    if !strings.Contains(err.Error(), "Directory must be empty") {
        t.Errorf("Error = %q, want substring %q", err.Error(), "Directory must be empty")
    }
}
```

#### Python

```python
# Feature: init_project
def test_init_project_non_empty_directory():
    """Validates rejection of non-empty directories"""
    with tempfile.TemporaryDirectory() as tmpdir:
        existing_file = os.path.join(tmpdir, 'existing.txt')
        with open(existing_file, 'w') as f:
            f.write('content')

        with pytest.raises(ValueError, match="Directory must be empty"):
            init_project(tmpdir)
```

## Test Coverage by Language

### Go

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Python

```bash
pytest --cov=src tests/
pytest --cov=src --cov-report=html tests/
```

### Coverage Guidelines

- **Target**: >80% coverage
- **Focus**: Cover critical paths first
- **Don't**: Chase 100% at expense of quality
- **Do**: Test edge cases and error paths
- **Review**: Uncovered code may indicate missing tests

## Validation Checklist

Use this checklist when reviewing unit tests:

### Test Structure

- [ ] Test file follows language conventions
- [ ] Test function has descriptive name
- [ ] Feature comment present: `Feature: feature_name`
- [ ] Feature name matches feature file name

### Test Quality

- [ ] Follows Arrange-Act-Assert pattern
- [ ] Tests one behavior per function
- [ ] Uses temp directories for file operations
- [ ] Verifies error messages (not just error existence)
- [ ] Uses appropriate assertions for language

### Coverage

- [ ] Happy path tested
- [ ] Error cases tested
- [ ] Edge cases tested (nil/null, empty, boundary values)
- [ ] Overall coverage >80%

### Integration

- [ ] Test links to feature file via comment
- [ ] Test verifies BDD scenario behavior
- [ ] Test supports ATDD acceptance criteria

---

## Integration with ATDD and BDD

### From BDD to TDD

BDD scenarios in behavior.feature guide unit test creation.

**BDD Scenario** (from `requirements/cli/init_project/behavior.feature`):

```gherkin
Scenario: Initialize creates config file
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And the file should contain valid YAML
```

**TDD Tests** (shown in any language):

- Test file creation
- Test YAML validity
- Test file permissions
- Test error handling

### From ATDD to TDD

ATDD acceptance criteria from acceptance.spec inform test priorities.

**ATDD Criterion** (from `requirements/cli/init_project/acceptance.spec`):

```markdown
* Command completes in under 2 seconds
```

**TDD Test** (Python example):

```python
# Feature: init_project
def test_init_project_performance():
    """Validates initialization speed"""
    import time
    start = time.time()
    init_project(tmpdir)
    duration = time.time() - start

    assert duration < 2.0, f"Too slow: {duration}s"
```

---

## Related Resources

- **[ATDD Guide](./atdd.md)** - Define business value and acceptance criteria
- **[BDD Guide](./bdd.md)** - Write scenarios that guide implementation
- **[Testing Overview](./index.md)** - Understand the complete testing strategy
