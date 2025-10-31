# Module Detection

> **Feature ID**: src_mcp_vscode_module_detection
> **BDD Scenarios**: See feature files below
> **Module**: src-mcp-vscode
> **Tags**: mcp, vscode, modules, critical

## BDD Feature Files

This feature is split across multiple focused files for maintainability:

- [automation_module_detection.feature](./automation_module_detection.feature) - automation/ paths
- [source_module_detection.feature](./source_module_detection.feature) - src/mcp/ paths
- [infrastructure_module_detection.feature](./infrastructure_module_detection.feature) - containers, contracts
- [documentation_module_detection.feature](./documentation_module_detection.feature) - docs, .claude, requirements
- [module_detection_edge_cases.feature](./module_detection_edge_cases.feature) - fallbacks, consistency

## User Story

* As a commit generation system
* I want to automatically detect which module a file belongs to based on its path
* So that I can organize commits by module and generate accurate semantic messages

## Acceptance Criteria

* Detects module for automation/<type>/<name>/ paths
* Detects module for src/mcp/<service>/ paths
* Detects module for containers/<name>/ paths
* Detects module for contracts/deployable-units/<version>/<module>.yml files
* Detects module for docs/ paths
* Detects module for .claude/ configuration paths
* Detects module for requirements/<module>/ paths
* Handles root-level files appropriately
* Returns consistent module names for files in same directory
* Uses fallback logic for unrecognized paths

## Acceptance Tests

### AC1: Detects module for automation paths
**Validated by**: automation_module_detection.feature -> @ac1 scenarios

Tags: critical, modules, automation

* Pass file path "automation/cli/deploy/script.sh"
* Determine module name
* Verify module is "cli-deploy"
* Verify format is "<type>-<name>"

### AC2: Detects module for src/mcp paths
**Validated by**: source_module_detection.feature -> @ac2 scenarios

Tags: critical, modules, mcp

* Pass file path "src/mcp/vscode/main.go"
* Determine module name
* Verify module is "src-mcp-vscode"
* Verify format is "src-mcp-<service>"

### AC3: Detects module for container paths
**Validated by**: infrastructure_module_detection.feature -> @ac3 scenarios

Tags: modules, containers

* Pass file path "containers/postgres/Dockerfile"
* Determine module name
* Verify module is "postgres"
* Verify format is container name only

### AC4: Detects module from contract filenames
**Validated by**: infrastructure_module_detection.feature -> @ac4 scenarios

Tags: critical, modules, contracts

* Pass file path "contracts/deployable-units/0.1.0/src-mcp-vscode.yml"
* Determine module name
* Verify module is "src-mcp-vscode"
* Extract module from filename without extension

### AC5: Detects module for docs paths
**Validated by**: documentation_module_detection.feature -> @ac5 scenarios

Tags: modules, docs

* Pass file path "docs/guide/vscode-extension/index.md"
* Determine module name
* Verify module is "docs"
* All docs paths map to single "docs" module

### AC6: Detects module for .claude config paths
**Validated by**: documentation_module_detection.feature -> @ac6 scenarios

Tags: modules, config

* Pass file path ".claude/agents/commit-gen.md"
* Determine module name
* Verify module is "claude-config"
* All .claude paths map to "claude-config" module

### AC7: Detects module for requirements paths
**Validated by**: documentation_module_detection.feature -> @ac7 scenarios

Tags: modules, requirements

* Pass file path "requirements/src-mcp-vscode/feature/acceptance.spec"
* Determine module name
* Verify module is "src-mcp-vscode"
* Extract module from second path segment

### AC8: Handles root-level files
**Validated by**: module_detection_edge_cases.feature -> @ac8 scenarios

Tags: modules

* Pass file path "README.md"
* Determine module name
* Verify module is "root" or appropriate default
* Consistent handling for all root files

### AC9: Returns consistent module names for same directory
**Validated by**: module_detection_edge_cases.feature -> @ac9 scenarios

Tags: critical, modules, consistency

* Pass multiple files from "src/mcp/vscode/"
* Determine module for each file
* Verify all files have same module "src-mcp-vscode"
* Module detection is deterministic

### AC10: Uses fallback logic for unrecognized paths
**Validated by**: module_detection_edge_cases.feature -> @ac10 scenarios

Tags: modules

* Pass file path "unknown/custom/path/file.txt"
* Determine module name
* Verify fallback module is assigned
* Fallback uses first directory or "root"
* No crashes or errors occur
