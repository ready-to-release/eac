# Intent: Allow developers to quickly run Go commands from their shell session by importing the Go command invoker module using platform-specific scripts
# Architecture: Affects implicit-cli importer scripts (importer.ps1 and importer.sh); loads go-invoker module from scripts/pwsh/go-invoker/ or scripts/sh/go-invoker/; supports --no-alias flag; handles missing module gracefully

@env:isolated-test-project
@ov @depm:eac
Feature: implicit-cli_importer

  As a developer
  I want to import the Go command invoker module using platform-specific scripts
  So that I can quickly run Go commands from my shell session

  Rule: Importer loads the command invoker module

    @ov @deps:windows @depm:eac @L2
    Scenario: PowerShell importer loads module successfully
      Given I am on Windows with PowerShell
      And the go-invoker module exists at "scripts/pwsh/go-invoker/go.psm1"
      When I run the PowerShell script "importer.ps1"
      Then the output contains "CommandRunner module imported successfully"
      And the exit code is 0

    @ov @deps:linux @depm:eac @L2
    Scenario: Bash importer loads module successfully
      Given I am on Linux with bash
      And the go-invoker module exists at "scripts/sh/go-invoker/go.sh"
      When I source the bash script "importer.sh"
      Then the output contains "go-invoker module imported successfully"
      And the exit code is 0

  Rule: Importer handles missing module gracefully

    @ov @deps:windows @depm:eac @L2
    Scenario: PowerShell importer fails when module is missing
      Given I am on Windows with PowerShell
      And the go-invoker module does not exist
      When I run the PowerShell script "importer.ps1"
      Then the output contains "Module not found"
      And the exit code is non-zero

    @ov @deps:linux @depm:eac @L2
    Scenario: Bash importer fails when module is missing
      Given I am on Linux with bash
      And the go-invoker module does not exist
      When I source the bash script "importer.sh"
      Then the output contains "Module not found"
      And the exit code is non-zero

  Rule: Importer supports --no-alias flag

    @ov @deps:windows @depm:eac @L2
    Scenario: PowerShell importer skips aliases with -NoAlias
      Given I am on Windows with PowerShell
      And the go-invoker module exists at "scripts/pwsh/go-invoker/go.psm1"
      When I run the PowerShell script "importer.ps1 -NoAlias"
      Then the output contains "Invoke-GoSrcCommand"
      And the exit code is 0

    @ov @deps:linux @depm:eac @L2
    Scenario: Bash importer skips aliases with --no-alias
      Given I am on Linux with bash
      And the go-invoker module exists at "scripts/sh/go-invoker/go.sh"
      When I source the bash script "importer.sh --no-alias"
      Then the output contains "invoke_go_src_command"
      And the exit code is 0
