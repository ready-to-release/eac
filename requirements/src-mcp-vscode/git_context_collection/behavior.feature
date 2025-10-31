# Feature ID: src_mcp_vscode_git_context_collection
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode

@mcp @vscode @git @critical
Feature: Git Context Collection

  Background:
    Given the workspace has a valid git repository
    And I am in the workspace root directory

  @success @ac1
  Scenario: Execute git status successfully
    Given there are staged changes in the repository
    When I gather git context
    Then git status command is executed
    And the status command completes without errors
    And the status output is captured
    And the status output contains staged file information

  @success @ac1
  Scenario: Parse git status for multiple staged files
    Given I have staged 3 files with changes
    When I gather git context
    Then git status output includes all 3 files
    And each file is marked as staged

  @error @ac1
  Scenario: Handle git status failure
    Given I am in a non-git directory
    When I attempt to gather git context
    Then git status command fails
    And an error is returned
    And the error describes the git failure

  @success @ac2
  Scenario: Execute git diff --staged successfully
    Given there are staged changes in the repository
    When I gather git context
    Then git diff --staged command is executed
    And the diff command completes without errors
    And the full diff content is captured
    And the diff shows line-by-line changes

  @success @ac2
  Scenario: Capture diff for multiple file types
    Given I have staged changes in .go, .md, and .yml files
    When I gather git context
    Then the diff output includes changes for all file types
    And the diff preserves formatting for each file type

  @success @ac2
  Scenario: Handle empty diff when files are staged but unchanged
    Given files are staged but have no actual changes
    When I gather git context
    Then git diff --staged returns empty output
    And the empty diff is captured in GitContext

  @success @ac3
  Scenario: Retrieve current HEAD SHA
    Given the repository has commits
    When I gather git context
    Then git rev-parse HEAD is executed
    And the HEAD SHA is captured
    And the SHA is a 40-character hexadecimal string
    And the SHA is stored in GitContext.HeadSHA

  @success @ac3
  Scenario: Validate HEAD SHA format
    When I gather git context
    Then the captured HEAD SHA matches pattern ^[0-9a-f]{40}$
    And the SHA corresponds to the current HEAD commit

  @error @ac3
  Scenario: Handle detached HEAD state
    Given the repository is in detached HEAD state
    When I gather git context
    Then git rev-parse HEAD still succeeds
    And the detached HEAD SHA is captured

  @success @ac4
  Scenario: Parse staged file additions
    Given I have staged 2 new files
    When I gather git context
    Then 2 FileChange objects are created
    And each FileChange has Type "Added"
    And each FileChange has the correct Path

  @success @ac4
  Scenario: Parse staged file modifications
    Given I have staged modifications to 3 existing files
    When I gather git context
    Then 3 FileChange objects are created
    And each FileChange has Type "Modified"
    And each FileChange has the correct Path

  @success @ac4
  Scenario: Parse staged file deletions
    Given I have staged 1 file deletion
    When I gather git context
    Then 1 FileChange object is created
    And the FileChange has Type "Deleted"
    And the FileChange has the correct Path

  @success @ac4
  Scenario: Parse mixed change types
    Given I have staged 2 additions, 3 modifications, and 1 deletion
    When I gather git context
    Then 6 FileChange objects are created
    And 2 have Type "Added"
    And 3 have Type "Modified"
    And 1 has Type "Deleted"

  @success @ac4
  Scenario: Parse renamed files
    Given I have staged 1 file rename
    When I gather git context
    Then 1 FileChange object is created
    And the FileChange has Type "Renamed"
    And the FileChange captures both old and new paths

  @success @ac5
  Scenario: Attribute module to src/mcp/vscode files
    Given I have staged "src/mcp/vscode/main.go"
    When I gather git context
    Then the FileChange has Module "src-mcp-vscode"

  @success @ac5
  Scenario: Attribute module to automation files
    Given I have staged "automation/cli/deploy/script.sh"
    When I gather git context
    Then the FileChange has Module "cli-deploy"

  @success @ac5
  Scenario: Attribute module to docs files
    Given I have staged "docs/guide/index.md"
    When I gather git context
    Then the FileChange has Module "docs"

  @success @ac5
  Scenario: Attribute module to multiple files from different modules
    Given I have staged files from 3 different modules
    When I gather git context
    Then each FileChange has the correct module attribution
    And modules are distinct for files in different directories

  @success @ac5
  Scenario: Attribute module to root-level files
    Given I have staged "README.md" in repository root
    When I gather git context
    Then the FileChange has Module "root" or appropriate default

  @success @ac6
  Scenario: Create complete GitContext structure
    Given I have staged changes in the repository
    When I gather git context
    Then a GitContext object is created
    And GitContext.HeadSHA is populated
    And GitContext.StatusOutput is populated
    And GitContext.DiffOutput is populated
    And GitContext.Changes is a non-empty array
    And the GitContext is returned

  @success @ac6
  Scenario: GitContext contains all file changes
    Given I have 5 staged files
    When I gather git context
    Then GitContext.Changes contains 5 FileChange objects
    And each FileChange has Path, Type, and Module set
    And all fields are properly populated

  @success @ac6
  Scenario: GitContext preserves raw git outputs
    Given I have staged changes
    When I gather git context
    Then GitContext.StatusOutput contains raw git status text
    And GitContext.DiffOutput contains raw git diff text
    And the raw outputs are unmodified

  @error @ac7
  Scenario: Handle git not installed
    Given git is not installed on the system
    When I attempt to gather git context
    Then the git command execution fails
    And an error is returned
    And the error indicates git is not available

  @error @ac7
  Scenario: Handle git command timeout
    Given git diff command hangs indefinitely
    When I attempt to gather git context
    Then the operation times out
    And an error is returned
    And the error describes the timeout

  @error @ac7
  Scenario: Handle corrupted git repository
    Given the .git directory is corrupted
    When I attempt to gather git context
    Then git commands fail
    And an error is returned
    And the error describes the repository issue

  @success @ac8
  Scenario: Detect no staged changes
    Given the repository has unstaged changes only
    When I gather git context
    Then git status shows no staged files
    And GitContext.Changes is an empty array
    And an appropriate warning or error is provided

  @success @ac8
  Scenario: Detect empty repository
    Given the repository has no commits
    When I attempt to gather git context
    Then an error is returned
    And the error indicates the repository is empty

  @success @ac8
  Scenario: Provide clear feedback when nothing to commit
    Given the working directory is clean
    When I gather git context
    Then the system detects no changes
    And a clear message indicates nothing to commit

  @success @ac1 @ac2 @ac3 @ac4 @ac5 @ac6
  Scenario: Full git context gathering workflow
    Given I have staged changes in multiple modules
    When I gather git context
    Then git status is executed and parsed
    And git diff --staged is executed and captured
    And HEAD SHA is retrieved
    And all file changes are parsed
    And modules are attributed to each file
    And a complete GitContext is returned
    And the GitContext is ready for commit generation
