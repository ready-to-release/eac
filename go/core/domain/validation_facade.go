// Package contracts provides backward-compatible access to validation types
// These are facades that re-export from go/eac/core/validation
package domain

import "github.com/ready-to-release/eac/go/core/validation"

// ValidationError is a facade for validation.ValidationError.
type ValidationError = validation.ValidationError

// ErrorCode is a facade for validation.ErrorCode.
type ErrorCode = validation.ErrorCode

// ErrorCategory is a facade for validation.ErrorCategory.
type ErrorCategory = validation.ErrorCategory

// ErrorSeverity is a facade for validation.ErrorSeverity.
type ErrorSeverity = validation.ErrorSeverity

// Validator is a facade for validation.Validator.
type Validator = validation.Validator

// NoOpValidator is a facade for validation.NoOpValidator.
type NoOpValidator = validation.NoOpValidator

// AIExecutor is a facade for validation.AIExecutor.
type AIExecutor = validation.AIExecutor

// AIExecutorWithProviderInfo is a facade for validation.AIExecutorWithProviderInfo.
type AIExecutorWithProviderInfo = validation.AIExecutorWithProviderInfo

// Error severity constants.
const (
	SeverityError   = validation.SeverityError
	SeverityWarning = validation.SeverityWarning
)

// Error category constants.
const (
	CategoryStructure  = validation.CategoryStructure
	CategorySemantic   = validation.CategorySemantic
	CategoryFormat     = validation.CategoryFormat
	CategoryConstraint = validation.CategoryConstraint
	CategoryCorruption = validation.CategoryCorruption
)

// Error code constructors.
func NewValidationError(code ErrorCode, message string, line int) *ValidationError {
	return validation.NewValidationError(code, message, line)
}

func NewValidationErrorWithContext(code ErrorCode, message string, line int, context string) *ValidationError {
	return validation.NewValidationErrorWithContext(code, message, line, context)
}

// Error code registry functions.
func GetErrorCode(code string) (*ErrorCode, bool) {
	return validation.GetErrorCode(code)
}

func MustGetErrorCode(code string) *ErrorCode {
	return validation.MustGetErrorCode(code)
}

func GetAllErrorCodes() map[string]*ErrorCode {
	return validation.GetAllErrorCodes()
}

func GetErrorCodesByCategory(category ErrorCategory) []*ErrorCode {
	return validation.GetErrorCodesByCategory(category)
}

func GetErrorCodesBySeverity(severity ErrorSeverity) []*ErrorCode {
	return validation.GetErrorCodesBySeverity(severity)
}

// Re-export all error code variables.
var (
	ErrEmptyOutput               = validation.ErrEmptyOutput
	ErrMissingTagsConfig         = validation.ErrMissingTagsConfig
	ErrNoContract                = validation.ErrNoContract
	ErrUnsupportedFormat         = validation.ErrUnsupportedFormat
	ErrInvalidJSON               = validation.ErrInvalidJSON
	ErrJSONSchemaViolation       = validation.ErrJSONSchemaViolation
	ErrMissingFeature            = validation.ErrMissingFeature
	ErrMissingRule               = validation.ErrMissingRule
	ErrMissingScenario           = validation.ErrMissingScenario
	ErrMultipleFeatures          = validation.ErrMultipleFeatures
	ErrRuleBeforeFeature         = validation.ErrRuleBeforeFeature
	ErrScenarioBeforeFeature     = validation.ErrScenarioBeforeFeature
	ErrBackgroundBeforeFeature   = validation.ErrBackgroundBeforeFeature
	ErrExamplesWithoutScenario   = validation.ErrExamplesWithoutScenario
	ErrRuleWithoutScenarios      = validation.ErrRuleWithoutScenarios
	ErrScenariosOutsideRule      = validation.ErrScenariosOutsideRule
	ErrMissingVerificationTag    = validation.ErrMissingVerificationTag
	ErrMissingRequiredElement    = validation.ErrMissingRequiredElement
	ErrInvalidSemanticType       = validation.ErrInvalidSemanticType
	ErrInvalidFeatureNaming      = validation.ErrInvalidFeatureNaming
	ErrIncorrectIndentation      = validation.ErrIncorrectIndentation
	ErrInvalidTagFormat          = validation.ErrInvalidTagFormat
	ErrInvalidTagLevel           = validation.ErrInvalidTagLevel
	ErrLineTooLong               = validation.ErrLineTooLong
	ErrInvalidPattern            = validation.ErrInvalidPattern
	ErrTooManyRules              = validation.ErrTooManyRules
	ErrLargeRuleCount            = validation.ErrLargeRuleCount
	ErrTooFewRules               = validation.ErrTooFewRules
	ErrTooManyScenarios          = validation.ErrTooManyScenarios
	ErrLargeScenarioCount        = validation.ErrLargeScenarioCount
	ErrMutualExclusionViolation  = validation.ErrMutualExclusionViolation
	ErrCriticalAspectRequiresGxP = validation.ErrCriticalAspectRequiresGxP
	ErrForbiddenPrefix           = validation.ErrForbiddenPrefix
	ErrForbiddenContent          = validation.ErrForbiddenContent
	ErrUnknownTag                = validation.ErrUnknownTag
	ErrGxPMissingControl         = validation.ErrGxPMissingControl
	ErrContractNameMismatch      = validation.ErrContractNameMismatch
	ErrNoTagContract             = validation.ErrNoTagContract
	ErrContractVersionMismatch   = validation.ErrContractVersionMismatch

	// Commit message validation error codes
	ErrEmptyMessage            = validation.ErrEmptyMessage
	ErrInvalidHeaderFormat     = validation.ErrInvalidHeaderFormat
	ErrHeaderTooLong           = validation.ErrHeaderTooLong
	ErrHeaderTrailingPeriod    = validation.ErrHeaderTrailingPeriod
	ErrMissingAuditorSummary   = validation.ErrMissingAuditorSummary
	ErrMissingTopLevelBody     = validation.ErrMissingTopLevelBody
	ErrMissingModuleSection    = validation.ErrMissingModuleSection
	ErrMissingSubjectLine      = validation.ErrMissingSubjectLine
	ErrInvalidSubjectFormat    = validation.ErrInvalidSubjectFormat
	ErrSubjectTooLong          = validation.ErrSubjectTooLong
	ErrSubjectTrailingPeriod   = validation.ErrSubjectTrailingPeriod
	ErrUnclosedCodeBlock       = validation.ErrUnclosedCodeBlock
	ErrOrphanedDashesLine      = validation.ErrOrphanedDashesLine
	ErrMalformedModuleSection  = validation.ErrMalformedModuleSection
	ErrMissingModuleName       = validation.ErrMissingModuleName
	ErrMissingModuleDashes     = validation.ErrMissingModuleDashes
	ErrContractLoadError       = validation.ErrContractLoadError
	ErrEmptyModuleSection      = validation.ErrEmptyModuleSection
	ErrInvalidModuleName       = validation.ErrInvalidModuleName
	ErrMissingDashes           = validation.ErrMissingDashes
	ErrInvalidDashes           = validation.ErrInvalidDashes
	ErrMissingBody             = validation.ErrMissingBody
	ErrUnexpectedModuleSection = validation.ErrUnexpectedModuleSection
	ErrUnexpectedFileList      = validation.ErrUnexpectedFileList
)
