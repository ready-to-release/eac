@L0 @ov @control:sa-3
Feature: repository_no-build-tags-in-steps

  As a repository maintainer
  I want to ensure godog step files have no Go build tags
  So that BDD tests run without requiring special build flags

  Godog handles test filtering via Gherkin tags (@L0, @L2, @ov, etc.),
  not Go build tags. Adding build tags to godog_test.go files causes
  tests to silently skip unless the exact tags are passed to `go test`.

  Rule: Godog test files must not have build constraints

    Build tags like `//go:build L2 && ov` prevent test discovery
    unless explicitly specified. This leads to false positives where
    the test suite reports success but tests never actually ran.

    @L0 @ov
    Scenario: No godog_test.go files have build tags
      Given I discover all godog_test.go files in "go/eac/specs/impl"
      When I check each file for Go build tags
      Then no files should have "//go:build" directives
      And no files should have "// +build" directives
      And if any files have build tags, I should see the file path and the tag
