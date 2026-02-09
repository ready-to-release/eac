@env:isolated-test-project
@ov @depm:clie
Feature: clie-installer_cli-installation

  As a user
  I want to install the CLIE CLI using platform-specific installer scripts
  So that I can quickly set up the CLI on my system without manual steps

  Background:
    Given the GitHub repository "ready-to-release/eac" has releases available

  Rule: Installer downloads and installs CLI binary

    @ov @deps:windows @depm:clie @L3
    Scenario: Install latest version on Windows
      Given I am on Windows with PowerShell 5.1 or later
      When I run the PowerShell installer
      Then the latest version is fetched from GitHub API
      And the binary is installed to "%LOCALAPPDATA%\clie\clie.exe"
      And the installation is verified by running "clie --version"

    @ov @deps:linux @deps:curl @depm:clie @L3
    Scenario: Install latest version on Linux
      Given I am on Linux with bash and curl available
      When I run the bash installer
      Then the latest version is fetched from GitHub API
      And the binary is installed to "$HOME/.local/bin/clie"
      And the installation is verified by running "clie --version"

  Rule: Installer handles errors gracefully

    @L2 @ov @deps:windows @depm:clie
    Scenario: Windows installer fails when release does not exist
      Given I am on Windows
      When I run the PowerShell installer with "-Version v999.999.999"
      Then the installer displays an error about the failed download
      And exits with a non-zero exit code

    @L2 @ov @deps:linux @deps:curl @depm:clie
    Scenario: Linux installer fails when release does not exist
      Given I am on Linux
      When I run the bash installer with "--version v999.999.999"
      Then the installer displays an error about the failed download
      And exits with a non-zero exit code
