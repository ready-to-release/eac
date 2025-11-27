@ov
Feature: src-cli_module-isolation

  As a repository maintainer
  I want to ensure src/cli production code has no local module dependencies
  So that the CLI binary remains lightweight and independently distributable

  Rule: src/cli production code must not import other local modules

    Production code includes all .go files except:
    - Files in test directories (*/tests/*)
    - Test files (*_test.go)

    Test code (godog tests, unit tests) MAY import other local modules like src/core.

    @L2 @ov
    Scenario: Production code has no local module imports
      Given I am checking module "src/cli"
      When I scan all production .go files in "src/cli"
      Then no production files should import local modules from "github.com/ready-to-release/eac/src/core"
      And no production files should import local modules from "github.com/ready-to-release/eac/src/commands"
      And no production files should import any other local modules outside "src/cli"

    @L2 @ov
    Scenario: Test code can import other local modules
      Given I am checking module "src/cli"
      When I scan test .go files in "src/cli/tests"
      Then test files MAY import local modules like "github.com/ready-to-release/eac/src/core"
      And this is allowed for test infrastructure purposes

  Rule: src/cli go.mod may have local module dependencies only for test code

    The presence of local module dependencies in src/cli/go.mod is acceptable
    if they are only used by test code.

    @L2 @ov
    Scenario: go.mod dependencies are test-only
      Given I am checking module "src/cli"
      And the go.mod file lists "github.com/ready-to-release/eac/src/core" as a dependency
      When I verify the dependency is only used in test files
      Then the dependency should only be imported by files matching:
        | Pattern                |
        | src/cli/tests/*.go     |
        | src/cli/**/*_test.go   |
      And no production files should import this dependency

  Rule: Rationale for isolation

    The src/cli module isolation ensures:
    - CLI binary has minimal dependencies
    - CLI can be distributed independently
    - CLI has fast build times
    - CLI avoids pulling in heavy dependencies like go-git
    - Test infrastructure can still leverage shared testing utilities
