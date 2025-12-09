@L1 @ov @depm:eac-core
Feature: Configuration Defaults System

  As a repository maintainer
  I want EAC to provide sensible defaults for all configuration
  So that projects work out of the box while allowing full customization

  The EAC configuration system loads defaults from contracts/eac-core/0.1.0/defaults/*.yml
  and merges them with user config from .r2r/eac/*.yml. User values override defaults.

  Note: Modules are defined in repository.yml (not a separate modules.yml file).

  Background:
    Given I am in an isolated test repository

  # ===========================================================================
  # Category A: Defaults-Only Loading (No User Config)
  # ===========================================================================

  Rule: When user config is absent, all defaults are loaded

    Scenario: A1 - No configuration directory uses all defaults
      Given the repository has no ".r2r/eac" directory
      When I load the EAC configuration
      Then the modules config contains module "default"
      And the module "default" has type "go"
      And the module "default" has files root "."
      And the module types config contains type "go"
      And the module types config contains type "container"
      And the module types config contains type "typescript"
      And the module types config contains type "static"
      And the repository paths.specs_root is "specs"
      And the repository paths.test_impl_root is "tests"
      And the repository paths.out.build is "out/build"
      And the system dependencies config contains "go"
      And the system dependencies config contains "docker"
      And the system dependencies config contains "git"

    Scenario: A2 - Empty configuration directory uses all defaults
      Given the repository has directory ".r2r/eac"
      When I load the EAC configuration
      Then the modules config contains module "default"
      And the repository paths.specs_root is "specs"

    Scenario: A3 - Missing modules in repository.yml uses default module
      Given the repository has file ".r2r/eac/module-types.yml" with:
        """
        types:
          - name: custom-type
            description: Custom type
            capabilities: [custom]
        """
      When I load the EAC configuration
      Then the modules config contains module "default"
      And the module types config contains type "custom-type"
      And the module types config contains type "go"

    Scenario: A4 - Missing module-types.yml uses default types
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
            name: My Application
            type: go
            files:
              root: go/myapp
        """
      When I load the EAC configuration
      Then the modules config contains module "myapp"
      And the module types config contains type "go"
      And the module types config contains type "container"

    Scenario: A5 - Repository with modules only uses default paths
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
            name: My Application
            type: go
            files:
              root: go/myapp
        """
      When I load the EAC configuration
      Then the repository paths.specs_root is "specs"
      And the repository paths.out.build is "out/build"
      And the repository paths.out.test is "out/test"

    Scenario: A6 - Missing system-dependencies.yml uses default deps
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
            name: My Application
            type: go
            files:
              root: go/myapp
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
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: app1
            name: Application One
            type: go
            files:
              root: src/app1
          - moniker: app2
            name: Application Two
            type: go
            files:
              root: src/app2
        """
      When I load the EAC configuration
      Then the modules config contains module "app1"
      And the modules config contains module "app2"
      And the modules config does not contain module "default"

    Scenario: B2 - User provides all config files
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        repository:
          type: poly
        paths:
          specs_root: features
          test_impl_root: impl
          out:
            root: dist
            build: dist/build
        modules:
          - moniker: myapp
            name: My App
            type: custom
            files:
              root: app
        """
      And the repository has file ".r2r/eac/module-types.yml" with:
        """
        types:
          - name: custom
            description: Custom type
            capabilities: [custom_cap]
        """
      And the repository has file ".r2r/eac/system-dependencies.yml" with:
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
      And the module types config contains type "custom"
      And the repository paths.specs_root is "features"
      And the system dependencies config contains "custom-tool"

  # ===========================================================================
  # Category C: Module Types Merging
  # ===========================================================================

  Rule: User module types merge with defaults correctly

    Scenario: C1 - User adds new type alongside defaults
      Given the repository has file ".r2r/eac/module-types.yml" with:
        """
        types:
          - name: custom-go-lib
            description: Custom Go Library
            capabilities: [go_module, testable]
        """
      When I load the EAC configuration
      Then the module types config contains type "go"
      And the module types config contains type "custom-go-lib"
      And the type "go" has capability "go_module"
      And the type "custom-go-lib" has capability "testable"

    Scenario: C2 - User overrides default type definition
      Given the repository has file ".r2r/eac/module-types.yml" with:
        """
        types:
          - name: go
            description: Overridden Go type
            capabilities: [go_module, custom_cap]
        """
      When I load the EAC configuration
      Then the module types config contains type "go"
      And the type "go" has description "Overridden Go type"
      And the type "go" has capability "custom_cap"

    Scenario: C3 - User type with defaults block
      Given the repository has file ".r2r/eac/module-types.yml" with:
        """
        types:
          - name: my-go-lib
            description: My Go Library
            capabilities: [go_module]
            defaults:
              files:
                source: ["lib/**/*.go"]
                tests: ["lib/**/*_test.go"]
        """
      When I load the EAC configuration
      Then the module types config contains type "my-go-lib"
      And the type "my-go-lib" has default source pattern "lib/**/*.go"

    Scenario: C4 - Empty user types list preserves defaults
      Given the repository has file ".r2r/eac/module-types.yml" with:
        """
        types: []
        """
      When I load the EAC configuration
      Then the module types config contains type "go"
      And the module types config contains type "container"

  # ===========================================================================
  # Category D: Repository Paths Merging
  # ===========================================================================

  Rule: Repository paths merge at field level

    Scenario: D1 - Override specs_root only
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        paths:
          specs_root: features
        modules:
          - moniker: myapp
            name: My App
            type: go
            files:
              root: app
        """
      When I load the EAC configuration
      Then the repository paths.specs_root is "features"
      And the repository paths.test_impl_root is "tests"
      And the repository paths.out.build is "out/build"

    Scenario: D2 - Override test_impl_root only
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        paths:
          test_impl_root: test-implementations
        modules:
          - moniker: myapp
            name: My App
            type: go
            files:
              root: app
        """
      When I load the EAC configuration
      Then the repository paths.specs_root is "specs"
      And the repository paths.test_impl_root is "test-implementations"

    Scenario: D3 - Override out.build only
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        paths:
          out:
            build: build/output
        modules:
          - moniker: myapp
            name: My App
            type: go
            files:
              root: app
        """
      When I load the EAC configuration
      Then the repository paths.out.build is "build/output"
      And the repository paths.out.test is "out/test"

    Scenario: D4 - Override all out paths
      Given the repository has file ".r2r/eac/repository.yml" with:
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
            type: go
            files:
              root: app
        """
      When I load the EAC configuration
      Then the repository paths.out.root is "dist"
      And the repository paths.out.build is "dist/build"
      And the repository paths.out.test is "dist/test"
      And the repository paths.out.logs is "dist/logs"

  # ===========================================================================
  # Category E: Type Defaults Application to Modules
  # ===========================================================================

  Rule: Type-specific defaults are applied to modules

    Scenario: E1 - Module gets type default source patterns
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: mylib
            name: My Library
            type: go
            files:
              root: go/mylib
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mylib" has source patterns containing "**/*.go"

    Scenario: E2 - User source patterns are not overridden
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: mylib
            name: My Library
            type: go
            files:
              root: go/mylib
              source: ["src/**/*.go"]
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mylib" has source patterns containing "src/**/*.go"
      And the module "mylib" does not have source pattern "**/*.go"

    Scenario: E3 - Container type gets default assets pattern
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: mycontainer
            name: My Container
            type: container
            files:
              root: containers/mycontainer
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mycontainer" has assets patterns containing "Dockerfile"

    Scenario: E4 - User changelog overrides type default
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: mylib
            name: My Library
            type: go
            files:
              root: go/mylib
              changelog: HISTORY.md
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mylib" has changelog "HISTORY.md"

    Scenario: E5 - Type defaults resolve {specs_root} variable
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        repository:
          type: poly
        paths:
          specs_root: features
        modules:
          - moniker: mylib
            name: My Library
            type: go
            files:
              root: go/mylib
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mylib" specs pattern resolves with "features"

    Scenario: E6 - Type defaults resolve {moniker} variable
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: mylib
            name: My Library
            type: go
            files:
              root: go/mylib
        """
      When I load the EAC configuration
      And I apply type defaults to modules
      Then the module "mylib" test_impl path contains "mylib"

  # ===========================================================================
  # Category F: System Dependencies Merging
  # ===========================================================================

  Rule: System dependencies merge correctly

    Scenario: F1 - User adds new dependency
      Given the repository has file ".r2r/eac/system-dependencies.yml" with:
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

    Scenario: F2 - User overrides default dependency version
      Given the repository has file ".r2r/eac/system-dependencies.yml" with:
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

    Scenario: F3 - Empty user dependencies preserves defaults
      Given the repository has file ".r2r/eac/system-dependencies.yml" with:
        """
        dependencies: []
        """
      When I load the EAC configuration
      Then the system dependencies config contains "go"
      And the system dependencies config contains "docker"

  # ===========================================================================
  # Category G: Loading Precedence
  # ===========================================================================

  Rule: Filesystem defaults take precedence over embedded

    # NOTE: This scenario tests embedded defaults fallback which requires
    # the config loader to use go:embed for defaults. Currently not implemented.
    @skip @pending
    Scenario: G1 - Embedded defaults are used when filesystem unavailable
      Given the repository has no ".r2r/eac" directory
      And the contracts directory does not exist
      When I load the EAC configuration
      Then the modules config contains module "default"
      And the module types config contains type "go"

  # ===========================================================================
  # Category H: Edge Cases
  # ===========================================================================

  Rule: Edge cases are handled gracefully

    # NOTE: Schema requires at least 1 module (minItems: 1), so empty list is invalid.
    # This tests that schema validation correctly rejects empty modules.
    Scenario: H1 - Empty modules list in user config fails validation
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules: []
        """
      When I try to load the EAC configuration
      Then an error is returned containing "minItems"

    # NOTE: When type is not specified, applyDefaults() sets it to "no-module-type"
    # to clearly indicate the module type is not configured. This is intentional.
    Scenario: H2 - Module missing type field gets placeholder type
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
            name: My App
            files:
              root: go/myapp
        """
      When I load the EAC configuration
      Then the module "myapp" has type "no-module-type"

    Scenario: H3 - Module missing description defaults to name
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
            name: My Application
            type: go
            files:
              root: go/myapp
        """
      When I load the EAC configuration
      Then the module "myapp" has description "My Application"

    Scenario: H4 - Invalid YAML returns error
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
          invalid yaml here
        """
      When I try to load the EAC configuration
      Then an error is returned containing "yaml"

    Scenario: H5 - Module with unknown type loads without type defaults
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: myapp
            name: My App
            type: unknown-type
            files:
              root: app
        """
      When I load the EAC configuration
      Then the module "myapp" has type "unknown-type"
      And the module "myapp" has no source patterns from type defaults

    Scenario: H6 - Very long module name is handled
      Given the repository has file ".r2r/eac/repository.yml" with:
        """
        modules:
          - moniker: this-is-a-very-long-module-moniker-name
            name: This Is A Very Long Module Name For Testing Purposes
            type: go
            files:
              root: go/longname
        """
      When I load the EAC configuration
      Then the modules config contains module "this-is-a-very-long-module-moniker-name"
