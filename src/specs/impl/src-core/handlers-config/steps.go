// Package handlersconfig contains godog step implementations for specs/src-core/handlers-config.
//
// This file contains step definitions for testing the handlers configuration system.
package handlersconfig

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/config"
	"github.com/ready-to-release/eac/src/specs/internal"
	"gopkg.in/yaml.v3"
)

// testContext holds state for handlers configuration scenarios
type testContext struct {
	// Configuration state
	cfg             *config.EACConfig
	handlers        *config.HandlersConfig
	customHandlers  *config.HandlersConfig
	loadError       error
	validationError error

	// Handler retrieval state
	handler         *config.Handler
	handlerName     string
	resolvedHandler string

	// Targets and platforms
	crossCompileTargets []config.CrossCompileTarget
	upxPlatforms        []string
	dockerfilePaths     []string
	ciPlatforms         []string
	ciPlatformsString   string

	// Dispatch rule state
	dispatchRule config.DispatchRule
	ruleMatches  bool

	// Path resolution state
	resolvedPath string

	// Custom YAML state
	customYAML string

	// Handler flags state
	buildFlags       []config.HandlerFlag
	currentFlag      *config.HandlerFlag
	flagDefault      interface{}
	cliFlagsMap      map[string]*config.HandlerFlag
	flagValidateErr  error
}

var hcCtx *testContext

// RegisterSteps registers step definitions for handlers config feature specs.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Initialize context at start of each scenario
	sc.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		hcCtx = &testContext{}
		return c, nil
	})

	// Background steps
	sc.Step(`^the EAC config is loaded$`, theEACConfigIsLoaded)

	// Loading steps
	sc.Step(`^I load the handlers configuration$`, iLoadTheHandlersConfiguration)
	sc.Step(`^the handlers should be loaded successfully$`, theHandlersShouldBeLoadedSuccessfully)
	sc.Step(`^I should have at least (\d+) handlers registered$`, iShouldHaveAtLeastHandlersRegistered)
	sc.Step(`^the handlers are loaded$`, theHandlersAreLoaded)
	sc.Step(`^the handlers are loaded with full config$`, theHandlersAreLoaded)

	// Handler validation steps
	sc.Step(`^each handler should have a valid type$`, eachHandlerShouldHaveAValidType)
	sc.Step(`^each handler should have a valid name$`, eachHandlerShouldHaveAValidName)
	sc.Step(`^the "([^"]*)" handler should have type "([^"]*)"$`, theHandlerShouldHaveType)

	// Handler retrieval steps
	sc.Step(`^I get the handler named "([^"]*)"$`, iGetTheHandlerNamed)
	sc.Step(`^I should receive the handler$`, iShouldReceiveTheHandler)
	sc.Step(`^I should receive nil$`, iShouldReceiveNil)
	sc.Step(`^the handler name should be "([^"]*)"$`, theHandlerNameShouldBe)
	sc.Step(`^the handler description should not be empty$`, theHandlerDescriptionShouldNotBeEmpty)
	sc.Step(`^the handler type should be "([^"]*)"$`, theHandlerTypeShouldBe)
	sc.Step(`^the handler should have a build image configured$`, theHandlerShouldHaveBuildImageConfigured)
	sc.Step(`^the handler should have a build command configured$`, theHandlerShouldHaveBuildCommandConfigured)
	sc.Step(`^the handler should have build steps configured$`, theHandlerShouldHaveBuildStepsConfigured)

	// Build handler dispatch steps
	sc.Step(`^I get the build handler for module type "([^"]*)" with capabilities "([^"]*)" and build dep "([^"]*)"$`,
		iGetTheBuildHandler)
	sc.Step(`^the resolved handler should be "([^"]*)"$`, theResolvedHandlerShouldBe)

	// Test handler dispatch steps
	sc.Step(`^I get the test handler for module type "([^"]*)" with capabilities "([^"]*)" and build dep "([^"]*)"$`,
		iGetTheTestHandler)

	// Match condition steps
	sc.Step(`^a dispatch rule with default match$`, aDispatchRuleWithDefaultMatch)
	sc.Step(`^a dispatch rule with type match "([^"]*)"$`, aDispatchRuleWithTypeMatch)
	sc.Step(`^a dispatch rule with build dep match "([^"]*)"$`, aDispatchRuleWithBuildDepMatch)
	sc.Step(`^a dispatch rule with capabilities match "([^"]*)"$`, aDispatchRuleWithCapabilitiesMatch)
	sc.Step(`^a dispatch rule with type "([^"]*)" and capabilities "([^"]*)"$`, aDispatchRuleWithTypeAndCapabilities)
	sc.Step(`^I evaluate the rule for any module type$`, iEvaluateTheRuleForAnyModuleType)
	sc.Step(`^I evaluate the rule for module type "([^"]*)"$`, iEvaluateTheRuleForModuleType)
	sc.Step(`^I evaluate the rule with build dep "([^"]*)"$`, iEvaluateTheRuleWithBuildDep)
	sc.Step(`^I evaluate the rule with capabilities "([^"]*)"$`, iEvaluateTheRuleWithCapabilities)
	sc.Step(`^I evaluate the rule for module type "([^"]*)" with capabilities "([^"]*)"$`,
		iEvaluateTheRuleForModuleTypeWithCapabilities)
	sc.Step(`^the rule should match$`, theRuleShouldMatch)
	sc.Step(`^the rule should not match$`, theRuleShouldNotMatch)

	// Cross-compile target steps
	sc.Step(`^I get the cross-compile targets$`, iGetTheCrossCompileTargets)
	sc.Step(`^I should have at least (\d+) targets$`, iShouldHaveAtLeastTargets)
	sc.Step(`^I should have target "([^"]*)"$`, iShouldHaveTarget)
	sc.Step(`^the "([^"]*)" target should have suffix "([^"]*)"$`, theTargetShouldHaveSuffix)

	// UPX platform steps
	sc.Step(`^I get the UPX platforms$`, iGetTheUPXPlatforms)
	sc.Step(`^UPX should be supported for "([^"]*)"$`, upxShouldBeSupportedFor)
	sc.Step(`^UPX should not be supported for "([^"]*)"$`, upxShouldNotBeSupportedFor)

	// Dockerfile path steps
	sc.Step(`^I get the dockerfile paths$`, iGetTheDockerfilePaths)
	sc.Step(`^I should have at least (\d+) paths$`, iShouldHaveAtLeastPaths)
	sc.Step(`^I should have path pattern "([^"]*)"$`, iShouldHavePathPattern)
	sc.Step(`^I resolve dockerfile path "([^"]*)" for moniker "([^"]*)" with root "([^"]*)"$`,
		iResolveDockerfilePath)
	sc.Step(`^the resolved path should be "([^"]*)"$`, theResolvedPathShouldBe)

	// CI platform steps
	sc.Step(`^I get the CI platforms$`, iGetTheCIPlatforms)
	sc.Step(`^I get the CI platforms string$`, iGetTheCIPlatformsString)
	sc.Step(`^I should have platform "([^"]*)"$`, iShouldHavePlatform)
	sc.Step(`^the result should contain "([^"]*)"$`, theResultShouldContain)

	// MkDocs handler steps
	sc.Step(`^I get the mkdocs handler$`, iGetTheMkdocsHandler)
	sc.Step(`^I get the mkdocs handler build config$`, iGetTheMkdocsHandlerBuildConfig)
	sc.Step(`^the config should have volumes defined$`, theConfigShouldHaveVolumesDefined)

	// NPM handler steps
	sc.Step(`^I get the npm handler$`, iGetTheNpmHandler)
	sc.Step(`^I get the npm handler build steps$`, iGetTheNpmHandlerBuildSteps)
	sc.Step(`^I get the npm handler test steps$`, iGetTheNpmHandlerTestSteps)
	sc.Step(`^the "([^"]*)" step should have a when condition$`, theStepShouldHaveWhenCondition)

	// Nil config steps
	sc.Step(`^a nil handlers configuration$`, aNilHandlersConfiguration)
	sc.Step(`^an empty handlers configuration$`, anEmptyHandlersConfiguration)

	// Validation steps
	sc.Step(`^a handler with name "([^"]*)" and type "([^"]*)"$`, aHandlerWithNameAndType)
	sc.Step(`^I validate the handler$`, iValidateTheHandler)
	sc.Step(`^the validation should succeed$`, theValidationShouldSucceed)
	sc.Step(`^the validation should fail$`, theValidationShouldFail)
	sc.Step(`^the error should mention "([^"]*)"$`, theErrorShouldMention)

	// Custom handler configuration steps
	sc.Step(`^a custom handlers YAML with a builtin handler "([^"]*)"$`, aCustomHandlersYAMLWithBuiltinHandler)
	sc.Step(`^a custom handlers YAML with a command handler "([^"]*)"$`, aCustomHandlersYAMLWithCommandHandler)
	sc.Step(`^a custom handlers YAML with a docker handler "([^"]*)"$`, aCustomHandlersYAMLWithDockerHandler)
	sc.Step(`^a custom handlers YAML with dispatch rules:$`, aCustomHandlersYAMLWithDispatchRules)
	sc.Step(`^I parse the handlers configuration$`, iParseTheHandlersConfiguration)
	sc.Step(`^the "([^"]*)" handler should exist$`, theCustomHandlerShouldExist)
	sc.Step(`^the "([^"]*)" handler should have build steps$`, theCustomHandlerShouldHaveBuildSteps)

	// Handler flag steps
	sc.Step(`^I get the build flags for handler "([^"]*)"$`, iGetTheBuildFlagsForHandler)
	sc.Step(`^I should have at least (\d+) build flags$`, iShouldHaveAtLeastBuildFlags)
	sc.Step(`^I should have no build flags$`, iShouldHaveNoBuildFlags)
	sc.Step(`^I get the build flag "([^"]*)" for handler "([^"]*)"$`, iGetTheBuildFlagForHandler)
	sc.Step(`^I get the build flag by CLI "([^"]*)" for handler "([^"]*)"$`, iGetTheBuildFlagByCLIForHandler)
	sc.Step(`^the flag should exist$`, theFlagShouldExist)
	sc.Step(`^the flag should not exist$`, theFlagShouldNotExist)
	sc.Step(`^the flag name should be "([^"]*)"$`, theFlagNameShouldBe)
	sc.Step(`^the flag type should be "([^"]*)"$`, theFlagTypeShouldBe)
	sc.Step(`^the flag CLI positive should be "([^"]*)"$`, theFlagCLIPositiveShouldBe)
	sc.Step(`^the flag CLI negative should be "([^"]*)"$`, theFlagCLINegativeShouldBe)
	sc.Step(`^the flag bool default for local should be (true|false)$`, theFlagBoolDefaultForLocalShouldBe)
	sc.Step(`^the flag bool default for CI should be (true|false)$`, theFlagBoolDefaultForCIShouldBe)
	sc.Step(`^the flag string default should be "([^"]*)"$`, theFlagStringDefaultShouldBe)
	sc.Step(`^I get all build CLI flags for handler "([^"]*)"$`, iGetAllBuildCLIFlagsForHandler)
	sc.Step(`^the CLI flags map should contain "([^"]*)"$`, theCLIFlagsMapShouldContain)
	sc.Step(`^the CLI flags map should be empty$`, theCLIFlagsMapShouldBeEmpty)
	sc.Step(`^a handler flag with name "([^"]*)", type "([^"]*)", cli_positive "([^"]*)", cli_negative "([^"]*)"$`,
		aHandlerFlagWithNameTypeCLIPositiveCLINegative)
	sc.Step(`^a handler flag with name "([^"]*)", type "([^"]*)", cli_positive "([^"]*)"$`,
		aHandlerFlagWithNameTypeCLIPositive)
	sc.Step(`^a handler flag with name "([^"]*)", type "([^"]*)", value_flag "([^"]*)"$`,
		aHandlerFlagWithNameTypeValueFlag)
	sc.Step(`^a handler flag with name "([^"]*)", type "([^"]*)"$`, aHandlerFlagWithNameType)
	sc.Step(`^a handler flag with type "([^"]*)", cli_positive "([^"]*)"$`, aHandlerFlagWithTypeCLIPositive)
	sc.Step(`^I validate the flag$`, iValidateTheFlag)
	sc.Step(`^the flag validation should succeed$`, theFlagValidationShouldSucceed)
	sc.Step(`^the flag validation should fail$`, theFlagValidationShouldFail)
	sc.Step(`^the flag error should mention "([^"]*)"$`, theFlagErrorShouldMention)
	sc.Step(`^a nil handler flag$`, aNilHandlerFlag)
	sc.Step(`^I get the flag default for local$`, iGetTheFlagDefaultForLocal)
	sc.Step(`^the result should be nil$`, theResultShouldBeNil)
}

func cleanupTestContext() {
	hcCtx = nil
}

// =============================================================================
// Background Steps
// =============================================================================

func theEACConfigIsLoaded() error {
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return fmt.Errorf("failed to load EAC config: %w", err)
	}
	hcCtx.cfg = cfg
	return nil
}

// =============================================================================
// Loading Steps
// =============================================================================

func iLoadTheHandlersConfiguration() error {
	if hcCtx.cfg == nil {
		cfg, err := config.Load(config.DefaultLoadOptions())
		if err != nil {
			hcCtx.loadError = err
			return nil
		}
		hcCtx.cfg = cfg
	}
	hcCtx.handlers = hcCtx.cfg.Handlers
	return nil
}

func theHandlersShouldBeLoadedSuccessfully() error {
	if hcCtx.loadError != nil {
		return fmt.Errorf("handlers failed to load: %w", hcCtx.loadError)
	}
	if hcCtx.handlers == nil {
		return fmt.Errorf("handlers is nil")
	}
	return nil
}

func iShouldHaveAtLeastHandlersRegistered(count int) error {
	if hcCtx.handlers == nil {
		return fmt.Errorf("handlers not loaded")
	}
	if len(hcCtx.handlers.Handlers) < count {
		return fmt.Errorf("expected at least %d handlers, got %d", count, len(hcCtx.handlers.Handlers))
	}
	return nil
}

func theHandlersAreLoaded() error {
	if err := iLoadTheHandlersConfiguration(); err != nil {
		return err
	}
	return theHandlersShouldBeLoadedSuccessfully()
}

// =============================================================================
// Handler Validation Steps
// =============================================================================

func eachHandlerShouldHaveAValidType() error {
	validTypes := map[string]bool{"builtin": true, "command": true, "script": true, "docker": true}
	for _, h := range hcCtx.handlers.Handlers {
		if !validTypes[h.Type] {
			return fmt.Errorf("handler %s has invalid type: %s", h.Name, h.Type)
		}
	}
	return nil
}

func eachHandlerShouldHaveAValidName() error {
	for _, h := range hcCtx.handlers.Handlers {
		if h.Name == "" {
			return fmt.Errorf("found handler with empty name")
		}
	}
	return nil
}

func theHandlerShouldHaveType(handlerName, expectedType string) error {
	// Check both handlers sources
	var h *config.Handler
	if hcCtx.customHandlers != nil {
		h = hcCtx.customHandlers.Get(handlerName)
	} else if hcCtx.handlers != nil {
		h = hcCtx.handlers.Get(handlerName)
	}
	if h == nil {
		return fmt.Errorf("handler %s not found", handlerName)
	}
	if h.Type != expectedType {
		return fmt.Errorf("handler %s expected type %s, got %s", handlerName, expectedType, h.Type)
	}
	return nil
}

// =============================================================================
// Handler Retrieval Steps
// =============================================================================

func iGetTheHandlerNamed(name string) error {
	if hcCtx.handlers != nil {
		hcCtx.handler = hcCtx.handlers.Get(name)
	} else if hcCtx.customHandlers != nil {
		hcCtx.handler = hcCtx.customHandlers.Get(name)
	} else {
		hcCtx.handler = nil
	}
	hcCtx.handlerName = name
	return nil
}

func iShouldReceiveTheHandler() error {
	if hcCtx.handler == nil {
		return fmt.Errorf("expected to receive handler, got nil")
	}
	return nil
}

func iShouldReceiveNil() error {
	if hcCtx.handler != nil {
		return fmt.Errorf("expected nil, got handler: %s", hcCtx.handler.Name)
	}
	return nil
}

func theHandlerNameShouldBe(expectedName string) error {
	if hcCtx.handler == nil {
		return fmt.Errorf("handler is nil")
	}
	if hcCtx.handler.Name != expectedName {
		return fmt.Errorf("expected name %s, got %s", expectedName, hcCtx.handler.Name)
	}
	return nil
}

func theHandlerDescriptionShouldNotBeEmpty() error {
	if hcCtx.handler == nil {
		return fmt.Errorf("handler is nil")
	}
	if hcCtx.handler.Description == "" {
		return fmt.Errorf("handler %s has empty description", hcCtx.handler.Name)
	}
	return nil
}

func theHandlerTypeShouldBe(expectedType string) error {
	if hcCtx.handler == nil {
		return fmt.Errorf("handler is nil")
	}
	if hcCtx.handler.Type != expectedType {
		return fmt.Errorf("expected type %s, got %s", expectedType, hcCtx.handler.Type)
	}
	return nil
}

func theHandlerShouldHaveBuildImageConfigured() error {
	if hcCtx.handler == nil {
		return fmt.Errorf("handler is nil")
	}
	if hcCtx.handler.Build == nil || hcCtx.handler.Build.Image == "" {
		return fmt.Errorf("handler %s does not have build image configured", hcCtx.handler.Name)
	}
	return nil
}

func theHandlerShouldHaveBuildCommandConfigured() error {
	if hcCtx.handler == nil {
		return fmt.Errorf("handler is nil")
	}
	if hcCtx.handler.Build == nil || len(hcCtx.handler.Build.Command) == 0 {
		return fmt.Errorf("handler %s does not have build command configured", hcCtx.handler.Name)
	}
	return nil
}

func theHandlerShouldHaveBuildStepsConfigured() error {
	if hcCtx.handler == nil {
		return fmt.Errorf("handler is nil")
	}
	if hcCtx.handler.Build == nil || len(hcCtx.handler.Build.Steps) == 0 {
		return fmt.Errorf("handler %s does not have build steps configured", hcCtx.handler.Name)
	}
	return nil
}

// =============================================================================
// Dispatch Steps
// =============================================================================

func iGetTheBuildHandler(moduleType, capsStr, buildDep string) error {
	caps := parseCapabilities(capsStr)
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.resolvedHandler = handlers.GetBuildHandler(moduleType, caps, buildDep)
	return nil
}

func iGetTheTestHandler(moduleType, capsStr, buildDep string) error {
	caps := parseCapabilities(capsStr)
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.resolvedHandler = handlers.GetTestHandler(moduleType, caps, buildDep)
	return nil
}

func theResolvedHandlerShouldBe(expected string) error {
	if hcCtx.resolvedHandler != expected {
		return fmt.Errorf("expected resolved handler %q, got %q", expected, hcCtx.resolvedHandler)
	}
	return nil
}

// =============================================================================
// Match Condition Steps
// =============================================================================

func aDispatchRuleWithDefaultMatch() error {
	hcCtx.dispatchRule = config.DispatchRule{
		Match:   config.MatchCondition{Default: true},
		Handler: "test",
	}
	return nil
}

func aDispatchRuleWithTypeMatch(typeMatch string) error {
	hcCtx.dispatchRule = config.DispatchRule{
		Match:   config.MatchCondition{Type: typeMatch},
		Handler: "test",
	}
	return nil
}

func aDispatchRuleWithBuildDepMatch(buildDep string) error {
	hcCtx.dispatchRule = config.DispatchRule{
		Match:   config.MatchCondition{BuildDep: buildDep},
		Handler: "test",
	}
	return nil
}

func aDispatchRuleWithCapabilitiesMatch(capsStr string) error {
	hcCtx.dispatchRule = config.DispatchRule{
		Match:   config.MatchCondition{Capabilities: parseCapabilities(capsStr)},
		Handler: "test",
	}
	return nil
}

func aDispatchRuleWithTypeAndCapabilities(typeMatch, capsStr string) error {
	hcCtx.dispatchRule = config.DispatchRule{
		Match: config.MatchCondition{
			Type:         typeMatch,
			Capabilities: parseCapabilities(capsStr),
		},
		Handler: "test",
	}
	return nil
}

func iEvaluateTheRuleForAnyModuleType() error {
	cfg := &config.HandlersConfig{}
	hcCtx.ruleMatches = cfg.MatchesRule(hcCtx.dispatchRule, "any-type", []string{"any"}, "any")
	return nil
}

func iEvaluateTheRuleForModuleType(moduleType string) error {
	cfg := &config.HandlersConfig{}
	hcCtx.ruleMatches = cfg.MatchesRule(hcCtx.dispatchRule, moduleType, nil, "")
	return nil
}

func iEvaluateTheRuleWithBuildDep(buildDep string) error {
	cfg := &config.HandlersConfig{}
	hcCtx.ruleMatches = cfg.MatchesRule(hcCtx.dispatchRule, "any", nil, buildDep)
	return nil
}

func iEvaluateTheRuleWithCapabilities(capsStr string) error {
	cfg := &config.HandlersConfig{}
	caps := parseCapabilities(capsStr)
	hcCtx.ruleMatches = cfg.MatchesRule(hcCtx.dispatchRule, "any", caps, "")
	return nil
}

func iEvaluateTheRuleForModuleTypeWithCapabilities(moduleType, capsStr string) error {
	cfg := &config.HandlersConfig{}
	caps := parseCapabilities(capsStr)
	hcCtx.ruleMatches = cfg.MatchesRule(hcCtx.dispatchRule, moduleType, caps, "")
	return nil
}

func theRuleShouldMatch() error {
	if !hcCtx.ruleMatches {
		return fmt.Errorf("expected rule to match, but it did not")
	}
	return nil
}

func theRuleShouldNotMatch() error {
	if hcCtx.ruleMatches {
		return fmt.Errorf("expected rule not to match, but it did")
	}
	return nil
}

// =============================================================================
// Cross-Compile Target Steps
// =============================================================================

func iGetTheCrossCompileTargets() error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.crossCompileTargets = handlers.GetCrossCompileTargets()
	return nil
}

func iShouldHaveAtLeastTargets(count int) error {
	if len(hcCtx.crossCompileTargets) < count {
		return fmt.Errorf("expected at least %d targets, got %d", count, len(hcCtx.crossCompileTargets))
	}
	return nil
}

func iShouldHaveTarget(target string) error {
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s (expected os/arch)", target)
	}
	os, arch := parts[0], parts[1]

	for _, t := range hcCtx.crossCompileTargets {
		if t.OS == os && t.Arch == arch {
			return nil
		}
	}
	return fmt.Errorf("target %s not found", target)
}

func theTargetShouldHaveSuffix(target, expectedSuffix string) error {
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", target)
	}
	os, arch := parts[0], parts[1]

	for _, t := range hcCtx.crossCompileTargets {
		if t.OS == os && t.Arch == arch {
			if t.Suffix != expectedSuffix {
				return fmt.Errorf("target %s expected suffix %q, got %q", target, expectedSuffix, t.Suffix)
			}
			return nil
		}
	}
	return fmt.Errorf("target %s not found", target)
}

// =============================================================================
// UPX Platform Steps
// =============================================================================

func iGetTheUPXPlatforms() error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.upxPlatforms = handlers.GetUPXPlatforms()
	return nil
}

func upxShouldBeSupportedFor(platform string) error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	if !handlers.IsUPXSupported(platform) {
		return fmt.Errorf("UPX should be supported for %s but it is not", platform)
	}
	return nil
}

func upxShouldNotBeSupportedFor(platform string) error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	if handlers.IsUPXSupported(platform) {
		return fmt.Errorf("UPX should not be supported for %s but it is", platform)
	}
	return nil
}

// =============================================================================
// Dockerfile Path Steps
// =============================================================================

func iGetTheDockerfilePaths() error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.dockerfilePaths = handlers.GetDockerfilePaths()
	return nil
}

func iShouldHaveAtLeastPaths(count int) error {
	if len(hcCtx.dockerfilePaths) < count {
		return fmt.Errorf("expected at least %d paths, got %d", count, len(hcCtx.dockerfilePaths))
	}
	return nil
}

func iShouldHavePathPattern(pattern string) error {
	for _, p := range hcCtx.dockerfilePaths {
		if p == pattern {
			return nil
		}
	}
	return fmt.Errorf("path pattern %q not found in %v", pattern, hcCtx.dockerfilePaths)
}

func iResolveDockerfilePath(template, moniker, root string) error {
	hcCtx.resolvedPath = config.ResolveDockerfilePath(template, moniker, root)
	return nil
}

func theResolvedPathShouldBe(expected string) error {
	if hcCtx.resolvedPath != expected {
		return fmt.Errorf("expected resolved path %q, got %q", expected, hcCtx.resolvedPath)
	}
	return nil
}

// =============================================================================
// CI Platform Steps
// =============================================================================

func iGetTheCIPlatforms() error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.ciPlatforms = handlers.GetCIPlatforms()
	return nil
}

func iGetTheCIPlatformsString() error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.ciPlatformsString = handlers.GetCIPlatformsString()
	return nil
}

func iShouldHavePlatform(platform string) error {
	for _, p := range hcCtx.ciPlatforms {
		if p == platform {
			return nil
		}
	}
	return fmt.Errorf("platform %q not found in %v", platform, hcCtx.ciPlatforms)
}

func theResultShouldContain(text string) error {
	if !strings.Contains(hcCtx.ciPlatformsString, text) {
		return fmt.Errorf("expected result to contain %q, got %q", text, hcCtx.ciPlatformsString)
	}
	return nil
}

// =============================================================================
// MkDocs Handler Steps
// =============================================================================

func iGetTheMkdocsHandler() error {
	hcCtx.handler = hcCtx.handlers.GetMkDocsHandler()
	return nil
}

func iGetTheMkdocsHandlerBuildConfig() error {
	handler := hcCtx.handlers.GetMkDocsHandler()
	if handler == nil {
		return fmt.Errorf("mkdocs handler not found")
	}
	hcCtx.handler = handler
	return nil
}

func theConfigShouldHaveVolumesDefined() error {
	if hcCtx.handler == nil || hcCtx.handler.Build == nil {
		return fmt.Errorf("handler or build config is nil")
	}
	if len(hcCtx.handler.Build.Volumes) == 0 {
		return fmt.Errorf("no volumes defined")
	}
	return nil
}

// =============================================================================
// NPM Handler Steps
// =============================================================================

func iGetTheNpmHandler() error {
	hcCtx.handler = hcCtx.handlers.Get("npm")
	return nil
}

func iGetTheNpmHandlerBuildSteps() error {
	handler := hcCtx.handlers.Get("npm")
	if handler == nil {
		return fmt.Errorf("npm handler not found")
	}
	hcCtx.handler = handler
	return nil
}

func iGetTheNpmHandlerTestSteps() error {
	handler := hcCtx.handlers.Get("npm")
	if handler == nil {
		return fmt.Errorf("npm handler not found")
	}
	hcCtx.handler = handler
	return nil
}

func theStepShouldHaveWhenCondition(stepName string) error {
	if hcCtx.handler == nil {
		return fmt.Errorf("handler is nil")
	}

	// Check build steps
	if hcCtx.handler.Build != nil {
		for _, step := range hcCtx.handler.Build.Steps {
			if step.Name == stepName {
				if step.When == "" {
					return fmt.Errorf("step %s does not have a when condition", stepName)
				}
				return nil
			}
		}
	}

	// Check test steps
	if hcCtx.handler.Test != nil {
		for _, step := range hcCtx.handler.Test.Steps {
			if step.Name == stepName {
				if step.When == "" {
					return fmt.Errorf("step %s does not have a when condition", stepName)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("step %s not found", stepName)
}

// =============================================================================
// Nil Config Steps
// =============================================================================

func aNilHandlersConfiguration() error {
	hcCtx.handlers = nil
	hcCtx.customHandlers = nil
	return nil
}

func anEmptyHandlersConfiguration() error {
	hcCtx.handlers = &config.HandlersConfig{
		Handlers: []config.Handler{},
	}
	hcCtx.handlers.BuildHandlerMap()
	return nil
}

// =============================================================================
// Validation Steps
// =============================================================================

func aHandlerWithNameAndType(name, handlerType string) error {
	hcCtx.handler = &config.Handler{
		Name: name,
		Type: handlerType,
	}
	return nil
}

func iValidateTheHandler() error {
	if hcCtx.handler == nil {
		return fmt.Errorf("handler is nil")
	}
	hcCtx.validationError = hcCtx.handler.Validate()
	return nil
}

func theValidationShouldSucceed() error {
	if hcCtx.validationError != nil {
		return fmt.Errorf("expected validation to succeed, got error: %v", hcCtx.validationError)
	}
	return nil
}

func theValidationShouldFail() error {
	if hcCtx.validationError == nil {
		return fmt.Errorf("expected validation to fail, but it succeeded")
	}
	return nil
}

func theErrorShouldMention(text string) error {
	if hcCtx.validationError == nil {
		return fmt.Errorf("no error to check")
	}
	if !strings.Contains(hcCtx.validationError.Error(), text) {
		return fmt.Errorf("expected error to mention %q, got: %v", text, hcCtx.validationError)
	}
	return nil
}

// =============================================================================
// Custom Handler Configuration Steps
// =============================================================================

func aCustomHandlersYAMLWithBuiltinHandler(name string) error {
	hcCtx.customYAML = fmt.Sprintf(`
handlers:
  - name: %s
    description: "Custom builtin handler"
    type: builtin
    build:
      handler: %s
      config:
        custom_key: custom_value
    test:
      handler: %s
`, name, name, name)
	return nil
}

func aCustomHandlersYAMLWithCommandHandler(name string) error {
	hcCtx.customYAML = fmt.Sprintf(`
handlers:
  - name: %s
    description: "Custom command handler"
    type: command
    build:
      steps:
        - name: install
          command: npm install
          workdir: "{root}"
        - name: build
          command: npm run build
          workdir: "{root}"
    test:
      steps:
        - name: test
          command: npm test
          workdir: "{root}"
`, name)
	return nil
}

func aCustomHandlersYAMLWithDockerHandler(name string) error {
	hcCtx.customYAML = fmt.Sprintf(`
handlers:
  - name: %s
    description: "Custom docker handler"
    type: docker
    build:
      image: custom-image:latest
      command: ["build", "--output", "/out"]
      workdir: /app
      volumes:
        - "{workspace}:/app"
`, name)
	return nil
}

func aCustomHandlersYAMLWithDispatchRules(docString *godog.DocString) error {
	// Start with a minimal handlers config and append the dispatch rules
	hcCtx.customYAML = fmt.Sprintf(`
handlers:
  - name: custom-handler
    description: "Custom handler for testing"
    type: builtin
    build:
      handler: custom
  - name: special-handler
    description: "Special handler for capability testing"
    type: builtin
    build:
      handler: special
  - name: go
    description: "Go handler"
    type: builtin
    build:
      handler: go

%s
`, docString.Content)
	return nil
}

func iParseTheHandlersConfiguration() error {
	var cfg config.HandlersConfig
	if err := yaml.Unmarshal([]byte(hcCtx.customYAML), &cfg); err != nil {
		return fmt.Errorf("failed to parse handlers YAML: %w", err)
	}
	cfg.BuildHandlerMap()
	hcCtx.customHandlers = &cfg
	hcCtx.handlers = nil // Clear default handlers to use custom
	return nil
}

func theCustomHandlerShouldExist(name string) error {
	if hcCtx.customHandlers == nil {
		return fmt.Errorf("custom handlers not loaded")
	}
	handler := hcCtx.customHandlers.Get(name)
	if handler == nil {
		// List available handlers for debugging
		var available []string
		for _, h := range hcCtx.customHandlers.Handlers {
			available = append(available, h.Name)
		}
		return fmt.Errorf("handler %s not found, available handlers: %v", name, available)
	}
	hcCtx.handler = handler
	return nil
}

func theCustomHandlerShouldHaveBuildSteps(name string) error {
	if hcCtx.customHandlers == nil {
		return fmt.Errorf("custom handlers not loaded")
	}
	handler := hcCtx.customHandlers.Get(name)
	if handler == nil {
		return fmt.Errorf("handler %s not found", name)
	}
	if handler.Build == nil || len(handler.Build.Steps) == 0 {
		return fmt.Errorf("handler %s does not have build steps", name)
	}
	return nil
}

// =============================================================================
// Handler Flag Steps
// =============================================================================

func iGetTheBuildFlagsForHandler(handlerName string) error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.buildFlags = handlers.GetBuildFlags(handlerName)
	return nil
}

func iShouldHaveAtLeastBuildFlags(count int) error {
	if len(hcCtx.buildFlags) < count {
		return fmt.Errorf("expected at least %d build flags, got %d", count, len(hcCtx.buildFlags))
	}
	return nil
}

func iShouldHaveNoBuildFlags() error {
	if hcCtx.buildFlags != nil && len(hcCtx.buildFlags) > 0 {
		return fmt.Errorf("expected no build flags, got %d", len(hcCtx.buildFlags))
	}
	return nil
}

func iGetTheBuildFlagForHandler(flagName, handlerName string) error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.currentFlag = handlers.GetBuildFlagByName(handlerName, flagName)
	return nil
}

func iGetTheBuildFlagByCLIForHandler(cliFlag, handlerName string) error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.currentFlag = handlers.GetBuildFlagByCLI(handlerName, cliFlag)
	return nil
}

func theFlagShouldExist() error {
	if hcCtx.currentFlag == nil {
		return fmt.Errorf("flag should exist but it is nil")
	}
	return nil
}

func theFlagShouldNotExist() error {
	if hcCtx.currentFlag != nil {
		return fmt.Errorf("flag should not exist but found: %s", hcCtx.currentFlag.Name)
	}
	return nil
}

func theFlagNameShouldBe(expected string) error {
	if hcCtx.currentFlag == nil {
		return fmt.Errorf("flag is nil")
	}
	if hcCtx.currentFlag.Name != expected {
		return fmt.Errorf("expected flag name %q, got %q", expected, hcCtx.currentFlag.Name)
	}
	return nil
}

func theFlagTypeShouldBe(expected string) error {
	if hcCtx.currentFlag == nil {
		return fmt.Errorf("flag is nil")
	}
	if hcCtx.currentFlag.Type != expected {
		return fmt.Errorf("expected flag type %q, got %q", expected, hcCtx.currentFlag.Type)
	}
	return nil
}

func theFlagCLIPositiveShouldBe(expected string) error {
	if hcCtx.currentFlag == nil {
		return fmt.Errorf("flag is nil")
	}
	if hcCtx.currentFlag.CLIPositive != expected {
		return fmt.Errorf("expected CLI positive %q, got %q", expected, hcCtx.currentFlag.CLIPositive)
	}
	return nil
}

func theFlagCLINegativeShouldBe(expected string) error {
	if hcCtx.currentFlag == nil {
		return fmt.Errorf("flag is nil")
	}
	if hcCtx.currentFlag.CLINegative != expected {
		return fmt.Errorf("expected CLI negative %q, got %q", expected, hcCtx.currentFlag.CLINegative)
	}
	return nil
}

func theFlagBoolDefaultForLocalShouldBe(expected string) error {
	if hcCtx.currentFlag == nil {
		return fmt.Errorf("flag is nil")
	}
	actual := hcCtx.currentFlag.GetBoolDefault(false) // false = local (not CI)
	expectedBool := expected == "true"
	if actual != expectedBool {
		return fmt.Errorf("expected bool default for local to be %v, got %v", expectedBool, actual)
	}
	return nil
}

func theFlagBoolDefaultForCIShouldBe(expected string) error {
	if hcCtx.currentFlag == nil {
		return fmt.Errorf("flag is nil")
	}
	actual := hcCtx.currentFlag.GetBoolDefault(true) // true = CI
	expectedBool := expected == "true"
	if actual != expectedBool {
		return fmt.Errorf("expected bool default for CI to be %v, got %v", expectedBool, actual)
	}
	return nil
}

func theFlagStringDefaultShouldBe(expected string) error {
	if hcCtx.currentFlag == nil {
		return fmt.Errorf("flag is nil")
	}
	actual := hcCtx.currentFlag.GetStringDefault(false)
	if actual != expected {
		return fmt.Errorf("expected string default %q, got %q", expected, actual)
	}
	return nil
}

func iGetAllBuildCLIFlagsForHandler(handlerName string) error {
	handlers := hcCtx.handlers
	if handlers == nil {
		handlers = hcCtx.customHandlers
	}
	hcCtx.cliFlagsMap = handlers.GetAllBuildCLIFlags(handlerName)
	return nil
}

func theCLIFlagsMapShouldContain(cliFlag string) error {
	if hcCtx.cliFlagsMap == nil {
		return fmt.Errorf("CLI flags map is nil")
	}
	if _, ok := hcCtx.cliFlagsMap[cliFlag]; !ok {
		return fmt.Errorf("CLI flag %q not found in map", cliFlag)
	}
	return nil
}

func theCLIFlagsMapShouldBeEmpty() error {
	if hcCtx.cliFlagsMap != nil && len(hcCtx.cliFlagsMap) > 0 {
		return fmt.Errorf("expected CLI flags map to be empty, got %d entries", len(hcCtx.cliFlagsMap))
	}
	return nil
}

func aHandlerFlagWithNameTypeCLIPositiveCLINegative(name, flagType, cliPositive, cliNegative string) error {
	hcCtx.currentFlag = &config.HandlerFlag{
		Name:        name,
		Type:        flagType,
		CLIPositive: cliPositive,
		CLINegative: cliNegative,
	}
	return nil
}

func aHandlerFlagWithNameTypeCLIPositive(name, flagType, cliPositive string) error {
	hcCtx.currentFlag = &config.HandlerFlag{
		Name:        name,
		Type:        flagType,
		CLIPositive: cliPositive,
	}
	return nil
}

func aHandlerFlagWithNameTypeValueFlag(name, flagType, valueFlag string) error {
	hcCtx.currentFlag = &config.HandlerFlag{
		Name:      name,
		Type:      flagType,
		ValueFlag: valueFlag,
	}
	return nil
}

func aHandlerFlagWithNameType(name, flagType string) error {
	hcCtx.currentFlag = &config.HandlerFlag{
		Name: name,
		Type: flagType,
	}
	return nil
}

func aHandlerFlagWithTypeCLIPositive(flagType, cliPositive string) error {
	hcCtx.currentFlag = &config.HandlerFlag{
		Name:        "", // Missing name
		Type:        flagType,
		CLIPositive: cliPositive,
	}
	return nil
}

func iValidateTheFlag() error {
	if hcCtx.currentFlag == nil {
		return fmt.Errorf("flag is nil")
	}
	hcCtx.flagValidateErr = hcCtx.currentFlag.Validate()
	return nil
}

func theFlagValidationShouldSucceed() error {
	if hcCtx.flagValidateErr != nil {
		return fmt.Errorf("expected flag validation to succeed, got error: %v", hcCtx.flagValidateErr)
	}
	return nil
}

func theFlagValidationShouldFail() error {
	if hcCtx.flagValidateErr == nil {
		return fmt.Errorf("expected flag validation to fail, but it succeeded")
	}
	return nil
}

func theFlagErrorShouldMention(text string) error {
	if hcCtx.flagValidateErr == nil {
		return fmt.Errorf("no flag error to check")
	}
	if !strings.Contains(hcCtx.flagValidateErr.Error(), text) {
		return fmt.Errorf("expected flag error to mention %q, got: %v", text, hcCtx.flagValidateErr)
	}
	return nil
}

func aNilHandlerFlag() error {
	hcCtx.currentFlag = nil
	return nil
}

func iGetTheFlagDefaultForLocal() error {
	hcCtx.flagDefault = hcCtx.currentFlag.GetDefault(false)
	return nil
}

func theResultShouldBeNil() error {
	if hcCtx.flagDefault != nil {
		return fmt.Errorf("expected result to be nil, got %v", hcCtx.flagDefault)
	}
	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func parseCapabilities(capsStr string) []string {
	if capsStr == "" {
		return nil
	}
	return strings.Split(capsStr, ",")
}
