# Feature ID: src_mcp_vscode_module_detection
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode
# Part: 5 of 5 - Module Detection Edge Cases

@mcp @vscode @modules @edge-cases
Feature: Module Detection Edge Cases

  Background:
    Given the module detection system is initialized

  @success @ac8
  Scenario: Detect root module for README
    When I determine module for "README.md"
    Then the detected module is "root"

  @success @ac8
  Scenario: Detect root module for LICENSE
    When I determine module for "LICENSE"
    Then the detected module is "root"

  @success @ac8
  Scenario: Detect root module for gitignore
    When I determine module for ".gitignore"
    Then the detected module is "root"

  @success @ac8
  Scenario: Detect root module for package config
    When I determine module for "package.json"
    Then the detected module is "root"

  @success @ac10
  Scenario: Fallback for unrecognized single-level path
    When I determine module for "custom/file.txt"
    Then the detected module is "custom" or "root"

  @success @ac10
  Scenario: Fallback for deeply nested unrecognized path
    When I determine module for "unknown/path/to/nested/file.txt"
    Then the detected module is "unknown" or "root"
    And no error is thrown

  @success @ac10
  Scenario: Fallback for empty directory structure
    When I determine module for "file.txt"
    Then the detected module is "root"

  @success @ac9
  Scenario: Deterministic detection for repeated calls
    When I determine module for "src/mcp/vscode/main.go" 5 times
    Then all 5 results are "src-mcp-vscode"
    And the results are identical

  @success @ac1 @ac2 @ac3 @ac4 @ac5 @ac6 @ac7
  Scenario: Mixed module detection in single commit
    When I determine modules for multiple files:
      | File Path                                                  | Expected Module      |
      | src/mcp/vscode/main.go                                     | src-mcp-vscode       |
      | automation/cli/deploy/script.sh                            | cli-deploy           |
      | docs/guide/index.md                                        | docs                 |
      | .claude/agents/commit-gen.md                               | claude-config        |
      | contracts/deployable-units/0.1.0/src-mcp-vscode.yml        | src-mcp-vscode       |
      | requirements/cli/feature/acceptance.spec                   | cli                  |
      | containers/postgres/Dockerfile                             | postgres             |
      | README.md                                                  | root                 |
    Then each file is correctly mapped to its expected module
    And module detection is consistent and deterministic
