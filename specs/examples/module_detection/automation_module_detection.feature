# Feature ID: src_mcp_vscode_module_detection
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode
# Part: 1 of 5 - Automation Module Detection

@mcp @vscode @modules @critical @automation
Feature: Automation Module Detection

  Background:
    Given the module detection system is initialized

  @success @ac1
  Scenario: Detect automation CLI deploy module
    When I determine module for "automation/cli/deploy/script.sh"
    Then the detected module is "cli-deploy"

  @success @ac1
  Scenario: Detect automation CLI build module
    When I determine module for "automation/cli/build/Makefile"
    Then the detected module is "cli-build"

  @success @ac1
  Scenario: Detect automation container registry module
    When I determine module for "automation/container/registry/config.yml"
    Then the detected module is "container-registry"

  @success @ac1
  Scenario: Detect deeply nested automation files
    When I determine module for "automation/cli/deploy/scripts/utils/helper.sh"
    Then the detected module is "cli-deploy"

  @success @ac9
  Scenario: Consistent detection across automation subdirectories
    When I determine module for "automation/cli/deploy/script1.sh"
    And I determine module for "automation/cli/deploy/subdir/script2.sh"
    And I determine module for "automation/cli/deploy/utils/helper.sh"
    Then all detected modules are "cli-deploy"
