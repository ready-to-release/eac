# Intent: Allow the build system to correctly order module builds and detect circular dependencies by ensuring the module dependency graph is valid
# Architecture: Affects repository module registry; validates bidirectional depends_on and used_by relationships; checks DAG acyclicity across all module contracts

@L1 @control:sa-3
Feature: Module Hierarchy Validation
  As a repository maintainer
  I want to ensure the module dependency graph is valid
  So that the build system can correctly order module builds and detect circular dependencies

  Background:
    Given the repository root exists

  @L1 @ov
  Scenario: Module dependencies and reverse dependencies are consistent
    Given I load all module contracts from the repository
    When I validate that all depends_on and used_by relationships are bidirectional
    Then all modules should have consistent dependency relationships
    And no module should reference a non-existent module

  @L1 @ov
  Scenario: Module dependency graph forms a valid directed acyclic graph (DAG)
    Given I load all module contracts from the repository
    When I build the complete dependency graph
    Then the graph should be a single connected component or forest
    And the graph should have no circular dependencies
    And all modules should be reachable from the root

  @L1 @ov
  Scenario: Every module in depends_on exists
    Given I load all module contracts from the repository
    When I check all depends_on references
    Then every referenced module should exist in the registry
    And I should see details of any missing modules

  @L1 @ov
  Scenario: Every module in used_by exists
    Given I load all module contracts from the repository
    When I check all used_by references
    Then every referenced module should exist in the registry
    And I should see details of any missing modules

  @L1 @ov
  Scenario: Bidirectional consistency between depends_on and used_by
    Given I load all module contracts from the repository
    When I validate bidirectional relationships
    Then if module A depends_on B, then B's used_by must include A
    And if module B has used_by A, then A's depends_on must include B
    And I should see details of any inconsistencies
