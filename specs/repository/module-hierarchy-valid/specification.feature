@L2 @verification @unisolated
Feature: Module Hierarchy Validation
  As a repository maintainer
  I want to ensure the module dependency graph is valid
  So that the build system can correctly order module builds and detect circular dependencies

  Background:
    Given the repository root exists

  @ov
  Scenario: Module dependencies and reverse dependencies are consistent
    Given I load all module contracts from the repository
    When I validate that all depends_on and used_by relationships are bidirectional
    Then all modules should have consistent dependency relationships
    And no module should reference a non-existent module

  @ov
  Scenario: Module dependency graph forms a valid directed acyclic graph (DAG)
    Given I load all module contracts from the repository
    When I build the complete dependency graph
    Then the graph should be a single connected component or forest
    And the graph should have no circular dependencies
    And all modules should be reachable from the root

  @ov
  Scenario: Every module in depends_on exists
    Given I load all module contracts from the repository
    When I check all depends_on references
    Then every referenced module should exist in the registry
    And I should see details of any missing modules

  @ov
  Scenario: Every module in used_by exists
    Given I load all module contracts from the repository
    When I check all used_by references
    Then every referenced module should exist in the registry
    And I should see details of any missing modules

  @ov
  Scenario: Bidirectional consistency between depends_on and used_by
    Given I load all module contracts from the repository
    When I validate bidirectional relationships
    Then if module A depends_on B, then B's used_by must include A
    And if module B has used_by A, then A's depends_on must include B
    And I should see details of any inconsistencies
