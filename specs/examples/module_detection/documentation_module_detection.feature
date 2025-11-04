# Feature ID: src_mcp_vscode_module_detection
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode
# Part: 4 of 5 - Documentation Module Detection

@mcp @vscode @modules @documentation
Feature: Documentation Module Detection

  Background:
    Given the module detection system is initialized

  @success @ac5
  Scenario: Detect docs module for guide
    When I determine module for "docs/guide/vscode-extension/index.md"
    Then the detected module is "docs"

  @success @ac5
  Scenario: Detect docs module for API reference
    When I determine module for "docs/api/mcp-protocol.md"
    Then the detected module is "docs"

  @success @ac5
  Scenario: Detect docs module for root doc files
    When I determine module for "docs/README.md"
    Then the detected module is "docs"

  @success @ac5
  Scenario: All docs paths map to same module
    When I determine module for "docs/guide/intro.md"
    And I determine module for "docs/api/reference.md"
    And I determine module for "docs/tutorials/getting-started.md"
    Then all detected modules are "docs"

  @success @ac6
  Scenario: Detect claude-config for agent files
    When I determine module for ".claude/agents/commit-gen.md"
    Then the detected module is "claude-config"

  @success @ac6
  Scenario: Detect claude-config for command files
    When I determine module for ".claude/commands/review.md"
    Then the detected module is "claude-config"

  @success @ac6
  Scenario: Detect claude-config for root config
    When I determine module for ".claude/config.yml"
    Then the detected module is "claude-config"

  @success @ac6
  Scenario: All .claude paths map to same module
    When I determine module for ".claude/agents/reviewer.md"
    And I determine module for ".claude/hooks/pre-commit.sh"
    And I determine module for ".claude/templates/issue.md"
    Then all detected modules are "claude-config"

  @success @ac7
  Scenario: Detect requirements module for src-mcp-vscode
    When I determine module for "specs/src-mcp-vscode/feature/acceptance.spec"
    Then the detected module is "src-mcp-vscode"

  @success @ac7
  Scenario: Detect requirements module for cli
    When I determine module for "specs/cli/testing/behavior.feature"
    Then the detected module is "cli"

  @success @ac7
  Scenario: Detect deeply nested requirements files
    When I determine module for "specs/src-mcp-vscode/semantic-commit/v2/acceptance.spec"
    Then the detected module is "src-mcp-vscode"
