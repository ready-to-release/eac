@env:isolated-test-project
@ov @depm:eac-cli
Feature: eac-installer_cli-installation

  As a user
  I want to install the EAC CLI using platform-specific installer scripts
  So that I can quickly set up the CLI on my system without manual steps

  Background:
    Given the GitHub repository "ready-to-release/eac" has releases available

  Rule: Installer downloads and installs CLI binary

    @ov @deps:windows @depm:eac-cli @L3
    Scenario: Install latest version on Windows
      Given I am on Windows with PowerShell 5.1 or later
      When I run the PowerShell installer
      Then the latest version is fetched from GitHub API
      And the binary is installed to "%LOCALAPPDATA%\eac\eac.exe"
      And the installation is verified by running "eac --version"

    @ov @deps:linux @deps:curl @depm:eac-cli @L3
    Scenario: Install latest version on Linux
      Given I am on Linux with bash and curl available
      When I run the bash installer
      Then the latest version is fetched from GitHub API
      And the binary is installed to "$HOME/.local/bin/eac"
      And the installation is verified by running "eac --version"

  Rule: Installer handles errors gracefully

    @L2 @ov @deps:windows @depm:eac-cli
    Scenario: Windows installer fails when release does not exist
      Given I am on Windows
      When I run the PowerShell installer with "-Version v999.999.999"
      Then the installer displays an error about the failed download
      And exits with a non-zero exit code

    @L2 @ov @deps:linux @deps:curl @depm:eac-cli
    Scenario: Linux installer fails when release does not exist
      Given I am on Linux
      When I run the bash installer with "--version v999.999.999"
      Then the installer displays an error about the failed download
      And exits with a non-zero exit code
