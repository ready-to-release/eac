@L1 @ov @depm:foundation
Feature: Tool Resolution for Component Types

  As a repository maintainer
  I want EAC to resolve the correct tools for each component type
  So that builds, tests, and scans use the appropriate tools for each language

  The tool system maps component types (from component-types.yml) to specific tools
  (from tool-config.yml) for each operation (build, test, lint, scan).

  Background:
    Given I am in an isolated test repository

  # ===========================================================================
  # Category A: Single-Language Tool Resolution
  # ===========================================================================

  Rule: Each language component type resolves to its appropriate builder

    @deps:go
    Scenario: A1 - Go module resolves to go builder
      Given I copy the test layout "single-go-module"
      When I load the EAC configuration
      Then the module "test-module" has component "go"
      And the builder for component type "go" is "go"

    Scenario: A2 - TypeScript module resolves to npm builder
      Given I copy the test layout "single-typescript-module"
      When I load the EAC configuration
      Then the module "test-module" has component "typescript"
      And the builder for component type "typescript" is "npm-build"

    Scenario: A3 - Python module resolves to python builder
      Given I copy the test layout "single-python-module"
      When I load the EAC configuration
      Then the module "test-module" has component "python"
      And the builder for component type "python" is "python"

    Scenario: A4 - .NET module resolves to dotnet builder
      Given I copy the test layout "single-dotnet-module"
      When I load the EAC configuration
      Then the module "test-module" has component "dotnet"
      And the builder for component type "dotnet" is "dotnet"

    Scenario: A5 - Rust module resolves to cargo builder
      Given I copy the test layout "single-rust-module"
      When I load the EAC configuration
      Then the module "test-module" has component "rust"
      And the builder for component type "rust" is "cargo"

    Scenario: A6 - Dockerfile module resolves to buildx builder
      Given I copy the test layout "single-dockerfile-module"
      When I load the EAC configuration
      Then the module "test-module" has component "dockerfile"
      And the builder for component type "dockerfile" is "buildx"

  # ===========================================================================
  # Category B: Multi-Language Tool Resolution
  # ===========================================================================

  Rule: Mixed-language repos resolve multiple builders correctly

    Scenario: B1 - Go and TypeScript repo resolves both builders
      Given I copy the test layout "multi-go-typescript"
      When I load the EAC configuration
      Then the configuration has 2 modules
      And the module "api-service" has component "go"
      And the module "web-app" has component "typescript"
      And the builder for component type "go" is "go"
      And the builder for component type "typescript" is "npm-build"

    Scenario: B2 - Python and TypeScript repo resolves both builders
      Given I copy the test layout "multi-python-typescript"
      When I load the EAC configuration
      Then the configuration has 2 modules
      And the module "backend" has component "python"
      And the module "frontend" has component "typescript"
      And the builder for component type "python" is "python"
      And the builder for component type "typescript" is "npm-build"

    Scenario: B3 - Three language repo resolves all builders
      Given I copy the test layout "multi-go-python-dotnet"
      When I load the EAC configuration
      Then the configuration has 3 modules
      And the module "api-service" has component "go"
      And the module "ml-service" has component "python"
      And the module "web-api" has component "dotnet"
      And the builder for component type "go" is "go"
      And the builder for component type "python" is "python"
      And the builder for component type "dotnet" is "dotnet"

    Scenario: B4 - Full stack repo resolves all six language builders
      Given I copy the test layout "multi-full-stack"
      When I load the EAC configuration
      Then the configuration has 6 modules
      And the module "api-gateway" has component "go"
      And the module "web-frontend" has component "typescript"
      And the module "ml-engine" has component "python"
      And the module "core-services" has component "dotnet"
      And the module "cli-tools" has component "rust"
      And the module "containers" has component "dockerfile"

  # ===========================================================================
  # Category C: Tester Tool Resolution
  # ===========================================================================

  Rule: Each component type resolves to its appropriate tester

    Scenario: C1 - Go module resolves to go tester
      Given I copy the test layout "single-go-module"
      When I load the EAC configuration
      Then the tester for component type "go" is "go"

    Scenario: C2 - TypeScript module resolves to npm tester
      Given I copy the test layout "single-typescript-module"
      When I load the EAC configuration
      Then the tester for component type "typescript" is "npm-test"

    Scenario: C3 - Python module resolves to pytest
      Given I copy the test layout "single-python-module"
      When I load the EAC configuration
      Then the tester for component type "python" is "pytest"

    Scenario: C4 - Rust module resolves to cargo tester
      Given I copy the test layout "single-rust-module"
      When I load the EAC configuration
      Then the tester for component type "rust" is "cargo"

    Scenario: C5 - .NET module resolves to dotnet tester
      Given I copy the test layout "single-dotnet-module"
      When I load the EAC configuration
      Then the tester for component type "dotnet" is "dotnet"

  # ===========================================================================
  # Category D: Scanner Tool Resolution
  # ===========================================================================

  Rule: Each component type has appropriate scanners assigned

    Scenario: D1 - Go module has vulnerability scanner
      Given I copy the test layout "single-go-module"
      When I load the EAC configuration
      Then the component type "go" has scanner "trivy-vuln"
      And the component type "go" has scanner "semgrep"

    Scenario: D2 - Python module has security scanners
      Given I copy the test layout "single-python-module"
      When I load the EAC configuration
      Then the component type "python" has scanner "trivy-vuln"
      And the component type "python" has scanner "trivy-secrets"
      And the component type "python" has scanner "semgrep"

    Scenario: D3 - Dockerfile module has IaC scanner
      Given I copy the test layout "single-dockerfile-module"
      When I load the EAC configuration
      Then the component type "dockerfile" has scanner "trivy-iac"

  # ===========================================================================
  # Category E: Component Type File Patterns
  # ===========================================================================

  Rule: Component types have correct file patterns

    Scenario: E1 - Go component has correct extensions
      When I load the EAC configuration
      Then the component type "go" has extension ".go"

    Scenario: E2 - Python component has correct extensions
      When I load the EAC configuration
      Then the component type "python" has extension ".py"

    Scenario: E3 - Rust component has correct extensions
      When I load the EAC configuration
      Then the component type "rust" has extension ".rs"

    Scenario: E4 - .NET component has correct extensions
      When I load the EAC configuration
      Then the component type "dotnet" has extension ".cs"
      And the component type "dotnet" has extension ".fs"

    Scenario: E5 - TypeScript component has correct extensions
      When I load the EAC configuration
      Then the component type "typescript" has extension ".ts"
      And the component type "typescript" has extension ".tsx"
