package validation

import "fmt"

// ErrorCode represents a structured validation error code with metadata.
type ErrorCode struct {
	Code        string
	Category    ErrorCategory
	Retriable   bool
	Severity    ErrorSeverity
	Description string
}

// ErrorCategory represents the category of a validation error.
type ErrorCategory string

const (
	CategoryStructure  ErrorCategory = "structure"  // Critical structural issues (empty output, missing config)
	CategorySemantic   ErrorCategory = "semantic"   // Semantic/logical issues (missing elements, wrong order)
	CategoryFormat     ErrorCategory = "format"     // Format issues (naming, patterns, indentation)
	CategoryConstraint ErrorCategory = "constraint" // Constraint violations (size limits, counts)
	CategoryCorruption ErrorCategory = "corruption" // AI noise corruption (forbidden content)
)

// ErrorSeverity represents the severity level of an error.
type ErrorSeverity string

const (
	SeverityError   ErrorSeverity = "error"
	SeverityWarning ErrorSeverity = "warning"
)

// String returns the error code string.
func (e ErrorCode) String() string {
	return e.Code
}

// ===============================================
// Structure Errors (Critical, Non-Retriable)
// ===============================================

var (
	// ErrEmptyOutput indicates AI generated empty output.
	ErrEmptyOutput = ErrorCode{
		Code:        "EMPTY_OUTPUT",
		Category:    CategoryStructure,
		Retriable:   false,
		Severity:    SeverityError,
		Description: "AI generated empty output",
	}

	// ErrMissingTagsConfig indicates testing tags configuration not loaded.
	ErrMissingTagsConfig = ErrorCode{
		Code:        "MISSING_TAGS_CONFIG",
		Category:    CategoryStructure,
		Retriable:   false,
		Severity:    SeverityError,
		Description: "Testing tags configuration not loaded",
	}

	// ErrNoContract indicates validation contract not loaded.
	ErrNoContract = ErrorCode{
		Code:        "NO_CONTRACT",
		Category:    CategoryStructure,
		Retriable:   false,
		Severity:    SeverityError,
		Description: "Validation contract not loaded",
	}

	// ErrUnsupportedFormat indicates unsupported output format specified.
	ErrUnsupportedFormat = ErrorCode{
		Code:        "UNSUPPORTED_FORMAT",
		Category:    CategoryStructure,
		Retriable:   false,
		Severity:    SeverityError,
		Description: "Unsupported output format specified",
	}

	// ErrInvalidJSON indicates invalid JSON syntax.
	ErrInvalidJSON = ErrorCode{
		Code:        "INVALID_JSON",
		Category:    CategoryStructure,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Invalid JSON syntax",
	}

	// ErrJSONSchemaViolation indicates JSON schema validation failure.
	ErrJSONSchemaViolation = ErrorCode{
		Code:        "JSON_SCHEMA_VIOLATION",
		Category:    CategoryStructure,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "JSON schema validation failure",
	}
)

// ===============================================
// Semantic Errors (Retriable)
// ===============================================

var (
	// ErrMissingFeature indicates missing Feature declaration.
	ErrMissingFeature = ErrorCode{
		Code:        "MISSING_FEATURE",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Missing Feature declaration",
	}

	// ErrMissingRule indicates missing Rule declaration.
	ErrMissingRule = ErrorCode{
		Code:        "MISSING_RULE",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Missing Rule declaration",
	}

	// ErrMissingScenario indicates missing Scenario declaration.
	ErrMissingScenario = ErrorCode{
		Code:        "MISSING_SCENARIO",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Missing Scenario declaration",
	}

	// ErrMultipleFeatures indicates multiple Feature declarations.
	ErrMultipleFeatures = ErrorCode{
		Code:        "MULTIPLE_FEATURES",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Multiple Feature declarations found",
	}

	// ErrRuleBeforeFeature indicates Rule declared before Feature.
	ErrRuleBeforeFeature = ErrorCode{
		Code:        "RULE_BEFORE_FEATURE",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Rule must come after Feature declaration",
	}

	// ErrScenarioBeforeFeature indicates Scenario declared before Feature.
	ErrScenarioBeforeFeature = ErrorCode{
		Code:        "SCENARIO_BEFORE_FEATURE",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Scenario must come after Feature declaration",
	}

	// ErrBackgroundBeforeFeature indicates Background declared before Feature.
	ErrBackgroundBeforeFeature = ErrorCode{
		Code:        "BACKGROUND_BEFORE_FEATURE",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Background must come after Feature declaration",
	}

	// ErrExamplesWithoutScenario indicates Examples without Scenario Outline.
	ErrExamplesWithoutScenario = ErrorCode{
		Code:        "EXAMPLES_WITHOUT_SCENARIO",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Examples must come after Scenario Outline",
	}

	// ErrRuleWithoutScenarios indicates Rule has no scenarios.
	ErrRuleWithoutScenarios = ErrorCode{
		Code:        "RULE_WITHOUT_SCENARIOS",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Rule has no scenarios nested under it",
	}

	// ErrScenariosOutsideRule indicates scenarios not nested under Rule.
	ErrScenariosOutsideRule = ErrorCode{
		Code:        "SCENARIOS_OUTSIDE_RULE",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Scenarios must be nested under Rule blocks",
	}

	// ErrMissingVerificationTag indicates scenario missing verification tag.
	ErrMissingVerificationTag = ErrorCode{
		Code:        "MISSING_VERIFICATION_TAG",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Scenario missing required verification tag",
	}

	// ErrMissingRequiredElement indicates required element not found.
	ErrMissingRequiredElement = ErrorCode{
		Code:        "MISSING_REQUIRED_ELEMENT",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Required element not found in output",
	}

	// ErrInvalidSemanticType indicates invalid semantic type (e.g., commit type)
	// RESERVED: Planned for commit message type validation (feat, fix, docs, etc.)
	ErrInvalidSemanticType = ErrorCode{
		Code:        "INVALID_SEMANTIC_TYPE",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Invalid semantic type",
	}
)

// ===============================================
// Format Errors (Retriable)
// ===============================================

var (
	// ErrInvalidFeatureNaming indicates feature name doesn't follow convention.
	ErrInvalidFeatureNaming = ErrorCode{
		Code:        "INVALID_FEATURE_NAMING",
		Category:    CategoryFormat,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Feature name doesn't follow naming convention",
	}

	// ErrIncorrectIndentation indicates incorrect indentation.
	ErrIncorrectIndentation = ErrorCode{
		Code:        "INCORRECT_INDENTATION",
		Category:    CategoryFormat,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Incorrect indentation",
	}

	// ErrInvalidTagFormat indicates invalid tag format.
	ErrInvalidTagFormat = ErrorCode{
		Code:        "INVALID_TAG_FORMAT",
		Category:    CategoryFormat,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Invalid tag format",
	}

	// ErrInvalidTagLevel indicates tag used at wrong level.
	ErrInvalidTagLevel = ErrorCode{
		Code:        "INVALID_TAG_LEVEL",
		Category:    CategoryFormat,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Tag used at wrong level (feature vs scenario)",
	}

	// ErrLineTooLong indicates line exceeds maximum length.
	ErrLineTooLong = ErrorCode{
		Code:        "LINE_TOO_LONG",
		Category:    CategoryFormat,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Line exceeds maximum length",
	}

	// ErrInvalidPattern indicates content doesn't match required pattern.
	ErrInvalidPattern = ErrorCode{
		Code:        "INVALID_PATTERN",
		Category:    CategoryFormat,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Content doesn't match required pattern",
	}
)

// ===============================================
// Constraint Errors (Retriable)
// ===============================================

var (
	// ErrTooManyRules indicates too many rules (>10).
	ErrTooManyRules = ErrorCode{
		Code:        "TOO_MANY_RULES",
		Category:    CategoryConstraint,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Too many Rules (must split feature)",
	}

	// ErrLargeRuleCount indicates large number of rules (>6).
	ErrLargeRuleCount = ErrorCode{
		Code:        "LARGE_RULE_COUNT",
		Category:    CategoryConstraint,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Large number of Rules (consider splitting)",
	}

	// ErrTooFewRules indicates too few rules (<2).
	ErrTooFewRules = ErrorCode{
		Code:        "TOO_FEW_RULES",
		Category:    CategoryConstraint,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Too few Rules (2-6 recommended)",
	}

	// ErrTooManyScenarios indicates too many scenarios (>30).
	ErrTooManyScenarios = ErrorCode{
		Code:        "TOO_MANY_SCENARIOS",
		Category:    CategoryConstraint,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Too many Scenarios (must split feature)",
	}

	// ErrLargeScenarioCount indicates large number of scenarios (>20).
	ErrLargeScenarioCount = ErrorCode{
		Code:        "LARGE_SCENARIO_COUNT",
		Category:    CategoryConstraint,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Large number of Scenarios (consider splitting)",
	}

	// ErrMutualExclusionViolation indicates mutually exclusive tags used together.
	ErrMutualExclusionViolation = ErrorCode{
		Code:        "MUTUAL_EXCLUSION_VIOLATION",
		Category:    CategoryConstraint,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Mutually exclusive tags used together",
	}

	// ErrCriticalAspectRequiresGxP indicates @gmp-critical-aspect requires @gxp.
	ErrCriticalAspectRequiresGxP = ErrorCode{
		Code:        "CRITICAL_ASPECT_REQUIRES_GXP",
		Category:    CategoryConstraint,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "@gmp-critical-aspect requires @gxp tag",
	}
)

// ===============================================
// Corruption Errors (Retriable)
// RESERVED for AI safety and content filtering features
// ===============================================

var (
	// ErrForbiddenPrefix indicates output contains forbidden prefix
	// RESERVED: Planned for AI safety filtering to detect harmful output patterns.
	ErrForbiddenPrefix = ErrorCode{
		Code:        "FORBIDDEN_PREFIX",
		Category:    CategoryCorruption,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Output contains forbidden prefix",
	}

	// ErrForbiddenContent indicates output contains forbidden content
	// RESERVED: Planned for AI safety filtering to detect harmful content.
	ErrForbiddenContent = ErrorCode{
		Code:        "FORBIDDEN_CONTENT",
		Category:    CategoryCorruption,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Output contains forbidden content",
	}
)

// ===============================================
// Structurizr DSL Errors (Retriable)
// ===============================================

var (
	// ErrMissingWorkspace indicates DSL missing workspace keyword.
	ErrMissingWorkspace = ErrorCode{
		Code:        "MISSING_WORKSPACE",
		Category:    CategoryStructure,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "DSL must start with 'workspace' keyword",
	}

	// ErrUnmatchedBrace indicates unmatched closing brace.
	ErrUnmatchedBrace = ErrorCode{
		Code:        "UNMATCHED_BRACE",
		Category:    CategoryStructure,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Unmatched closing brace",
	}

	// ErrUnclosedBrace indicates unclosed opening brace.
	ErrUnclosedBrace = ErrorCode{
		Code:        "UNCLOSED_BRACE",
		Category:    CategoryStructure,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Unclosed opening brace",
	}

	// ErrInvalidIdentifier indicates invalid identifier name.
	ErrInvalidIdentifier = ErrorCode{
		Code:        "INVALID_IDENTIFIER",
		Category:    CategoryFormat,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Invalid identifier name",
	}

	// ErrParentChildRelationship indicates parent-child relationship error.
	ErrParentChildRelationship = ErrorCode{
		Code:        "PARENT_CHILD_RELATIONSHIP",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Relationships cannot be added between parents and children",
	}

	// ErrDSLValidation indicates general DSL validation error.
	ErrDSLValidation = ErrorCode{
		Code:        "DSL_VALIDATION",
		Category:    CategoryStructure,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "DSL validation error",
	}

	// ErrSetupError indicates setup/initialization error.
	ErrSetupError = ErrorCode{
		Code:        "SETUP_ERROR",
		Category:    CategoryStructure,
		Retriable:   false,
		Severity:    SeverityError,
		Description: "Setup or initialization error",
	}

	// ErrValidationError indicates validation execution error.
	ErrValidationError = ErrorCode{
		Code:        "VALIDATION_ERROR",
		Category:    CategoryStructure,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Validation execution error",
	}

	// ErrQuickValidationFailed indicates quick validation skipped full validation.
	ErrQuickValidationFailed = ErrorCode{
		Code:        "QUICK_VALIDATION_FAILED",
		Category:    CategoryStructure,
		Retriable:   false,
		Severity:    SeverityWarning,
		Description: "Skipping full validation due to syntax errors",
	}
)

// ===============================================
// OSCAL Errors (Retriable)
// ===============================================

var (
	// ErrOSCALInvalidDocument indicates document is not an OSCAL document type.
	ErrOSCALInvalidDocument = ErrorCode{
		Code:        "OSCAL_INVALID_DOCUMENT",
		Category:    CategoryStructure,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Document is not a valid OSCAL document type",
	}

	// ErrOSCALMissingUUID indicates missing required UUID field.
	ErrOSCALMissingUUID = ErrorCode{
		Code:        "OSCAL_MISSING_UUID",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Missing required UUID field",
	}

	// ErrOSCALMissingTitle indicates missing required metadata title.
	ErrOSCALMissingTitle = ErrorCode{
		Code:        "OSCAL_MISSING_TITLE",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Missing required metadata title",
	}

	// ErrOSCALMissingLastModified indicates missing required last-modified timestamp.
	ErrOSCALMissingLastModified = ErrorCode{
		Code:        "OSCAL_MISSING_LAST_MODIFIED",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Missing required metadata last-modified timestamp",
	}

	// ErrOSCALEmptyCatalog indicates catalog has no controls or groups.
	ErrOSCALEmptyCatalog = ErrorCode{
		Code:        "OSCAL_EMPTY_CATALOG",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Catalog must have at least one control or group",
	}

	// ErrOSCALMissingImports indicates profile has no imports.
	ErrOSCALMissingImports = ErrorCode{
		Code:        "OSCAL_MISSING_IMPORTS",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Profile must have at least one import",
	}

	// ErrOSCALMissingImportHref indicates import missing href field.
	ErrOSCALMissingImportHref = ErrorCode{
		Code:        "OSCAL_MISSING_IMPORT_HREF",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityError,
		Description: "Import missing required href field",
	}

	// ErrOSCALInvalidControlID indicates control ID may not be valid format.
	ErrOSCALInvalidControlID = ErrorCode{
		Code:        "OSCAL_INVALID_CONTROL_ID",
		Category:    CategoryFormat,
		Retriable:   true,
		Severity:    SeverityWarning,
		Description: "Control ID may not be valid NIST 800-53 format",
	}

	// ErrOSCALEmptyImport indicates import has no control selections.
	ErrOSCALEmptyImport = ErrorCode{
		Code:        "OSCAL_EMPTY_IMPORT",
		Category:    CategorySemantic,
		Retriable:   true,
		Severity:    SeverityWarning,
		Description: "Import has no control selections",
	}

	// ErrOSCALVersionMismatch indicates OSCAL version differs from expected.
	ErrOSCALVersionMismatch = ErrorCode{
		Code:        "OSCAL_VERSION_MISMATCH",
		Category:    CategoryStructure,
		Retriable:   true,
		Severity:    SeverityWarning,
		Description: "OSCAL version differs from expected version",
	}

	// ErrOSCALFileRead indicates failed to read file.
	ErrOSCALFileRead = ErrorCode{
		Code:        "OSCAL_FILE_READ",
		Category:    CategoryStructure,
		Retriable:   false,
		Severity:    SeverityError,
		Description: "Failed to read OSCAL file",
	}
)

// ===============================================
// Warning-Level Errors
// ===============================================

var (
	// ErrUnknownTag indicates unknown tag not in testing-tags.yml (warning).
	ErrUnknownTag = ErrorCode{
		Code:        "UNKNOWN_TAG",
		Category:    CategoryFormat,
		Retriable:   true,
		Severity:    SeverityWarning,
		Description: "Unknown tag not defined in testing-tags.yml",
	}

	// ErrGxPMissingControl indicates @gxp without control tag (warning).
	ErrGxPMissingControl = ErrorCode{
		Code:        "GXP_MISSING_CONTROL",
		Category:    CategoryConstraint,
		Retriable:   true,
		Severity:    SeverityWarning,
		Description: "@gxp tag should link to OSCAL controls",
	}

	// ErrContractNameMismatch indicates contract name mismatch (warning).
	ErrContractNameMismatch = ErrorCode{
		Code:        "CONTRACT_NAME_MISMATCH",
		Category:    CategoryStructure,
		Retriable:   false,
		Severity:    SeverityWarning,
		Description: "Contract name doesn't match expected value",
	}

	// ErrNoTagContract indicates tag contract not loaded (warning).
	ErrNoTagContract = ErrorCode{
		Code:        "NO_TAG_CONTRACT",
		Category:    CategoryStructure,
		Retriable:   false,
		Severity:    SeverityWarning,
		Description: "Tag contract not loaded",
	}

	// ErrContractVersionMismatch indicates contract version mismatch (warning).
	ErrContractVersionMismatch = ErrorCode{
		Code:        "CONTRACT_VERSION_MISMATCH",
		Category:    CategoryStructure,
		Retriable:   false,
		Severity:    SeverityWarning,
		Description: "Contract version doesn't match expected version",
	}
)

// ===============================================
// Error Code Registry and Lookup
// ===============================================

var errorCodeRegistry = map[string]*ErrorCode{
	// Structure errors
	"EMPTY_OUTPUT":          &ErrEmptyOutput,
	"MISSING_TAGS_CONFIG":   &ErrMissingTagsConfig,
	"NO_CONTRACT":           &ErrNoContract,
	"UNSUPPORTED_FORMAT":    &ErrUnsupportedFormat,
	"INVALID_JSON":          &ErrInvalidJSON,
	"JSON_SCHEMA_VIOLATION": &ErrJSONSchemaViolation,

	// Semantic errors
	"MISSING_FEATURE":           &ErrMissingFeature,
	"MISSING_RULE":              &ErrMissingRule,
	"MISSING_SCENARIO":          &ErrMissingScenario,
	"MULTIPLE_FEATURES":         &ErrMultipleFeatures,
	"RULE_BEFORE_FEATURE":       &ErrRuleBeforeFeature,
	"SCENARIO_BEFORE_FEATURE":   &ErrScenarioBeforeFeature,
	"BACKGROUND_BEFORE_FEATURE": &ErrBackgroundBeforeFeature,
	"EXAMPLES_WITHOUT_SCENARIO": &ErrExamplesWithoutScenario,
	"RULE_WITHOUT_SCENARIOS":    &ErrRuleWithoutScenarios,
	"SCENARIOS_OUTSIDE_RULE":    &ErrScenariosOutsideRule,
	"MISSING_VERIFICATION_TAG":  &ErrMissingVerificationTag,
	"MISSING_REQUIRED_ELEMENT":  &ErrMissingRequiredElement,
	"INVALID_SEMANTIC_TYPE":     &ErrInvalidSemanticType,

	// Format errors
	"INVALID_FEATURE_NAMING": &ErrInvalidFeatureNaming,
	"INCORRECT_INDENTATION":  &ErrIncorrectIndentation,
	"INVALID_TAG_FORMAT":     &ErrInvalidTagFormat,
	"INVALID_TAG_LEVEL":      &ErrInvalidTagLevel,
	"LINE_TOO_LONG":          &ErrLineTooLong,
	"INVALID_PATTERN":        &ErrInvalidPattern,

	// Constraint errors
	"TOO_MANY_RULES":               &ErrTooManyRules,
	"LARGE_RULE_COUNT":             &ErrLargeRuleCount,
	"TOO_FEW_RULES":                &ErrTooFewRules,
	"TOO_MANY_SCENARIOS":           &ErrTooManyScenarios,
	"LARGE_SCENARIO_COUNT":         &ErrLargeScenarioCount,
	"MUTUAL_EXCLUSION_VIOLATION":   &ErrMutualExclusionViolation,
	"CRITICAL_ASPECT_REQUIRES_GXP": &ErrCriticalAspectRequiresGxP,

	// Corruption errors
	"FORBIDDEN_PREFIX":  &ErrForbiddenPrefix,
	"FORBIDDEN_CONTENT": &ErrForbiddenContent,

	// Structurizr DSL errors
	"MISSING_WORKSPACE":         &ErrMissingWorkspace,
	"UNMATCHED_BRACE":           &ErrUnmatchedBrace,
	"UNCLOSED_BRACE":            &ErrUnclosedBrace,
	"INVALID_IDENTIFIER":        &ErrInvalidIdentifier,
	"PARENT_CHILD_RELATIONSHIP": &ErrParentChildRelationship,
	"DSL_VALIDATION":            &ErrDSLValidation,
	"SETUP_ERROR":               &ErrSetupError,
	"VALIDATION_ERROR":          &ErrValidationError,
	"QUICK_VALIDATION_FAILED":   &ErrQuickValidationFailed,

	// OSCAL errors
	"OSCAL_INVALID_DOCUMENT":      &ErrOSCALInvalidDocument,
	"OSCAL_MISSING_UUID":          &ErrOSCALMissingUUID,
	"OSCAL_MISSING_TITLE":         &ErrOSCALMissingTitle,
	"OSCAL_MISSING_LAST_MODIFIED": &ErrOSCALMissingLastModified,
	"OSCAL_EMPTY_CATALOG":         &ErrOSCALEmptyCatalog,
	"OSCAL_MISSING_IMPORTS":       &ErrOSCALMissingImports,
	"OSCAL_MISSING_IMPORT_HREF":   &ErrOSCALMissingImportHref,
	"OSCAL_INVALID_CONTROL_ID":    &ErrOSCALInvalidControlID,
	"OSCAL_EMPTY_IMPORT":          &ErrOSCALEmptyImport,
	"OSCAL_VERSION_MISMATCH":      &ErrOSCALVersionMismatch,
	"OSCAL_FILE_READ":             &ErrOSCALFileRead,

	// Warning-level errors
	"UNKNOWN_TAG":               &ErrUnknownTag,
	"GXP_MISSING_CONTROL":       &ErrGxPMissingControl,
	"CONTRACT_NAME_MISMATCH":    &ErrContractNameMismatch,
	"NO_TAG_CONTRACT":           &ErrNoTagContract,
	"CONTRACT_VERSION_MISMATCH": &ErrContractVersionMismatch,
}

// GetErrorCode retrieves error code metadata by string code
// Returns (errorCode, found). If not found, returns nil and false.
func GetErrorCode(code string) (*ErrorCode, bool) {
	ec, ok := errorCodeRegistry[code]
	return ec, ok
}

// MustGetErrorCode retrieves error code metadata by string code
// Panics if the code is not found (for cases where code should always exist).
func MustGetErrorCode(code string) *ErrorCode {
	ec, ok := errorCodeRegistry[code]
	if !ok {
		panic(fmt.Sprintf("error code not found: %s", code))
	}
	return ec
}

// GetAllErrorCodes returns all registered error codes.
func GetAllErrorCodes() map[string]*ErrorCode {
	// Return a copy to prevent external modification
	result := make(map[string]*ErrorCode, len(errorCodeRegistry))
	for k, v := range errorCodeRegistry {
		result[k] = v
	}
	return result
}

// GetErrorCodesByCategory returns all error codes in a category.
func GetErrorCodesByCategory(category ErrorCategory) []*ErrorCode {
	var result []*ErrorCode
	for _, ec := range errorCodeRegistry {
		if ec.Category == category {
			result = append(result, ec)
		}
	}
	return result
}

// GetErrorCodesBySeverity returns all error codes with a severity level.
func GetErrorCodesBySeverity(severity ErrorSeverity) []*ErrorCode {
	var result []*ErrorCode
	for _, ec := range errorCodeRegistry {
		if ec.Severity == severity {
			result = append(result, ec)
		}
	}
	return result
}
