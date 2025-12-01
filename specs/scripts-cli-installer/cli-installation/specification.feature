@env:isolated-test-project
@ov @depm:scripts-cli-installer
Feature: scripts-cli-installer_cli-installation

  As a user
  I want to install the R2R CLI using platform-specific installer scripts
  So that I can quickly set up the CLI on my system without manual steps

  Background:
    Given the GitHub repository "ready-to-release/eac" has releases available

  Rule: Installer downloads and installs CLI binary

    @ov @deps:windows @depm:scripts-cli-installer @L3
    Scenario: Install latest version on Windows
      Given I am on Windows with PowerShell 5.1 or later
      When I run the PowerShell installer
      Then the latest version is fetched from GitHub API
      And the binary is installed to "%LOCALAPPDATA%\r2r\r2r.exe"
      And the installation is verified by running "r2r --version"

    @ov @deps:linux @depm:scripts-cli-installer @L3
    Scenario: Install latest version on Linux
      Given I am on Linux with bash and curl available
      When I run the bash installer
      Then the latest version is fetched from GitHub API
      And the binary is installed to "$HOME/.local/bin/r2r"
      And the installation is verified by running "r2r --version"

  Rule: Installer handles errors gracefully

    @L2 @ov @deps:windows @depm:scripts-cli-installer
    Scenario: Windows installer fails when release does not exist
      Given I am on Windows
      When I run the PowerShell installer with "-Version v999.999.999"
      Then the installer displays an error about the failed download
      And exits with a non-zero exit code

    @L2 @ov @deps:linux @depm:scripts-cli-installer
    Scenario: Linux installer fails when release does not exist
      Given I am on Linux
      When I run the bash installer with "--version v999.999.999"
      Then the installer displays an error about the failed download
      And exits with a non-zero exit code
