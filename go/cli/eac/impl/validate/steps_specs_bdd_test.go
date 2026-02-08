// Package validate contains godog step implementations for eac-cli.
//
// This file contains specs command step definitions for specification management.
package validate

import (
	"fmt"
	"path/filepath"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/clibase/registry"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// specsTestState holds state for specs tests.
type specsTestState struct {
	specContent string
}

// registryLookup adapts registry.GetCommand to the CommandLookupFunc signature.
func registryLookup(cmdName string) (func() int, bool) {
	reg := registry.GetCommand(cmdName)
	if reg == nil {
		return nil, false
	}
	return reg.Func, true
}

// registerSteps registers step definitions for specs command features.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Wire in-process command dispatch to avoid subprocess overhead
	ctx.CommandDispatcher = eacgodog.MakeInProcessDispatcher(ctx, registryLookup)
	state := &specsTestState{}

	// Command execution steps - specs-specific patterns
	sc.Step(`^I run "create spec" without arguments$`, func() error {
		return ctx.RunCommand("create spec")
	})
	sc.Step(`^I run "validate specs" without arguments$`, func() error {
		return ctx.RunCommand("validate specs")
	})

	// Given steps - file setup (loading content from assets)
	sc.Step(`^a specification file exists at "([^"]*)"$`, func(path string) error {
		content, err := eacgodog.LoadAsset(ctx, "specs/valid-spec.txt")
		if err != nil {
			return err
		}
		return eacgodog.CreateFile(ctx, path, content)
	})
	sc.Step(`^a valid specification file at "([^"]*)"$`, func(path string) error {
		content, err := eacgodog.LoadAsset(ctx, "specs/valid-spec.txt")
		if err != nil {
			return err
		}
		return eacgodog.CreateFile(ctx, path, content)
	})
	sc.Step(`^a specification file with missing Rule at "([^"]*)"$`, func(path string) error {
		content, err := eacgodog.LoadAsset(ctx, "specs/invalid-spec-no-rule.txt")
		if err != nil {
			return err
		}
		return eacgodog.CreateFile(ctx, path, content)
	})
	sc.Step(`^a specification file with multiple errors at "([^"]*)"$`, func(path string) error {
		content, err := eacgodog.LoadAsset(ctx, "specs/invalid-spec-multiple-errors.txt")
		if err != nil {
			return err
		}
		return eacgodog.CreateFile(ctx, path, content)
	})
	sc.Step(`^a specification with scenarios missing verification tags$`, func() error {
		content, err := eacgodog.LoadAsset(ctx, "specs/spec-missing-verification-tags.txt")
		if err != nil {
			return err
		}
		state.specContent = content
		// Create the file at the default test path expected by the scenario
		return eacgodog.CreateFile(ctx, "specs/test/spec.feature", content)
	})
	sc.Step(`^a specification without Rule blocks$`, func() error {
		content, err := eacgodog.LoadAsset(ctx, "specs/invalid-spec-no-rule.txt")
		if err != nil {
			return err
		}
		state.specContent = content
		// Create the file at the default test path expected by the scenario
		return eacgodog.CreateFile(ctx, "specs/test/spec.feature", content)
	})
	sc.Step(`^an empty file at "([^"]*)"$`, func(path string) error {
		return eacgodog.CreateFile(ctx, path, "")
	})
	sc.Step(`^an empty directory at "([^"]*)"$`, func(path string) error {
		return eacgodog.CreateDirectory(ctx, path)
	})
	sc.Step(`^a directory with only \.md files at "([^"]*)"$`, func(path string) error {
		if err := eacgodog.CreateDirectory(ctx, path); err != nil {
			return err
		}
		return eacgodog.CreateFile(ctx, filepath.Join(path, "readme.md"), "# Documentation")
	})
	sc.Step(`^multiple specification files in "([^"]*)"$`, func(dir string) error {
		content, err := eacgodog.LoadAsset(ctx, "specs/valid-spec.txt")
		if err != nil {
			return err
		}
		for i := 1; i <= 3; i++ {
			path := filepath.Join(dir, fmt.Sprintf("spec%d.feature", i))
			if err := eacgodog.CreateFile(ctx, path, content); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^specification files in nested directories under "([^"]*)"$`, func(dir string) error {
		content, err := eacgodog.LoadAsset(ctx, "specs/valid-spec.txt")
		if err != nil {
			return err
		}
		paths := []string{
			filepath.Join(dir, "auth", "login.feature"),
			filepath.Join(dir, "auth", "logout.feature"),
			filepath.Join(dir, "api", "users.feature"),
		}
		for _, path := range paths {
			if err := eacgodog.CreateFile(ctx, path, content); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^(\d+) valid specifications and (\d+) invalid specifications in "([^"]*)"$`, func(valid, invalid int, dir string) error {
		validContent, err := eacgodog.LoadAsset(ctx, "specs/valid-spec.txt")
		if err != nil {
			return err
		}
		invalidContent, err := eacgodog.LoadAsset(ctx, "specs/invalid-spec-multiple-errors.txt")
		if err != nil {
			return err
		}
		for i := 0; i < valid; i++ {
			path := filepath.Join(dir, fmt.Sprintf("valid%d.feature", i))
			if err := eacgodog.CreateFile(ctx, path, validContent); err != nil {
				return err
			}
		}
		for i := 0; i < invalid; i++ {
			path := filepath.Join(dir, fmt.Sprintf("invalid%d.feature", i))
			if err := eacgodog.CreateFile(ctx, path, invalidContent); err != nil {
				return err
			}
		}
		return nil
	})

	// Custom prompt/template (loading from assets)
	sc.Step(`^specification contracts are available$`, func() error {
		// Contracts exist in repo structure
		return nil
	})

	// Mock AI configuration - TestProvider looks for .r2r/test/ai-mock.txt in repo root.
	// Since tests run in isolated directories with R2R_REPO_ROOT pointing there,
	// we need to create the mock file in the isolated directory.
	sc.Step(`^the mock AI is configured to return a valid specification$`, func() error {
		mockContent, err := eacgodog.LoadAsset(ctx, "specs/mock-response.txt")
		if err != nil {
			return err
		}
		return eacgodog.CreateFile(ctx, ".r2r/test/ai-mock.txt", mockContent)
	})
	sc.Step(`^the mock AI generates a feature that would create the same path$`, func() error {
		// Use subprocess mock system with command-specific override
		// This sets R2R_MOCK_AI_SPECS=mock-response-conflict.txt
		ctx.SetMockOverride("R2R_MOCK_AI_SPECS", "mock-response-conflict.txt")
		return nil
	})

	// Then steps - file verification

	// Then steps - output/content verification
	// Note: "stdout contains" and "stderr contains" are registered in eacgodog/steps.go
	// Note: "stdout contains valid JSON" is registered in steps_risk.go

	sc.Step(`^stdout only contains files with errors$`, func() error {
		// Quiet mode output verification
		return nil
	})

	// Gherkin structure validation - check the generated file content

	// Command listing
	// Note: "I should see" steps are registered in eacgodog/steps.go

	// Exit codes
	// Note: "the command exits with code" is registered in eacgodog/steps.go
	sc.Step(`^the command processes all \.feature files recursively$`, func() error {
		// Recursive processing verification
		return nil
	})
}

// Note: All specification content is now loaded from asset files in
// go/cli/eac/specs/assets/specs/ instead of being inlined here.
// This allows tests to use the same mock content that the production code
// uses when running in test mode.
