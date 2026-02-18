# Intent: Keep the codebase organized and scripts easy to find by ensuring scripts are only in designated locations
# Architecture: Affects repository script location validation; scans for .sh, .ps1, .psm1, .bat, .cmd files; enforces approved locations (scripts/pwsh, scripts/sh, scripts/cmd packages, .claude/hooks, containers, root importers); validates package naming convention

@L1 @ov @control:sa-3
Feature: repository_no-scattered-scripts

  As a repository maintainer
  I want to ensure scripts are only in designated locations
  So that the codebase remains organized and scripts are easy to find

  Background:
    Given the repository root exists

  Rule: Scripts must be in approved locations only

    Shell scripts (.sh, .ps1, .psm1, .bat, .cmd) may only exist in:
    - .claude/hooks/ - Claude Code hook scripts
    - scripts/<type>/<package>/ - Organized script packages
    - Repository root as importers (importer.sh, importer.ps1)

    Where <type> is one of: pwsh, sh, cmd
    And <package> is a descriptive name like: cli, claude, vscode, go-invoker

    @L1 @ov
    Scenario: All shell scripts are in approved locations
      Given the following script extensions are tracked:
        | Extension | Type       |
        | .sh       | Shell      |
        | .ps1      | PowerShell |
        | .psm1     | PowerShell |
        | .bat      | Batch      |
        | .cmd      | Batch      |
      When I scan the repository for script files
      Then all scripts should be in one of these locations:
        | Pattern                              | Purpose                    |
        | .claude/hooks/*                      | Claude Code hooks          |
        | scripts/pwsh/<package>/*             | PowerShell packages        |
        | scripts/sh/<package>/*               | Shell packages             |
        | scripts/cmd/<package>/*              | Batch packages             |
        | containers/*/entrypoint.sh           | Container entrypoints      |
        | containers/*/*.sh                    | Container scripts          |
        | importer.sh                          | Root shell importer        |
        | importer.ps1                         | Root PowerShell importer   |
      And no scripts should exist in:
        | Location              | Reason                              |
        | src/                  | Source code, not scripts            |
        | docs/                 | Documentation, not scripts          |
      And these locations are excluded from validation:
        | Location              | Reason                              |
        | out/                  | Output directory (not in repo)      |
        | .vscode/node_modules/ | Generated npm stubs (gitignored)    |

    @L1 @ov
    Scenario: Script packages follow naming convention
      When I scan the scripts directory structure
      Then each script type directory should contain only package subdirectories
      And package names should be lowercase with hyphens
      And each package should contain at least one script file

  Rule: No loose scripts in scripts root

    Scripts must be organized into packages, not placed directly in
    scripts/pwsh/, scripts/sh/, or scripts/cmd/.

    @L1 @ov
    Scenario: No scripts directly in type directories
      When I check the scripts directory structure
      Then scripts/pwsh/ should contain only directories
      And scripts/sh/ should contain only directories
      And scripts/cmd/ should contain only directories if it exists
