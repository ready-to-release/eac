@L1 @ov @depm:core @deps:go @env:isolated-test-project
Feature: Cache Invalidation System

  As a developer using EAC
  I want the build cache to correctly detect when modules need rebuilding
  So that I don't waste time on unnecessary builds or miss required builds

  The cache invalidation system operates at two levels:
  1. Component-Level: File hash tracking (used locally via get changed-modules-local)
  2. Macro-Level: Module dependency chain (used in CI via get changed-modules-ci)

  Background:
    Given I am in an isolated test repository
    And the repository has a multi-module structure with:
      | moniker         | depends_on    | go_root            |
      | test-core       |               | go/test-core       |
      | test-commands   | test-core     | go/test-commands   |
      | test-api        | test-commands | go/test-api        |
      | test-standalone |               | go/test-standalone |

  # ===========================================================================
  # Category A: Fresh Build Detection
  # ===========================================================================

  Rule: Fresh builds detect all modules as changed

    Scenario: A1 - No build state marks all modules for build
      Given there is no build state file
      When I run "get changed-modules-local "
      Then the exit code is 0
      And the YAML output field "is_fresh_build" is "true"
      And the YAML output field "modules" contains "test-core"
      And the YAML output field "modules" contains "test-commands"
      And the YAML output field "modules" contains "test-api"
      And the YAML output field "modules" contains "test-standalone"

    Scenario: A2 - Deleted build state triggers fresh detection
      Given I have built all modules successfully
      When I delete the build state directory
      And I run "get changed-modules-local "
      Then the YAML output field "is_fresh_build" is "true"

  # ===========================================================================
  # Category B: Source File Change Detection
  # ===========================================================================

  Rule: Source file changes invalidate the owning module only

    Scenario: B1 - Modified Go file invalidates its module
      Given I have built all modules successfully
      When I append "// cache-test-comment" to file "go/test-core/main.go"
      And I run "get changed-modules-local "
      Then the YAML output field "modules" contains "test-core"
      And the YAML output field "change_reasons.test-core" contains "input hash mismatch"

    Scenario: B2 - Unchanged modules remain up-to-date
      Given I have built all modules successfully
      When I append "// cache-test-comment" to file "go/test-core/main.go"
      And I run "get changed-modules-local "
      Then the YAML output field "up_to_date" contains "test-standalone"

    Scenario: B3 - Multiple files in same module count as one change
      Given I have built all modules successfully
      When I append "// comment1" to file "go/test-core/main.go"
      And I append "// comment2" to file "go/test-core/helper.go"
      And I run "get changed-modules-local "
      Then the YAML output field "modules" contains exactly 1 occurrence of "test-core"

    Scenario: B4 - File in nested directory detected correctly
      Given I have built all modules successfully
      When I append "// nested-comment" to file "go/test-commands/internal/handler.go"
      And I run "get changed-modules-local "
      Then the YAML output field "modules" contains "test-commands"
      And the YAML output field "modules" does not contain "test-core"

  # ===========================================================================
  # Category C: Up-to-Date Detection (Fast Path)
  # ===========================================================================

  Rule: Unchanged modules use fast-path detection

    Scenario: C1 - No changes means all modules up-to-date
      Given I have built all modules successfully
      And no files have been modified
      When I run "get changed-modules-local "
      Then the YAML output field "modules" is empty
      And the YAML output field "up_to_date" contains "test-core"
      And the YAML output field "up_to_date" contains "test-commands"

    Scenario: C2 - Detection completes quickly when unchanged
      Given I have built all modules successfully
      And no files have been modified
      When I run "get changed-modules-local "
      Then the YAML output field "detection_time" indicates under 200 milliseconds

  # ===========================================================================
  # Category D: Build Dry-Run Verification
  # ===========================================================================

  Rule: Build dry-run reflects cache state correctly

    Scenario: D1 - Dry-run shows changed module would be built
      Given I have built all modules successfully
      When I append "// trigger-rebuild" to file "go/test-core/main.go"
      And I run "build --dry-run test-core"
      Then the exit code is 0
      And the output indicates "test-core" would be built

    Scenario: D2 - Dry-run shows up-to-date module would be skipped
      Given I have built all modules successfully
      And no files have been modified
      When I run "build --dry-run test-core"
      Then the exit code is 0
      And the output indicates "test-core" is up-to-date or would be skipped

  # ===========================================================================
  # Category E: Transitive Dependency Invalidation (Macro Level)
  # ===========================================================================

  Rule: Changes propagate through dependency chain in CI detection

    Scenario: E1 - Direct dependent is invalidated
      Given the mocked CI status shows:
        | module        | last_success_sha | has_files_changed |
        | test-core     | abc123           | true              |
        | test-commands | def456           | false             |
        | test-api      | ghi789           | false             |
      When I run "get changed-modules-ci " with mocked CI
      Then the YAML output field "directly_changed" contains "test-core"
      And the YAML output field "invalidated" contains "test-commands"

    Scenario: E2 - Multi-level transitive invalidation
      Given the mocked CI status shows:
        | module        | last_success_sha | has_files_changed |
        | test-core     | abc123           | true              |
        | test-commands | def456           | false             |
        | test-api      | ghi789           | false             |
      When I run "get changed-modules-ci " with mocked CI
      Then the YAML output field "invalidated" contains "test-commands"
      And the YAML output field "invalidated" contains "test-api"

    Scenario: E3 - Independent module not affected
      Given the mocked CI status shows:
        | module          | last_success_sha | has_files_changed |
        | test-core       | abc123           | true              |
        | test-standalone | jkl012           | false             |
      When I run "get changed-modules-ci " with mocked CI
      Then the YAML output field "invalidated" does not contain "test-standalone"
      And the YAML output field "skipped" contains "test-standalone"

  # ===========================================================================
  # Category F: CI-Excluded Files
  # ===========================================================================

  Rule: CI-excluded files do not trigger module invalidation

    Scenario: F1 - CHANGELOG.md change does not trigger CI
      Given the mocked CI shows "test-core" has valid CI at HEAD
      And the only changed file is "go/test-core/CHANGELOG.md"
      When I run "get changed-modules-ci " with mocked CI
      Then the YAML output field "skipped" contains "test-core"
      And the YAML output field "directly_changed" does not contain "test-core"

    Scenario: F2 - Source change alongside CHANGELOG still triggers CI
      Given the mocked CI shows "test-core" CI at different SHA
      And the changed files are "go/test-core/main.go" and "go/test-core/CHANGELOG.md"
      When I run "get changed-modules-ci " with mocked CI
      Then the YAML output field "directly_changed" contains "test-core"

  # ===========================================================================
  # Category G: Per-Module CI Status
  # ===========================================================================

  Rule: Each module has independent CI status tracking

    Scenario: G1 - Module with valid CI at HEAD is skipped
      Given the mocked CI shows "test-core" has valid CI at current HEAD
      When I run "get changed-modules-ci " with mocked CI
      Then the YAML output field "module_status.test-core.has_valid_ci" is "true"
      And the YAML output field "skipped" contains "test-core"

    Scenario: G2 - Module with no CI history needs CI
      Given the mocked CI shows "test-core" has no successful runs
      When I run "get changed-modules-ci " with mocked CI
      Then the YAML output field "modules" contains "test-core"
      And the YAML output field "module_status.test-core.reason" contains "no_ci"

  # ===========================================================================
  # Category H: Lint State Tracking
  # ===========================================================================

  Rule: Lint state tracks pass/fail independently

    Scenario: H1 - Passed lint skips unchanged module
      Given I have linted "test-core" successfully
      And no files in "go/test-core" have been modified
      When I run "lint --dry-run test-core"
      Then the output indicates "test-core" would be skipped

    Scenario: H2 - Failed lint re-runs even without changes
      Given I have a lint state showing "test-core" failed
      And no files in "go/test-core" have been modified
      When I run "lint --dry-run test-core"
      Then the output indicates "test-core" would be linted

  # ===========================================================================
  # Category K: Edge Cases
  # ===========================================================================

  Rule: Edge cases are handled gracefully

    Scenario: K1 - New module detected as needing build
      Given I have built modules "test-core" and "test-commands"
      And a new module "test-new" is configured with go_root "go/test-new"
      When I run "get changed-modules-local "
      Then the YAML output field "modules" contains "test-new"
      And the YAML output field "change_reasons.test-new" contains "no build manifests found"

    Scenario: K2 - Circular dependency is detected and rejected
      Given modules are configured with circular dependency:
        | moniker | depends_on |
        | mod-a   | mod-b      |
        | mod-b   | mod-a      |
      When I run "validate module-hierarchy"
      Then the exit code is not 0
      And the output contains "Circular"

    Scenario: K3 - Corrupted build state triggers fresh build
      Given the build state file contains invalid JSON "{ corrupted"
      When I run "get changed-modules-local "
      Then the YAML output field "is_fresh_build" is "true"
