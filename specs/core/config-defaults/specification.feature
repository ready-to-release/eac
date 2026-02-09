@L1 @ov @depm:core
Feature: Configuration Defaults System

  As a repository maintainer
  I want EAC to provide sensible defaults for all configuration
  So that projects work out of the box while allowing full customization

  The EAC configuration system loads defaults from contracts/core/0.1.0/defaults/*.yml
  and merges them with user config from .eac/*.yml. User values override defaults.

  Note: Modules are defined in repository.yml (not a separate repository.yml file).

  Background:
    Given I am in an isolated test repository

  # ===========================================================================
  # Category A: Defaults-Only Loading (No User Config)
  # ===========================================================================

  Rule: When user config is absent, all defaults are loaded

    @skip:broken
    Scenario: A1 - No configuration directory uses all defaults
      Given the repository has no ".eac" directory
      When I load the EAC configuration
      Then the modules config contains module "default"
      And the module "default" has component "go"
      And the module "default" has component root "go" as "."
      And the component types config contains type "go"
      And the component types config contains type "dockerfile"
      And the component types config contains type "typescript"
      And the component types config contains type "assets"
      And the repository paths.specs_root is "specs"
      And the repository paths.out.build is "out/build"
      And the system dependencies config contains "go"
      And the system dependencies config contains "docker"
      And the system dependencies config contains "git"

    @skip:broken
    Scenario: A2 - Empty configuration directory uses all defaults
      Given the repository has directory ".eac"
      When I load the EAC configuration
      Then the modules config contains module "default"
      And the repository paths.specs_root is "specs"

    @skip:broken
    Scenario: A3 - Missing modules in repository.yml uses default module
      Given the repository has file ".eac/component-types.yml" with:
        """
        component-types:
          custom-type:
            extensions: [".custom"]
        """
      When I load the EAC configuration
      Then the modules config contains module "default"
      And the component types config contains type "custom-type"
      And the component types config contains type "go"

    Scenario: A4 - Missing component-types.yml uses default types
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
            name: My Application
            components:
              go: go/myapp
        """
      When I load the EAC configuration
      Then the modules config contains module "myapp"
      And the component types config contains type "go"
      And the component types config contains type "dockerfile"

    Scenario: A5 - Repository with modules only uses default paths
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
            name: My Application
            components:
              go: go/myapp
        """
      When I load the EAC configuration
      Then the repository paths.specs_root is "specs"
      And the repository paths.out.build is "out/build"
      And the repository paths.out.test is "out/test"

    @skip:broken
    Scenario: A6 - Missing system-dependencies.yml uses default deps
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
            name: My Application
            components:
              go: go/myapp
        """
      When I load the EAC configuration
      Then the system dependencies config contains "go"
      And the system dependencies config contains "docker"
      And the dependency "go" has version ">=1.21"

  # ===========================================================================
  # Category B: User Config Only (Complete Override)
  # ===========================================================================

  Rule: User can provide complete config without needing defaults

    Scenario: B1 - User provides complete module list in repository.yml
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: app1
            name: Application One
            components:
              go: src/app1
          - moniker: app2
            name: Application Two
            components:
              go: src/app2
        """
      When I load the EAC configuration
      Then the modules config contains module "app1"
      And the modules config contains module "app2"
      And the modules config does not contain module "default"

    @skip:broken
    Scenario: B2 - User provides all config files
      Given the repository has file ".eac/repository.yml" with:
        """
        repository:
          type: poly
        paths:
          specs_root: features
          out:
            root: dist
            build: dist/build
        modules:
          - moniker: myapp
            name: My App
            components:
              custom: app
        """
      And the repository has file ".eac/component-types.yml" with:
        """
        component-types:
          custom:
            extensions: [".custom"]
        """
      And the repository has file ".eac/system-dependencies.yml" with:
        """
        dependencies:
          - moniker: custom-tool
            name: Custom Tool
            version: ">=1.0"
            verify:
              command: "custom-tool --version"
              pattern: "(\\d+\\.\\d+)"
        """
      When I load the EAC configuration
      Then the modules config contains module "myapp"
      And the component types config contains type "custom"
      And the repository paths.specs_root is "features"
      And the system dependencies config contains "custom-tool"

  # ===========================================================================
  # Category C: Component Types Merging
  # ===========================================================================

  Rule: User component types merge with defaults correctly

    Scenario: C1 - User adds new type alongside defaults
      Given the repository has file ".eac/blueprints.yml" with:
        """
        component-kinds:
          custom-go-lib:
            extensions: [".go"]
            builders: [go]
        """
      When I load the EAC configuration
      Then the component types config contains type "go"
      And the component types config contains type "custom-go-lib"
      And the type "custom-go-lib" has builder "go"

    Scenario: C2 - User overrides default type definition
      Given the repository has file ".eac/blueprints.yml" with:
        """
        component-kinds:
          go:
            extensions: [".go", ".go2"]
            builders: [go]
        """
      When I load the EAC configuration
      Then the component types config contains type "go"
      And the type "go" has extension ".go2"

    Scenario: C3 - User type with default file patterns
      Given the repository has file ".eac/blueprints.yml" with:
        """
        component-kinds:
          my-go-lib:
            extensions: [".go"]
            builders: [go]
            files:
              source: ["lib/**/*.go"]
              tests: ["lib/**/*_test.go"]
        """
      When I load the EAC configuration
      Then the component types config contains type "my-go-lib"
      And the type "my-go-lib" has default source pattern "lib/**/*.go"

    Scenario: C4 - Empty user types list preserves defaults
      Given the repository has file ".eac/blueprints.yml" with:
        """
        component-kinds: {}
        """
      When I load the EAC configuration
      Then the component types config contains type "go"
      And the component types config contains type "dockerfile"

  # ===========================================================================
  # Category D: Repository Paths Merging
  # ===========================================================================

  Rule: Repository paths merge at field level

    Scenario: D1 - Override specs_root only
      Given the repository has file ".eac/repository.yml" with:
        """
        paths:
          specs_root: features
        modules:
          - moniker: myapp
            name: My App
            components:
              go: app
        """
      When I load the EAC configuration
      Then the repository paths.specs_root is "features"
      And the repository paths.out.build is "out/build"

    Scenario: D2 - Override out.build only
      Given the repository has file ".eac/repository.yml" with:
        """
        paths:
          out:
            build: build/output
        modules:
          - moniker: myapp
            name: My App
            components:
              go: app
        """
      When I load the EAC configuration
      Then the repository paths.out.build is "build/output"
      And the repository paths.out.test is "out/test"

    Scenario: D3 - Override all out paths
      Given the repository has file ".eac/repository.yml" with:
        """
        paths:
          out:
            root: dist
            build: dist/build
            test: dist/test
            logs: dist/logs
        modules:
          - moniker: myapp
            name: My App
            components:
              go: app
        """
      When I load the EAC configuration
      Then the repository paths.out.root is "dist"
      And the repository paths.out.build is "dist/build"
      And the repository paths.out.test is "dist/test"
      And the repository paths.out.logs is "dist/logs"

  # ===========================================================================
  # Category E: Component Type Defaults Application to Modules
  # ===========================================================================

  Rule: Component-type-specific defaults are applied to modules

    Scenario: E1 - Module gets type default source patterns
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: mylib
            name: My Library
            components:
              go: go/mylib
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mylib" component "go" has source patterns containing "**/*.go"

    Scenario: E2 - User source patterns are not overridden
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: mylib
            name: My Library
            components:
              go:
                root: go/mylib
                patterns:
                  source: ["src/**/*.go"]
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mylib" component "go" has source patterns containing "src/**/*.go"
      And the module "mylib" component "go" does not have source pattern "**/*.go"

    Scenario: E3 - Dockerfile type gets default assets pattern
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: mycontainer
            name: My Container
            components:
              dockerfile: containers/mycontainer
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mycontainer" component "dockerfile" has source patterns containing "Dockerfile"

    Scenario: E4 - User changelog overrides default
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: mylib
            name: My Library
            versioning:
              scheme: SemVer
              changelog: HISTORY.md
            components:
              go: go/mylib
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mylib" has changelog "HISTORY.md"

    Scenario: E5 - Gherkin component uses specs_root path
      Given the repository has file ".eac/repository.yml" with:
        """
        repository:
          type: poly
        paths:
          specs_root: features
        modules:
          - moniker: mylib
            name: My Library
            components:
              go: go/mylib
              gherkin: features/mylib
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mylib" component "gherkin" has root "features/mylib"

    Scenario: E6 - Gherkin component default uses moniker in path
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: mylib
            name: My Library
            components:
              go: go/mylib
              gherkin: specs/mylib
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mylib" component "gherkin" has root "specs/mylib"

  # ===========================================================================
  # Category F: System Dependencies Merging
  # ===========================================================================

  Rule: System dependencies merge correctly

    @skip:broken
    Scenario: F1 - User adds new dependency
      Given the repository has file ".eac/system-dependencies.yml" with:
        """
        dependencies:
          - moniker: trivy
            name: Trivy Scanner
            version: ">=0.50"
            verify:
              command: "trivy --version"
              pattern: "Version: (\\d+\\.\\d+)"
        """
      When I load the EAC configuration
      Then the system dependencies config contains "go"
      And the system dependencies config contains "trivy"
      And the dependency "trivy" has version ">=0.50"

    @skip:broken
    Scenario: F2 - User overrides default dependency version
      Given the repository has file ".eac/system-dependencies.yml" with:
        """
        dependencies:
          - moniker: go
            name: Go
            version: ">=1.22"
            verify:
              command: "go version"
              pattern: "go version go(\\d+\\.\\d+)"
        """
      When I load the EAC configuration
      Then the dependency "go" has version ">=1.22"

    @skip:broken
    Scenario: F3 - Empty user dependencies preserves defaults
      Given the repository has file ".eac/system-dependencies.yml" with:
        """
        dependencies: []
        """
      When I load the EAC configuration
      Then the system dependencies config contains "go"
      And the system dependencies config contains "docker"

  # ===========================================================================
  # Category H: Edge Cases
  # ===========================================================================

  Rule: Edge cases are handled gracefully

    # NOTE: Schema requires at least 1 module (minItems: 1), so empty list is invalid.
    # This tests that schema validation correctly rejects empty modules.
    Scenario: H1 - Empty modules list in user config fails validation
      Given the repository has file ".eac/repository.yml" with:
        """
        modules: []
        """
      When I try to load the EAC configuration
      Then an error is returned containing "minItems"

    # NOTE: Module type concept removed - modules have components, not types

    Scenario: H3 - Module missing description defaults to name
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
            name: My Application
            components:
              go: go/myapp
        """
      When I load the EAC configuration
      Then the module "myapp" has description "My Application"

    Scenario: H4 - Invalid YAML returns error
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
          invalid yaml here
        """
      When I try to load the EAC configuration
      Then an error is returned containing "yaml"

    # Removed: H5 - Module type concept replaced by components

    Scenario: H6 - Very long module name is handled
      Given the repository has file ".eac/repository.yml" with:
        """
        modules:
          - moniker: this-is-a-very-long-module-moniker-name
            name: This Is A Very Long Module Name For Testing Purposes
            components:
              go: go/longname
        """
      When I load the EAC configuration
      Then the modules config contains module "this-is-a-very-long-module-moniker-name"
