# Git Context Collection

> **Feature ID**: src_mcp_vscode_git_context_collection
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: src-mcp-vscode
> **Tags**: mcp, vscode, git, critical

## User Story

* As a commit generation system
* I want to collect comprehensive git context from the repository
* So that I can generate accurate commit messages based on actual changes

## Acceptance Criteria

* Executes git status command and captures output
* Executes git diff --staged command and captures diff content
* Retrieves current HEAD commit SHA
* Parses git status output into file change list
* Attributes module name to each changed file
* Creates GitContext data structure with all information
* Handles git command failures gracefully
* Detects when no staged changes exist

## Acceptance Tests

### AC1: Executes git status and captures output
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, git

* Execute git status command in workspace root
* Verify command completes successfully
* Capture stdout output
* Parse status for staged files
* Extract file paths from status output

### AC2: Executes git diff --staged and captures diff
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: critical, git

* Execute git diff --staged command in workspace root
* Verify command completes successfully
* Capture full diff output
* Include diff in GitContext structure
* Preserve diff formatting

### AC3: Retrieves current HEAD commit SHA
**Validated by**: behavior.feature -> @ac3 scenarios

Tags: git

* Execute git rev-parse HEAD command
* Capture 40-character SHA hash
* Include SHA in GitContext structure
* Verify SHA format is valid

### AC4: Parses git status into file change list
**Validated by**: behavior.feature -> @ac4 scenarios

Tags: critical, parsing

* Parse git status output line by line
* Identify staged file changes (A, M, D, R)
* Create FileChange object for each file
* Extract file path for each change
* Extract change type (Added, Modified, Deleted, Renamed)
* Store changes in GitContext.Changes array

### AC5: Attributes module name to each changed file
**Validated by**: behavior.feature -> @ac5 scenarios

Tags: critical, modules

* For each FileChange parse the file path
* Call module detection logic
* Assign module name to FileChange.Module
* Verify module attribution is correct for all files

### AC6: Creates GitContext data structure
**Validated by**: behavior.feature -> @ac6 scenarios

Tags: critical, data

* Create GitContext structure
* Populate HeadSHA field
* Populate StatusOutput field
* Populate DiffOutput field
* Populate Changes array with FileChange objects
* Return populated GitContext

### AC7: Handles git command failures gracefully
**Validated by**: behavior.feature -> @ac7 scenarios

Tags: error, git

* Execute git command in non-git directory
* Catch command execution error
* Return descriptive error message
* Do not crash or panic
* Include git error details in message

### AC8: Detects when no staged changes exist
**Validated by**: behavior.feature -> @ac8 scenarios

Tags: validation, git

* Execute git status in repo with no staged changes
* Parse status output
* Detect empty staged changes
* Return error or empty Changes array
* Provide clear feedback about missing changes
