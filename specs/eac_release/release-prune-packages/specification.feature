@deps:go @env:mock-github @pending
Feature: eac-cli_release-prune-packages

  As a repository maintainer
  I want to prune old container image versions from GHCR
  So that storage costs are managed and the registry stays clean

  Background:
    Given I am in a git repository
    And the repository has registries configured for "ghcr.io"
    And the registry cleanup policy is enabled

  Rule: Safety - Released versions must never be deleted

    @L1 @ov
    Scenario: Versions with release tags are protected
      Given package "eac-ext" has versions:
        | digest      | tags                    | created_at  |
        | sha256:aaa1 | v1.0.0                  | 30 days ago |
        | sha256:bbb2 | sha-abc123              | 20 days ago |
        | sha256:ccc3 | sha-def456              | 10 days ago |
      And preserve_patterns includes "v*"
      When I run "release prune-packages eac-ext --keep 1"
      Then version "sha256:aaa1" should be protected
      And the protection reason should be "tag matches preserve pattern"

    @L1 @ov
    Scenario: Versions associated with GitHub releases are protected
      Given package "eac-ext" has versions:
        | digest      | tags                    | created_at  |
        | sha256:aaa1 | eac-ext/1.0.0           | 30 days ago |
        | sha256:bbb2 | sha-abc123              | 20 days ago |
      And GitHub release "eac-ext/1.0.0" exists
      When I run "release prune-packages eac-ext --keep 1"
      Then version "sha256:aaa1" should be protected
      And the protection reason should be "associated with GitHub release"

    @L1 @ov
    Scenario: Versions whose digest matches a released version are protected
      Given package "eac-ext" has versions:
        | digest      | tags                    | created_at  |
        | sha256:aaa1 | v1.0.0                  | 30 days ago |
        | sha256:aaa1 | sha-abc123              | 30 days ago |
      And preserve_patterns includes "v*"
      When I run "release prune-packages eac-ext --keep 0"
      Then version with tag "sha-abc123" should be protected
      And the protection reason should be "digest matches released version"

    @L1 @ov
    Scenario: Recent versions are protected by min_age_days
      Given package "eac-ext" has versions:
        | digest      | tags                    | created_at  |
        | sha256:aaa1 | sha-abc123              | 3 days ago  |
        | sha256:bbb2 | sha-def456              | 20 days ago |
      And min_age_days is 7
      When I run "release prune-packages eac-ext --keep 0"
      Then version "sha256:aaa1" should be protected
      And the protection reason should be "created less than min_age_days ago"

  Rule: Only versions matching prune patterns are candidates for deletion

    @L2 @ov
    Scenario: Only prune-pattern matching versions are candidates
      Given package "eac-ext" has versions:
        | digest      | tags                    | created_at  |
        | sha256:aaa1 | sha-abc123              | 30 days ago |
        | sha256:bbb2 | feature-branch          | 30 days ago |
        | sha256:ccc3 | dev-latest              | 30 days ago |
      And prune_patterns includes "sha-*" and "dev-*"
      When I run "release prune-packages eac-ext --keep 1"
      Then version "sha256:bbb2" should be protected
      And the protection reason should be "no tags match prune patterns"

    @L2 @ov
    Scenario: Untagged versions are eligible for pruning
      Given package "eac-ext" has versions:
        | digest      | tags                    | created_at  |
        | sha256:aaa1 |                         | 30 days ago |
        | sha256:bbb2 | sha-abc123              | 20 days ago |
      When I run "release prune-packages eac-ext --keep 1"
      Then version "sha256:aaa1" should be prunable

  Rule: Keep-latest-N policy retains newest versions

    @L2 @ov
    Scenario: Keeps newest N prunable versions
      Given package "eac-ext" has versions:
        | digest      | tags                    | created_at  |
        | sha256:aaa1 | sha-oldest              | 30 days ago |
        | sha256:bbb2 | sha-middle              | 20 days ago |
        | sha256:ccc3 | sha-newest              | 10 days ago |
      And prune_patterns includes "sha-*"
      When I run "release prune-packages eac-ext --keep 2"
      Then version "sha256:aaa1" should be marked for deletion
      And version "sha256:bbb2" should be kept
      And version "sha256:ccc3" should be kept

  Rule: Dry-run mode is the default for safety

    @L1 @ov
    Scenario: Default mode is dry-run
      Given package "eac-ext" has versions:
        | digest      | tags                    | created_at  |
        | sha256:aaa1 | sha-old                 | 30 days ago |
      When I run "release prune-packages eac-ext --keep 0"
      Then no versions should be deleted
      And the output should contain "DRY RUN"
      And the output should contain "would delete"

    @L1 @ov
    Scenario: Force flag required for actual deletion
      Given package "eac-ext" has versions:
        | digest      | tags                    | created_at  |
        | sha256:aaa1 | sha-old                 | 30 days ago |
      When I run "release prune-packages eac-ext --keep 0 --force"
      Then version "sha256:aaa1" should be deleted
      And the output should contain "Deleted: 1"

  Rule: Command requires package name or all flag

    @L2 @ov
    Scenario: Error when no package specified
      When I run "release prune-packages"
      Then the command should fail
      And the output should contain "package name required"
      And the output should contain "use --all"

    @L2 @ov
    Scenario: All packages processed with all flag
      Given packages exist: "eac-ext", "mkdocs-pdf"
      When I run "release prune-packages --all"
      Then both packages should be processed
