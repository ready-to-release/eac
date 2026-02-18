# Intent: Ensure the repository maintains a valid architecture with clear module boundaries and dependencies by validating module contracts are well-formed and consistent
# Architecture: Affects go/core module contract validation; validates JSON schema, dependency resolution, DAG acyclicity, file ownership, type defaults, and bidirectional relationships across all module contracts in .eac/repository.yml

@skip:broken
Feature: core_validate-module-contracts

As a repository maintainer
I want to validate module contracts are well-formed and consistent
So that the repository maintains a valid architecture with clear module boundaries and dependencies

Background:
  Given a repository with module contracts defined in .eac/repository.yml
  And component type definitions in the blueprints contract

Rule: Module contract fields must meet schema requirements

@ov @L1
Scenario: Module contract schema is valid
  Given a module contract with moniker, name, and files.root
  When the module contract is validated against the JSON schema
  Then the validation passes

@ov @L1
Scenario: Module contract moniker must follow naming pattern
  Given a module contract with an invalid moniker containing uppercase letters
  When the module contract is validated
  Then the validation fails with a pattern error

@ov @L1
Scenario: Module contract requires moniker and name fields
  Given a module contract missing the required moniker field
  When the module contract is validated
  Then the validation fails with a required field error

Rule: Module dependencies must be resolvable

@ov @L1
Scenario: All declared dependencies exist in the registry
  Given a module that depends on other modules
  When the module registry is validated
  Then each dependency is verified to exist in the registry

@ov @L1
Scenario: Invalid dependency reference is detected
  Given a module that depends on a non-existent module
  When the module registry is validated
  Then the validation fails with a dependency error

@ov @L1
Scenario: Circular dependencies are detected
  Given two modules that depend on each other
  When the module registry is validated
  Then the validation fails with a circular dependency error

Rule: Module dependency graph must be acyclic

@ov @L1
Scenario: Valid acyclic dependency graph is accepted
  Given modules organized in a valid dependency hierarchy
  When the module hierarchy is validated
  Then all modules are identified as properly ordered

@ov @L1
Scenario: Dependency graph forms a valid DAG
  Given multiple modules with complex dependency chains
  When the module hierarchy is validated
  Then the resulting graph is confirmed to be acyclic

Rule: File ownership must be clearly defined and non-overlapping

@ov @L1
Scenario: Each file belongs to exactly one module
  Given modules with non-overlapping file patterns
  When module file ownership is validated
  Then each file is assigned to exactly one module

@ov @L1
Scenario: Multi-module file ownership is rejected
  Given a file pattern that matches multiple modules
  When module file ownership is validated
  Then the validation fails with an ownership conflict error

@ov @L1
Scenario: No orphaned files exist in the repository
  Given a repository where all files are assigned to modules
  When module file ownership is validated
  Then no orphaned files are reported

Rule: Module type defaults must be properly applied

@ov @L1
Scenario: Default values from module type are applied to module contract
  Given a module with type "go" and missing optional fields
  When the module contract is processed with type defaults
  Then default values from the module type are applied
  And explicit values are not overwritten

@ov @L1
Scenario: Description defaults to name when not specified
  Given a module contract without an explicit description
  When the module contract is processed
  Then the description field is set to the module name

@ov @L1
Scenario: Module type-specific defaults are applied
  Given a module with type "go"
  When the module contract is processed
  Then build dependencies from the "go" type are available

Rule: Module contracts must reference valid module types

@ov @L1
Scenario: Valid module type reference is accepted
  Given a module with type "go" from the module type registry
  When the module contract is validated
  Then the module type reference is confirmed as valid

@ov @L1
Scenario: Invalid module type reference is rejected
  Given a module with a non-existent module type
  When the module contract is validated
  Then the validation fails with a module type error

Rule: Bidirectional module relationships must be consistent

@ov @L1
Scenario: Reverse dependencies are computed correctly
  Given a module that depends on another module
  When the module registry is processed
  Then the dependent module appears in the reverse dependency list of the dependency

@ov @L1
Scenario: Bidirectional relationships remain consistent
  Given multiple modules with interdependencies
  When the module registry is validated
  Then bidirectional relationships are confirmed as consistent

Rule: Module file patterns must be valid glob patterns

@ov @L1
Scenario: Valid glob pattern in file configuration is accepted
  Given a module with file patterns like "**/*.go" and "src/**"
  When the module contract is validated
  Then the glob patterns are confirmed as valid

@ov @L1
Scenario: File patterns support variable substitution
  Given a module with file patterns containing {moniker}, {root}, and {type}
  When the module contract is processed
  Then pattern variables are substituted with actual values

Rule: Module changelog file must exist or use default

@ov @L1
Scenario: Changelog file path is set to default CHANGELOG.md
  Given a module contract without explicit changelog file path
  When the module contract is processed
  Then the changelog file defaults to CHANGELOG.md in the module root

@ov @L1
Scenario: Explicit changelog file path is preserved
  Given a module contract with explicit changelog file path
  When the module contract is processed
  Then the explicit changelog path is not overwritten

Rule: All required module boundaries must be enforced

@ov @L1
Scenario: Module files.root is required for file ownership
  Given a module contract without files.root specified
  When the module contract is validated
  Then the validation fails with a required field error

@ov @L1
Scenario: Module specification repository paths are optional
  Given a module contract without repo.specs defined
  When the module contract is processed
  Then the default spec path pattern is applied