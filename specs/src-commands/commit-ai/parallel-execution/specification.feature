@deps:go @deps:claude @ov
Feature: src-commands_commit-ai_parallel-execution

  As a developer using commit-ai
  I want module sections to be generated concurrently
  So that multi-module commits complete faster

  Background:
    Given the commit-ai command is available
    And the repository has valid module contracts
    And AI provider is configured and available

  Rule: Module sections are generated concurrently for multi-module commits

    Scenario: Sequential execution baseline (reference implementation)
      Given I have staged files affecting 3 modules:
        | module       | files |
        | src-cli      | 2     |
        | src-core     | 2     |
        | src-commands | 2     |
      When I run commit-ai with sequential execution
      Then each module section is generated one after another
      And the total generation time is approximately the sum of individual times
      And the commit message contains sections for all 3 modules

    Scenario: Parallel execution reduces total time significantly
      Given I have staged files affecting 3 modules:
        | module       | files |
        | src-cli      | 2     |
        | src-core     | 2     |
        | src-commands | 2     |
      When I run commit-ai with parallel execution
      Then all module sections are generated concurrently
      And the total generation time is less than sequential execution
      And the speedup is at least 40 percent
      And the commit message contains sections for all 3 modules

    Scenario: Module section order is preserved despite concurrent execution
      Given I have staged files affecting modules in order:
        | module       |
        | src-cli      |
        | src-core     |
        | src-commands |
      When module sections are generated in parallel
      Then the commit message contains sections in the exact same order
      And section 1 contains "src-cli"
      And section 2 contains "src-core"
      And section 3 contains "src-commands"

    Scenario: Parallel execution scales to many modules
      Given I have staged files affecting 10 modules:
        | module   | files |
        | module-1 | 1     |
        | module-2 | 1     |
        | module-3 | 1     |
        | module-4 | 1     |
        | module-5 | 1     |
        | module-6 | 1     |
        | module-7 | 1     |
        | module-8 | 1     |
        | module-9 | 1     |
        | module-10| 1     |
      When I run commit-ai with parallel execution
      Then the speedup is at least 70 percent
      And all 10 module sections are present in the output
      And all sections are in the correct order

  Rule: Error handling works correctly in parallel execution

    Scenario: AI failure for one module is reported correctly
      Given I have staged files affecting 3 modules:
        | module       | files |
        | src-cli      | 2     |
        | src-core     | 2     |
        | src-commands | 2     |
      And the AI provider will fail for module "src-core"
      When I run commit-ai with parallel execution
      Then an error is reported mentioning "src-core"
      And the error message contains "failed to generate module sections"
      And the command exits with code 1

    Scenario: Multiple AI failures report the first error
      Given I have staged files affecting 5 modules
      And the AI provider will fail for modules:
        | module       |
        | module-2     |
        | module-4     |
      When I run commit-ai with parallel execution
      Then an error is reported for one of the failed modules
      And the command exits with code 1

  Rule: Thread-safe debug file writing

    Scenario: Debug mode writes files safely with concurrent execution
      Given I enable debug mode with --debug flag
      And I have staged files affecting 5 modules
      When module sections are generated in parallel
      Then debug files are created for all 5 modules:
        | file_pattern                    |
        | debug-module-1-*-context.md     |
        | debug-module-2-*-context.md     |
        | debug-module-3-*-context.md     |
        | debug-module-4-*-context.md     |
        | debug-module-5-*-context.md     |
        | debug-module-1-*-output.md      |
        | debug-module-2-*-output.md      |
        | debug-module-3-*-output.md      |
        | debug-module-4-*-output.md      |
        | debug-module-5-*-output.md      |
      And no file corruption occurs
      And no "concurrent write" errors are logged

    Scenario: Concurrent debug writes complete without data races
      Given I enable debug mode
      And I have staged files affecting 20 modules
      When I run commit-ai with race detector enabled
      Then no data race warnings are reported
      And all debug files are written successfully

  Rule: Single-module commits are unaffected

    Scenario: Single-module commit skips parallel execution
      Given I have staged files in only 1 module "src-cli"
      When I run commit-ai
      Then no module sections are generated
      And the output contains only the top-level summary
      And the generation completes successfully

  Rule: Behavioral compatibility with sequential implementation

    Scenario: Parallel execution produces identical output format
      Given I have staged files affecting 3 modules
      When I generate a commit message with parallel execution
      And I generate a commit message with sequential execution
      Then both messages have the same structure
      And both messages contain sections for all 3 modules
      And both messages pass contract validation

    Scenario: Empty module sections are filtered correctly
      Given I have staged files affecting 3 modules
      And the AI returns an empty response for module "src-core"
      When I run commit-ai with parallel execution
      Then the empty section for "src-core" is filtered out
      And sections for other modules are preserved
      And missing module stub is added for "src-core"

  Rule: Performance monitoring and metrics

    Scenario: Performance metrics are collected and reported
      Given I enable performance metrics with COMMIT_AI_PERF_METRICS=1
      And I have staged files affecting 5 modules
      When I run commit-ai with parallel execution
      Then performance metrics are logged to stderr
      And metrics include total time
      And metrics include parallel execution flag set to true
      And metrics include module count of 5
