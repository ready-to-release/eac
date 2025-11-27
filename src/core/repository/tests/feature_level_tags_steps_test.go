package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

// featureLevelTagsContext holds state for feature-level tag validation
type featureLevelTagsContext struct {
	repoRoot      string
	featureFiles  []string
	validationErrors map[string][]string // file path -> errors
}

var featureLevelTagsCtx featureLevelTagsContext

func InitializeFeatureLevelTagsScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		featureLevelTagsCtx = featureLevelTagsContext{
			validationErrors: make(map[string][]string),
		}
		return ctx, nil
	})

	sc.Step(`^I discover all Gherkin feature files in the repository$`, iDiscoverAllGherkinFeatureFiles)
	sc.Step(`^I validate each feature file for conflicting L-level tags$`, iValidateEachFeatureFileForConflictingLLevelTags)
	sc.Step(`^no feature should have an L-tag when its scenarios have different L-tags$`, noFeatureShouldHaveAnLTagWhenItsScenariosHaveDifferentLTags)
	sc.Step(`^no feature should have a verification tag when its scenarios have different verification tags$`, noFeatureShouldHaveAVerificationTagWhenItsScenariosHaveDifferentVerificationTags)
	sc.Step(`^if any conflicts are found, I should see the file path, scenario name, and conflicting tags$`, ifAnyConflictsAreFoundIShouldSeeTheFilePathScenarioNameAndConflictingTags)
}

func iDiscoverAllGherkinFeatureFiles() error {
	// Get repository root
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Navigate to repository root (assuming we're in src/core/repository/tests)
	repoRoot := filepath.Join(cwd, "..", "..", "..", "..")
	featureLevelTagsCtx.repoRoot = repoRoot

	// Find all .feature files in specs/
	specsDir := filepath.Join(repoRoot, "specs")
	featureFiles := []string{}

	err = filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".feature") {
			featureFiles = append(featureFiles, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk specs directory: %w", err)
	}

	featureLevelTagsCtx.featureFiles = featureFiles

	if len(featureFiles) == 0 {
		return fmt.Errorf("no feature files found in %s", specsDir)
	}

	return nil
}

func iValidateEachFeatureFileForConflictingLLevelTags() error {
	for _, featureFile := range featureLevelTagsCtx.featureFiles {
		errors, err := coretesting.ValidateFeatureLevelTags(featureFile)
		if err != nil {
			return fmt.Errorf("failed to validate %s: %w", featureFile, err)
		}

		if len(errors) > 0 {
			featureLevelTagsCtx.validationErrors[featureFile] = errors
		}
	}

	return nil
}

func noFeatureShouldHaveAnLTagWhenItsScenariosHaveDifferentLTags() error {
	if len(featureLevelTagsCtx.validationErrors) > 0 {
		// Build error message
		var errorMessages []string
		for filePath, errors := range featureLevelTagsCtx.validationErrors {
			relPath, _ := filepath.Rel(featureLevelTagsCtx.repoRoot, filePath)
			errorMessages = append(errorMessages, fmt.Sprintf("\n%s:", relPath))
			for _, err := range errors {
				errorMessages = append(errorMessages, fmt.Sprintf("  - %s", err))
			}
		}

		return fmt.Errorf("found %d feature file(s) with conflicting L-level tags:%s",
			len(featureLevelTagsCtx.validationErrors),
			strings.Join(errorMessages, "\n"))
	}

	return nil
}

func noFeatureShouldHaveAVerificationTagWhenItsScenariosHaveDifferentVerificationTags() error {
	// This check is already performed by ValidateFeatureLevelTags() which checks both L-tags and verification tags
	// The validation errors are already collected in validationErrors map
	// This step is effectively the same as the L-tag check since they share the same validation logic
	return noFeatureShouldHaveAnLTagWhenItsScenariosHaveDifferentLTags()
}

func ifAnyConflictsAreFoundIShouldSeeTheFilePathScenarioNameAndConflictingTags() error {
	// This is a documentation step - the previous step already handles error reporting
	return nil
}

func resetFeatureLevelTagsContext() {
	featureLevelTagsCtx = featureLevelTagsContext{
		validationErrors: make(map[string][]string),
	}
}
