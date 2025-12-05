@L1 @ov @depm:eac-core
Feature: eac-core_handlers-config

  As a developer
  I want to configure build and test handlers through YAML
  So that I can customize how different module types are built and tested

  Background:
    Given the EAC config is loaded

  # =============================================================================
  # Rule: Handler Loading and Validation
  # =============================================================================

  Rule: Handler configuration must be loadable and valid

    Scenario: Load handlers from default location
      When I load the handlers configuration
      Then the handlers should be loaded successfully
      And I should have at least 4 handlers registered

    Scenario: All defined handlers are valid
      When I load the handlers configuration
      Then each handler should have a valid type
      And each handler should have a valid name

    Scenario: Handler types are correctly identified
      When I load the handlers configuration
      Then the "go" handler should have type "builtin"
      And the "docker" handler should have type "builtin"
      And the "mkdocs" handler should have type "docker"
      And the "npm" handler should have type "command"
      And the "none" handler should have type "builtin"

  # =============================================================================
  # Rule: Handler Lookup by Name
  # =============================================================================

  Rule: Handlers must be retrievable by name

    Scenario: Get existing handler by name
      Given the handlers are loaded
      When I get the handler named "go"
      Then I should receive the handler
      And the handler name should be "go"

    Scenario: Get unknown handler returns nil
      Given the handlers are loaded
      When I get the handler named "unknown-handler"
      Then I should receive nil

    Scenario: Get handler with empty name returns nil
      Given the handlers are loaded
      When I get the handler named ""
      Then I should receive nil

    Scenario Outline: Get all defined handlers by name
      Given the handlers are loaded
      When I get the handler named "<handler_name>"
      Then I should receive the handler
      And the handler description should not be empty

      Examples:
        | handler_name |
        | go           |
        | docker       |
        | mkdocs       |
        | npm          |
        | none         |

  # =============================================================================
  # Rule: Dispatch Rules - Build Handler Selection
  # =============================================================================

  Rule: Build handlers must be selected using dispatch rules

    Scenario: Go module uses go build handler
      Given the handlers are loaded
      When I get the build handler for module type "go-library" with capabilities "go_module" and build dep "go"
      Then the resolved handler should be "go"

    Scenario: Docker module uses docker build handler
      Given the handlers are loaded
      When I get the build handler for module type "r2r-extension" with capabilities "container" and build dep "docker"
      Then the resolved handler should be "docker"

    Scenario: MkDocs site uses mkdocs build handler via dispatch rule
      Given the handlers are loaded
      When I get the build handler for module type "mkdocs-site" with capabilities "documentation,serveable,container" and build dep "docker"
      Then the resolved handler should be "mkdocs"

    Scenario: NPM module uses npm build handler
      Given the handlers are loaded
      When I get the build handler for module type "vscode-ext" with capabilities "npm_package,typescript" and build dep "npm"
      Then the resolved handler should be "npm"

    Scenario: Scripts package uses scripts handler
      Given the handlers are loaded
      When I get the build handler for module type "scripts-package" with capabilities "" and build dep ""
      Then the resolved handler should be "scripts"

    Scenario: Primary build dep placeholder is resolved
      Given the handlers are loaded
      When I get the build handler for module type "custom-type" with capabilities "" and build dep "go"
      Then the resolved handler should be "go"

  # =============================================================================
  # Rule: Dispatch Rules - Test Handler Selection
  # =============================================================================

  Rule: Test handlers must be selected using dispatch rules

    Scenario: Go module uses go test handler
      Given the handlers are loaded
      When I get the test handler for module type "go-library" with capabilities "go_module" and build dep "go"
      Then the resolved handler should be "go"

    Scenario: NPM module uses npm test handler
      Given the handlers are loaded
      When I get the test handler for module type "vscode-ext" with capabilities "npm_package" and build dep "npm"
      Then the resolved handler should be "npm"

    Scenario: Static module with no build dep uses empty test handler
      Given the handlers are loaded
      When I get the test handler for module type "configuration" with capabilities "" and build dep ""
      Then the resolved handler should be ""

  # =============================================================================
  # Rule: Match Condition Evaluation
  # =============================================================================

  Rule: Dispatch rules must evaluate match conditions correctly

    Scenario: Default match condition always matches
      Given a dispatch rule with default match
      When I evaluate the rule for any module type
      Then the rule should match

    Scenario: Type match condition filters by module type
      Given a dispatch rule with type match "go-cli"
      When I evaluate the rule for module type "go-cli"
      Then the rule should match
      When I evaluate the rule for module type "go-library"
      Then the rule should not match

    Scenario: Build dep match condition filters by primary build dependency
      Given a dispatch rule with build dep match "go"
      When I evaluate the rule with build dep "go"
      Then the rule should match
      When I evaluate the rule with build dep "npm"
      Then the rule should not match

    Scenario: Capabilities match requires all specified capabilities
      Given a dispatch rule with capabilities match "documentation,container"
      When I evaluate the rule with capabilities "documentation,container,serveable"
      Then the rule should match
      When I evaluate the rule with capabilities "documentation"
      Then the rule should not match
      When I evaluate the rule with capabilities "container"
      Then the rule should not match

    Scenario: Combined match with type and capabilities
      Given a dispatch rule with type "mkdocs-site" and capabilities "documentation"
      When I evaluate the rule for module type "mkdocs-site" with capabilities "documentation,container"
      Then the rule should match
      When I evaluate the rule for module type "other-type" with capabilities "documentation,container"
      Then the rule should not match
      When I evaluate the rule for module type "mkdocs-site" with capabilities "container"
      Then the rule should not match

  # =============================================================================
  # Rule: Go Handler Configuration
  # =============================================================================

  Rule: Go handler must provide cross-compile targets configuration

    Scenario: Cross-compile targets are loaded from config
      Given the handlers are loaded
      When I get the cross-compile targets
      Then I should have at least 5 targets
      And I should have target "linux/amd64"
      And I should have target "linux/arm64"
      And I should have target "darwin/amd64"
      And I should have target "darwin/arm64"
      And I should have target "windows/amd64"

    Scenario: Windows targets have .exe suffix
      Given the handlers are loaded
      When I get the cross-compile targets
      Then the "windows/amd64" target should have suffix ".exe"

    Scenario: Non-Windows targets have no suffix
      Given the handlers are loaded
      When I get the cross-compile targets
      Then the "linux/amd64" target should have suffix ""
      And the "darwin/amd64" target should have suffix ""

    Scenario: UPX supported platforms are configured
      Given the handlers are loaded
      When I get the UPX platforms
      Then UPX should be supported for "linux"
      And UPX should be supported for "windows"
      And UPX should not be supported for "darwin"

    Scenario: Default cross-compile targets when config is missing
      Given an empty handlers configuration
      When I get the cross-compile targets
      Then I should have at least 5 targets

  # =============================================================================
  # Rule: Docker Handler Configuration
  # =============================================================================

  Rule: Docker handler must provide dockerfile path configuration

    Scenario: Dockerfile paths are loaded from config
      Given the handlers are loaded
      When I get the dockerfile paths
      Then I should have at least 2 paths
      And I should have path pattern "containers/{moniker}/Dockerfile"
      And I should have path pattern "{root}/Dockerfile"

    Scenario: Dockerfile path placeholders are resolved
      Given the handlers are loaded
      When I resolve dockerfile path "containers/{moniker}/Dockerfile" for moniker "ext-eac" with root "containers/ext-eac"
      Then the resolved path should be "containers/ext-eac/Dockerfile"
      When I resolve dockerfile path "{root}/Dockerfile" for moniker "ext-eac" with root "containers/ext-eac"
      Then the resolved path should be "containers/ext-eac/Dockerfile"

    Scenario: CI platforms are configured
      Given the handlers are loaded
      When I get the CI platforms
      Then I should have platform "linux/amd64"
      And I should have platform "linux/arm64"

    Scenario: CI platforms string format
      Given the handlers are loaded
      When I get the CI platforms string
      Then the result should contain "linux/amd64"
      And the result should contain ","

  # =============================================================================
  # Rule: MkDocs Handler Configuration
  # =============================================================================

  Rule: MkDocs handler must provide docker-based configuration

    Scenario: MkDocs handler has docker image configuration
      Given the handlers are loaded
      When I get the mkdocs handler
      Then the handler type should be "docker"
      And the handler should have a build image configured
      And the handler should have a build command configured

    Scenario: MkDocs handler has volume mount configuration
      Given the handlers are loaded
      When I get the mkdocs handler build config
      Then the config should have volumes defined

  # =============================================================================
  # Rule: NPM Handler Configuration
  # =============================================================================

  Rule: NPM handler must provide command-based configuration

    Scenario: NPM handler has build steps
      Given the handlers are loaded
      When I get the npm handler
      Then the handler type should be "command"
      And the handler should have build steps configured

    Scenario: NPM build steps have conditional execution
      Given the handlers are loaded
      When I get the npm handler build steps
      Then the "compile" step should have a when condition

    Scenario: NPM test steps have conditional execution
      Given the handlers are loaded
      When I get the npm handler test steps
      Then the "test" step should have a when condition

  # =============================================================================
  # Rule: Nil Config Safety
  # =============================================================================

  Rule: Handler accessors must be safe with nil configuration

    Scenario: Get handler from nil config returns nil
      Given a nil handlers configuration
      When I get the handler named "go"
      Then I should receive nil

    Scenario: Get build handler from nil config returns fallback
      Given a nil handlers configuration
      When I get the build handler for module type "test" with capabilities "" and build dep "go"
      Then the resolved handler should be "go"

    Scenario: Get cross-compile targets from nil config returns defaults
      Given a nil handlers configuration
      When I get the cross-compile targets
      Then I should have at least 5 targets

    Scenario: Get UPX platforms from nil config returns defaults
      Given a nil handlers configuration
      When I get the UPX platforms
      Then UPX should be supported for "linux"

    Scenario: Get dockerfile paths from nil config returns defaults
      Given a nil handlers configuration
      When I get the dockerfile paths
      Then I should have at least 2 paths

    Scenario: Get CI platforms from nil config returns defaults
      Given a nil handlers configuration
      When I get the CI platforms
      Then I should have platform "linux/amd64"

  # =============================================================================
  # Rule: Handler Validation
  # =============================================================================

  Rule: Handlers must validate correctly

    Scenario Outline: Valid handler types pass validation
      Given a handler with name "test" and type "<type>"
      When I validate the handler
      Then the validation should succeed

      Examples:
        | type    |
        | builtin |
        | command |
        | script  |
        | docker  |

    Scenario: Handler with missing name fails validation
      Given a handler with name "" and type "builtin"
      When I validate the handler
      Then the validation should fail
      And the error should mention "name"

    Scenario: Handler with invalid type fails validation
      Given a handler with name "test" and type "invalid"
      When I validate the handler
      Then the validation should fail
      And the error should mention "invalid type"

  # =============================================================================
  # Rule: Integration - Full Dispatch Scenarios
  # =============================================================================

  Rule: Full dispatch scenarios must work end-to-end

    Scenario Outline: Build handler dispatch for all module types
      Given the handlers are loaded with full config
      When I get the build handler for module type "<module_type>" with capabilities "<capabilities>" and build dep "<build_dep>"
      Then the resolved handler should be "<expected_handler>"

      Examples:
        | module_type     | capabilities                        | build_dep | expected_handler |
        | go-cli          | go_module,executable,cross_compile  | go        | go               |
        | go-library      | go_module                           | go        | go               |
        | r2r-extension   | container                           | docker    | docker           |
        | mkdocs-site     | documentation,serveable,container   | docker    | mkdocs           |
        | vscode-ext      | npm_package,typescript              | npm       | npm              |
        | scripts-package |                                     |           | scripts          |
        | configuration   |                                     |           |                  |

    Scenario Outline: Test handler dispatch for all module types
      Given the handlers are loaded with full config
      When I get the test handler for module type "<module_type>" with capabilities "<capabilities>" and build dep "<build_dep>"
      Then the resolved handler should be "<expected_handler>"

      Examples:
        | module_type     | capabilities       | build_dep | expected_handler |
        | go-cli          | go_module          | go        | go               |
        | go-library      | go_module          | go        | go               |
        | mkdocs-site     | documentation      | docker    | docker           |
        | vscode-ext      | npm_package        | npm       | npm              |
        | scripts-package |                    |           |                  |

  # =============================================================================
  # Rule: Custom Handler Configurations
  # =============================================================================

  Rule: Custom handler configurations can be created and loaded

    Scenario: Create custom builtin handler configuration
      Given a custom handlers YAML with a builtin handler "custom-go"
      When I parse the handlers configuration
      Then the "custom-go" handler should exist
      And the "custom-go" handler should have type "builtin"

    Scenario: Create custom command handler configuration
      Given a custom handlers YAML with a command handler "custom-npm"
      When I parse the handlers configuration
      Then the "custom-npm" handler should exist
      And the "custom-npm" handler should have type "command"
      And the "custom-npm" handler should have build steps

    Scenario: Create custom docker handler configuration
      Given a custom handlers YAML with a docker handler "custom-docker"
      When I parse the handlers configuration
      Then the "custom-docker" handler should exist
      And the "custom-docker" handler should have type "docker"

    Scenario: Create dispatch rules for custom types
      Given a custom handlers YAML with dispatch rules:
        """
        dispatch:
          build:
            - match:
                type: custom-module
              handler: custom-handler
            - match:
                default: true
              handler: "{primary_build_dep}"
        """
      When I parse the handlers configuration
      And I get the build handler for module type "custom-module" with capabilities "" and build dep "go"
      Then the resolved handler should be "custom-handler"

    Scenario: Create dispatch rules with capability matching
      Given a custom handlers YAML with dispatch rules:
        """
        dispatch:
          build:
            - match:
                capabilities: [custom_capability]
              handler: special-handler
            - match:
                default: true
              handler: "{primary_build_dep}"
        """
      When I parse the handlers configuration
      And I get the build handler for module type "any" with capabilities "custom_capability,other" and build dep "go"
      Then the resolved handler should be "special-handler"

  # =============================================================================
  # Rule: Handler Flags Configuration
  # =============================================================================

  Rule: Handler flags must be configurable and retrievable

    Scenario: Go handler has build flags defined
      Given the handlers are loaded
      When I get the build flags for handler "go"
      Then I should have at least 3 build flags

    Scenario: Get specific flag by name
      Given the handlers are loaded
      When I get the build flag "tidy" for handler "go"
      Then the flag should exist
      And the flag type should be "bool"
      And the flag CLI positive should be "--tidy-first"
      And the flag CLI negative should be "--no-tidy"

    Scenario: Get flag by CLI positive form
      Given the handlers are loaded
      When I get the build flag by CLI "--tidy-first" for handler "go"
      Then the flag should exist
      And the flag name should be "tidy"

    Scenario: Get flag by CLI negative form
      Given the handlers are loaded
      When I get the build flag by CLI "--no-tidy" for handler "go"
      Then the flag should exist
      And the flag name should be "tidy"

    Scenario: Get flag by value flag form
      Given the handlers are loaded
      When I get the build flag by CLI "--version" for handler "go"
      Then the flag should exist
      And the flag name should be "version"
      And the flag type should be "string"

    Scenario: Unknown flag by name returns nil
      Given the handlers are loaded
      When I get the build flag "unknown-flag" for handler "go"
      Then the flag should not exist

    Scenario: Unknown handler for flag returns nil
      Given the handlers are loaded
      When I get the build flag "tidy" for handler "unknown-handler"
      Then the flag should not exist

  # =============================================================================
  # Rule: Handler Flag Defaults
  # =============================================================================

  Rule: Handler flags must have appropriate defaults

    Scenario: Tidy flag defaults to true for local builds
      Given the handlers are loaded
      When I get the build flag "tidy" for handler "go"
      Then the flag bool default for local should be true

    Scenario: Tidy flag defaults to false for CI
      Given the handlers are loaded
      When I get the build flag "tidy" for handler "go"
      Then the flag bool default for CI should be false

    Scenario: Compressed flag defaults to false
      Given the handlers are loaded
      When I get the build flag "compressed" for handler "go"
      Then the flag bool default for local should be false
      And the flag bool default for CI should be false

    Scenario: Version flag defaults to empty string
      Given the handlers are loaded
      When I get the build flag "version" for handler "go"
      Then the flag string default should be ""

  # =============================================================================
  # Rule: Get All CLI Flags
  # =============================================================================

  Rule: All CLI flags must be retrievable as a map

    Scenario: Get all CLI flags for go handler
      Given the handlers are loaded
      When I get all build CLI flags for handler "go"
      Then the CLI flags map should contain "--tidy-first"
      And the CLI flags map should contain "--no-tidy"
      And the CLI flags map should contain "--compressed"
      And the CLI flags map should contain "--compressed-upx"
      And the CLI flags map should contain "--version"

    Scenario: Get all CLI flags for handler without flags
      Given the handlers are loaded
      When I get all build CLI flags for handler "docker"
      Then the CLI flags map should be empty

  # =============================================================================
  # Rule: Handler Flag Validation
  # =============================================================================

  Rule: Handler flags must validate correctly

    Scenario: Valid bool flag with positive and negative
      Given a handler flag with name "test", type "bool", cli_positive "--test", cli_negative "--no-test"
      When I validate the flag
      Then the flag validation should succeed

    Scenario: Valid bool flag with only positive
      Given a handler flag with name "test", type "bool", cli_positive "--test"
      When I validate the flag
      Then the flag validation should succeed

    Scenario: Valid string flag with value_flag
      Given a handler flag with name "version", type "string", value_flag "--version"
      When I validate the flag
      Then the flag validation should succeed

    Scenario: Bool flag without cli_positive fails
      Given a handler flag with name "test", type "bool"
      When I validate the flag
      Then the flag validation should fail
      And the flag error should mention "cli_positive"

    Scenario: String flag without value_flag or cli_positive fails
      Given a handler flag with name "test", type "string"
      When I validate the flag
      Then the flag validation should fail

    Scenario: Flag with missing name fails
      Given a handler flag with type "bool", cli_positive "--test"
      When I validate the flag
      Then the flag validation should fail
      And the flag error should mention "name"

    Scenario: Flag with invalid type fails
      Given a handler flag with name "test", type "invalid", cli_positive "--test"
      When I validate the flag
      Then the flag validation should fail
      And the flag error should mention "invalid type"

  # =============================================================================
  # Rule: Handler Flags Nil Safety
  # =============================================================================

  Rule: Handler flags methods must be nil-safe

    Scenario: Get build flags from nil config
      Given a nil handlers configuration
      When I get the build flags for handler "go"
      Then I should have no build flags

    Scenario: Get flag default from nil flag
      Given a nil handler flag
      When I get the flag default for local
      Then the result should be nil
