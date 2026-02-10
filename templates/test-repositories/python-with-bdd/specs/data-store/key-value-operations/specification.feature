@L2 @deps:python @depm:data-store
Feature: Key-Value Store Operations
  As a developer
  I want a reliable key-value store
  So that I can persist and retrieve data

  Rule: Storing and retrieving values

    Scenario: Store and retrieve a value
      Given an empty data store
      When I set key "name" to value "Alice"
      Then getting key "name" should return "Alice"

    Scenario: Retrieve a non-existent key
      Given an empty data store
      When I get key "missing"
      Then the result should be empty

    Scenario: Overwrite an existing value
      Given an empty data store
      When I set key "name" to value "Alice"
      And I set key "name" to value "Bob"
      Then getting key "name" should return "Bob"

  Rule: Deleting values

    Scenario: Delete an existing key
      Given an empty data store
      And I set key "name" to value "Alice"
      When I delete key "name"
      Then getting key "name" should be empty
      And the store should have 0 entries

    Scenario: Delete a non-existent key
      Given an empty data store
      When I delete key "missing"
      Then the delete result should be false
