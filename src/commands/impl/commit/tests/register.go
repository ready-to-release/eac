// Package tests provides BDD step definitions for the commit command.
//
// This file contains the step registration function and exported wrappers
// for cross-package state synchronization.
package tests

import "github.com/cucumber/godog"

// InitializeCommitScenario registers all commit-specific step definitions.
func InitializeCommitScenario(sc *godog.ScenarioContext) {
	registerContractSteps(sc)
	registerSetupSteps(sc)
	registerProcessingSteps(sc)
	registerValidationSteps(sc)
	registerVerificationSteps(sc)
}

func registerContractSteps(sc *godog.ScenarioContext) {
	sc.Step(`^no version mismatch errors should occur$`, noVersionMismatchErrorsShouldOccur)
	sc.Step(`^the contract must include "([^"]*)" section$`, theContractMustIncludeModuleSectionsSection)
	sc.Step(`^the contract must include semantic types: (.*)$`, theContractMustIncludeSemanticTypes)
	sc.Step(`^the contract is loaded$`, theContractIsLoaded)
	sc.Step(`^the contract implementation is verified$`, theContractImplementationIsVerified)
	sc.Step(`^a commit message contract$`, aCommitMessageContract)
	sc.Step(`^a commit message contract with version "([^"]*)"$`, aCommitMessageContractWithVersion)
	sc.Step(`^the contract file does not exist$`, theContractFileDoesNotExistCommit)
	sc.Step(`^a contract file with invalid YAML$`, aContractFileDoesNotExistCommit)
}

func registerSetupSteps(sc *godog.ScenarioContext) {
	// Commit message setup
	sc.Step(`^a commit message with header "([^"]*)"$`, aCommitMessageWithHeader)
	sc.Step(`^a commit message with header longer than (\d+) characters$`, func(length int) error {
		return aCommitMessageWithHeaderLongerThan72Characters()
	})
	sc.Step(`^a commit message with header ending in a period$`, aCommitMessageWithHeaderEndingInAPeriod)
	sc.Step(`^a commit message without Auditor-Summary$`, aCommitMessageWithoutAuditorSummary)
	sc.Step(`^a commit message header ending with a period$`, aCommitMessageHeaderEndingWithAPeriod)
	sc.Step(`^a commit message with an opening code fence but no closing fence$`, aCommitMessageWithAnOpeningCodeFenceButNoClosingFence)
	sc.Step(`^a commit message with multiple consecutive blank lines$`, aCommitMessageWithMultipleConsecutiveBlankLines)
	sc.Step(`^a commit message with Unicode characters$`, aCommitMessageWithUnicodeCharacters)
	sc.Step(`^a multi-module commit message$`, aMultiModuleCommitMessage)
	sc.Step(`^a top-level body section$`, aTopLevelBodySection)
	sc.Step(`^an Auditor-Summary field$`, anAuditorSummaryField)
	sc.Step(`^a body text line longer than (\d+) characters$`, func(length int) error {
		return aBodyTextLineLongerThan72Characters()
	})
	sc.Step(`^a code block without blank lines before and after$`, aCodeBlockWithoutBlankLinesBeforeAndAfter)

	// AI output setup
	sc.Step(`^AI output starting with "([^"]*)"$`, aiOutputStartingWith)
	sc.Step(`^followed by a valid commit header "([^"]*)"$`, followedByAValidCommitHeader)
	sc.Step(`^followed by module section "([^"]*)"$`, followedByModuleSection)
	sc.Step(`^AI output wrapped in triple backticks$`, aiOutputWrappedInTripleBackticksCommit)

	// Module setup
	sc.Step(`^multiple affected modules$`, multipleAffectedModules)
	sc.Step(`^one affected module$`, oneAffectedModule)
	sc.Step(`^module sections for each affected module$`, moduleSectionsForEachAffectedModule)
	sc.Step(`^module names with edge cases \(([^)]*)\)$`, func(cases string) error {
		return moduleNamesWithEdgeCases()
	})
	sc.Step(`^some modules missing from the output$`, someModulesMissingFromTheOutput)

	// Git state setup
	sc.Step(`^a full git diff$`, aFullGitDiff)
	sc.Step(`^a full git diff with multiple files$`, aFullGitDiffWithMultipleFiles)
	sc.Step(`^a git diff larger than (\d+) MB$`, func(size int) error {
		return aGitDiffLargerThan10MB()
	})
	sc.Step(`^a git diff for those files$`, aGitDiffForThoseFiles)
	sc.Step(`^staged files belonging to one module$`, stagedFilesBelongingToOneModule)
	sc.Step(`^staged files belonging to multiple modules$`, stagedFilesBelongingToMultipleModules)
	sc.Step(`^a module with multiple files$`, aModuleWithMultipleFiles)
	sc.Step(`^a module with specific files$`, aModuleWithSpecificFiles)
	sc.Step(`^a module with files not in the diff$`, aModuleWithFilesNotInTheDiff)
	sc.Step(`^a module with one file$`, aModuleWithOneFile)
	sc.Step(`^git diff command fails$`, gitDiffCommandFails)
	sc.Step(`^no staged changes in git$`, noStagedChangesInGit)
}

func registerProcessingSteps(sc *godog.ScenarioContext) {
	sc.Step(`^noise filtering is applied$`, noiseFilteringIsApplied)
	sc.Step(`^auto-cleanup is applied$`, autoCleanupIsApplied)
	sc.Step(`^module sections are generated$`, moduleSectionsAreGenerated)
	sc.Step(`^execution context is built$`, executionContextIsBuiltCommit)
	sc.Step(`^module context is built$`, moduleContextIsBuilt)
	sc.Step(`^module names are validated$`, moduleNamesAreValidated)
	sc.Step(`^top-level context is built$`, topLevelContextIsBuilt)
	sc.Step(`^missing modules are added$`, missingModulesAreAdded)
	sc.Step(`^the commit command is run$`, theCommitAiCommandIsRun)
}

func registerValidationSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the message is validated$`, theMessageIsValidated)
	sc.Step(`^validation should correctly accept or reject based on rules$`, validationShouldCorrectlyAcceptOrRejectBasedOnRulesCommit)
	sc.Step(`^a "HEADER_TOO_LONG" error should occur$`, aHeaderTooLongErrorShouldOccur)
	sc.Step(`^a "HEADER_TRAILING_PERIOD" error should occur$`, aHeaderTrailingPeriodErrorShouldOccur)
	sc.Step(`^a "MISSING_AUDITOR_SUMMARY" error should occur$`, aMissingAuditorSummaryErrorShouldOccur)
}

func registerVerificationSteps(sc *godog.ScenarioContext) {
	// Output format
	sc.Step(`^the output should start with "([^"]*)"$`, theOutputShouldStartWith)

	// Text cleanup
	sc.Step(`^the code fences should be removed$`, theCodeFencesShouldBeRemoved)
	sc.Step(`^the period should be removed$`, thePeriodShouldBeRemoved)
	sc.Step(`^the line should be wrapped at word boundaries$`, theLineShouldBeWrappedAtWordBoundaries)
	sc.Step(`^a closing fence should be added$`, aClosingFenceShouldBeAdded)
	sc.Step(`^duplicate blank lines should be reduced to single blank lines$`, duplicateBlankLinesShouldBeReducedToSingleBlankLines)
	sc.Step(`^blank lines should be added before and after the code block$`, blankLinesShouldBeAddedBeforeAndAfterTheCodeBlock)

	// Context verification
	sc.Step(`^the context should include the git diff$`, theContextShouldIncludeTheGitDiff)
	sc.Step(`^the context should include the staged files table$`, theContextShouldIncludeTheStagedFilesTable)
	sc.Step(`^the context should include only files for that module$`, theContextShouldIncludeOnlyFilesForThatModule)
	sc.Step(`^the diff should be filtered to only that module's changes$`, theDiffShouldBeFilteredToOnlyThatModulesChanges)
	sc.Step(`^the context should list all affected modules$`, theContextShouldListAllAffectedModules)
	sc.Step(`^the context should list the affected module$`, theContextShouldListTheAffectedModule)
	sc.Step(`^the context should indicate "([^"]*)"$`, theContextShouldIndicate)
	sc.Step(`^the context should indicate the count as "([^"]*)"$`, theContextShouldIndicateTheCountAs)

	// Module sections
	sc.Step(`^no module sections should be created$`, noModuleSectionsShouldBeCreatedCommit)
	sc.Step(`^a section should be created for each module$`, aSectionShouldBeCreatedForEachModule)
	sc.Step(`^stubs should indicate "([^"]*)"$`, stubsShouldIndicateModuleChangesNotDescribedByAIAgent)
	sc.Step(`^stub sections should be generated for missing modules$`, stubSectionsShouldBeGeneratedForMissingModules)
	sc.Step(`^all of that module's files should be included$`, allOfThatModulesFilesShouldBeIncluded)

	// Diff filtering
	sc.Step(`^only that file's diff should be included$`, onlyThatFilesDiffShouldBeIncludedCommit)
	sc.Step(`^other files should be excluded$`, otherFilesShouldBeExcludedCommit)
	sc.Step(`^the diff is filtered for that module$`, theDiffIsFilteredForThatModule)

	// Edge cases
	sc.Step(`^the message "([^"]*)" should be displayed$`, theMessageNoStagedChangesShouldBeDisplayed)
	sc.Step(`^the error should indicate git failure$`, theErrorShouldIndicateGitFailure)
	sc.Step(`^the error should indicate diff size limit exceeded$`, theErrorShouldIndicateDiffSizeLimitExceeded)
}

// ============================================================================
// Exported Step Functions (for cross-package state synchronization)
// ============================================================================

// GitDiffCommandFails is the exported wrapper for gitDiffCommandFails.
func GitDiffCommandFails() error { return gitDiffCommandFails() }

// AGitDiffLargerThan10MB is the exported wrapper (ignores size param).
func AGitDiffLargerThan10MB(size int) error { return aGitDiffLargerThan10MB() }

// TheContractFileDoesNotExistCommit is the exported wrapper.
func TheContractFileDoesNotExistCommit() error { return theContractFileDoesNotExistCommit() }

// AContractFileWithInvalidYAMLCommit is the exported wrapper.
func AContractFileWithInvalidYAMLCommit() error { return aContractFileDoesNotExistCommit() }

// NoStagedChangesInGit is the exported wrapper.
func NoStagedChangesInGit() error { return noStagedChangesInGit() }

// ModuleSectionsAreGenerated is the exported wrapper.
func ModuleSectionsAreGenerated() error { return moduleSectionsAreGenerated() }

// MultipleAffectedModules is the exported wrapper.
func MultipleAffectedModules() error { return multipleAffectedModules() }

// OneAffectedModule is the exported wrapper.
func OneAffectedModule() error { return oneAffectedModule() }

// ExecutionContextIsBuiltCommit is the exported wrapper.
func ExecutionContextIsBuiltCommit() error { return executionContextIsBuiltCommit() }

// ACommitMessageWithHeader is the exported wrapper.
func ACommitMessageWithHeader(header string) error { return aCommitMessageWithHeader(header) }

// ACommitMessageWithHeaderLongerThan72CharactersInt is the exported wrapper (ignores length param).
func ACommitMessageWithHeaderLongerThan72CharactersInt(length int) error {
	return aCommitMessageWithHeaderLongerThan72Characters()
}

// ACommitMessageWithHeaderEndingInAPeriod is the exported wrapper.
func ACommitMessageWithHeaderEndingInAPeriod() error { return aCommitMessageWithHeaderEndingInAPeriod() }

// ACommitMessageWithoutAuditorSummary is the exported wrapper.
func ACommitMessageWithoutAuditorSummary() error { return aCommitMessageWithoutAuditorSummary() }

// NoiseFilteringIsApplied is the exported wrapper.
func NoiseFilteringIsApplied() error { return noiseFilteringIsApplied() }

// TheMessageIsValidated is the exported wrapper.
func TheMessageIsValidated() error { return theMessageIsValidated() }

// AiOutputStartingWith is the exported wrapper.
func AiOutputStartingWith(prefix string) error { return aiOutputStartingWith(prefix) }

// FollowedByAValidCommitHeader is the exported wrapper.
func FollowedByAValidCommitHeader(header string) error { return followedByAValidCommitHeader(header) }

// FollowedByModuleSection is the exported wrapper.
func FollowedByModuleSection(moduleName string) error { return followedByModuleSection(moduleName) }

// AiOutputWrappedInTripleBackticksCommit is the exported wrapper.
func AiOutputWrappedInTripleBackticksCommit() error { return aiOutputWrappedInTripleBackticksCommit() }
