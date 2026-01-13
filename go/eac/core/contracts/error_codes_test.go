package contracts

import (
	"testing"
)

// TestGetErrorCode tests error code retrieval
func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantFound bool
		wantCode  string
	}{
		{
			name:      "valid structure error code",
			code:      "EMPTY_OUTPUT",
			wantFound: true,
			wantCode:  "EMPTY_OUTPUT",
		},
		{
			name:      "valid semantic error code",
			code:      "MISSING_FEATURE",
			wantFound: true,
			wantCode:  "MISSING_FEATURE",
		},
		{
			name:      "valid format error code",
			code:      "LINE_TOO_LONG",
			wantFound: true,
			wantCode:  "LINE_TOO_LONG",
		},
		{
			name:      "invalid error code",
			code:      "NONEXISTENT_ERROR",
			wantFound: false,
			wantCode:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ec, found := GetErrorCode(tt.code)
			if found != tt.wantFound {
				t.Errorf("GetErrorCode() found = %v, want %v", found, tt.wantFound)
			}
			if found && ec.Code != tt.wantCode {
				t.Errorf("GetErrorCode() code = %v, want %v", ec.Code, tt.wantCode)
			}
		})
	}
}

// TestErrorCodeMetadata tests error code metadata
func TestErrorCodeMetadata(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantCategory   ErrorCategory
		wantRetriable  bool
		wantSeverity   ErrorSeverity
	}{
		{
			name:           "EMPTY_OUTPUT is critical and non-retriable",
			code:           "EMPTY_OUTPUT",
			wantCategory:   CategoryStructure,
			wantRetriable:  false,
			wantSeverity:   SeverityError,
		},
		{
			name:           "MISSING_FEATURE is semantic and retriable",
			code:           "MISSING_FEATURE",
			wantCategory:   CategorySemantic,
			wantRetriable:  true,
			wantSeverity:   SeverityError,
		},
		{
			name:           "LINE_TOO_LONG is format and retriable",
			code:           "LINE_TOO_LONG",
			wantCategory:   CategoryFormat,
			wantRetriable:  true,
			wantSeverity:   SeverityError,
		},
		{
			name:           "UNKNOWN_TAG is warning",
			code:           "UNKNOWN_TAG",
			wantCategory:   CategoryFormat,
			wantRetriable:  true,
			wantSeverity:   SeverityWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ec, found := GetErrorCode(tt.code)
			if !found {
				t.Fatalf("Error code not found: %s", tt.code)
			}

			if ec.Category != tt.wantCategory {
				t.Errorf("Category = %v, want %v", ec.Category, tt.wantCategory)
			}
			if ec.Retriable != tt.wantRetriable {
				t.Errorf("Retriable = %v, want %v", ec.Retriable, tt.wantRetriable)
			}
			if ec.Severity != tt.wantSeverity {
				t.Errorf("Severity = %v, want %v", ec.Severity, tt.wantSeverity)
			}
		})
	}
}

// TestGetErrorCodesByCategory tests filtering by category
func TestGetErrorCodesByCategory(t *testing.T) {
	tests := []struct {
		name         string
		category     ErrorCategory
		wantMinCodes int // Minimum expected number of codes
	}{
		{
			name:         "structure category has codes",
			category:     CategoryStructure,
			wantMinCodes: 3,
		},
		{
			name:         "semantic category has codes",
			category:     CategorySemantic,
			wantMinCodes: 10,
		},
		{
			name:         "format category has codes",
			category:     CategoryFormat,
			wantMinCodes: 5,
		},
		{
			name:         "constraint category has codes",
			category:     CategoryConstraint,
			wantMinCodes: 5,
		},
		{
			name:         "corruption category has codes",
			category:     CategoryCorruption,
			wantMinCodes: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes := GetErrorCodesByCategory(tt.category)
			if len(codes) < tt.wantMinCodes {
				t.Errorf("GetErrorCodesByCategory() returned %d codes, want at least %d", len(codes), tt.wantMinCodes)
			}

			// Verify all returned codes match the category
			for _, ec := range codes {
				if ec.Category != tt.category {
					t.Errorf("Error code %s has category %v, expected %v", ec.Code, ec.Category, tt.category)
				}
			}
		})
	}
}

// TestGetErrorCodesBySeverity tests filtering by severity
func TestGetErrorCodesBySeverity(t *testing.T) {
	tests := []struct {
		name         string
		severity     ErrorSeverity
		wantMinCodes int
	}{
		{
			name:         "error severity has many codes",
			severity:     SeverityError,
			wantMinCodes: 20,
		},
		{
			name:         "warning severity has some codes",
			severity:     SeverityWarning,
			wantMinCodes: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes := GetErrorCodesBySeverity(tt.severity)
			if len(codes) < tt.wantMinCodes {
				t.Errorf("GetErrorCodesBySeverity() returned %d codes, want at least %d", len(codes), tt.wantMinCodes)
			}

			// Verify all returned codes match the severity
			for _, ec := range codes {
				if ec.Severity != tt.severity {
					t.Errorf("Error code %s has severity %v, expected %v", ec.Code, ec.Severity, tt.severity)
				}
			}
		})
	}
}

// TestValidationErrorWithStructuredCode tests ValidationError with structured code
func TestValidationErrorWithStructuredCode(t *testing.T) {
	tests := []struct {
		name        string
		errorCode   ErrorCode
		message     string
		line        int
		wantCode    string
		wantError   bool
		wantWarning bool
	}{
		{
			name:        "error with EMPTY_OUTPUT",
			errorCode:   ErrEmptyOutput,
			message:     "Test message",
			line:        0,
			wantCode:    "EMPTY_OUTPUT",
			wantError:   true,
			wantWarning: false,
		},
		{
			name:        "warning with UNKNOWN_TAG",
			errorCode:   ErrUnknownTag,
			message:     "Test warning",
			line:        10,
			wantCode:    "UNKNOWN_TAG",
			wantError:   false,
			wantWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verr := NewValidationError(tt.errorCode, tt.message, tt.line)

			if verr.GetCode() != tt.wantCode {
				t.Errorf("GetCode() = %v, want %v", verr.GetCode(), tt.wantCode)
			}
			if verr.Message != tt.message {
				t.Errorf("Message = %v, want %v", verr.Message, tt.message)
			}
			if verr.Line != tt.line {
				t.Errorf("Line = %v, want %v", verr.Line, tt.line)
			}
			if verr.IsCritical() == tt.errorCode.Retriable {
				t.Errorf("IsCritical() should be inverse of Retriable")
			}
			if verr.IsWarning() != tt.wantWarning {
				t.Errorf("IsWarning() = %v, want %v", verr.IsWarning(), tt.wantWarning)
			}
		})
	}
}

// TestLegacyValidationError tests backward compatibility
func TestLegacyValidationError(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		message     string
		line        int
		severity    string
		wantKnown   bool
	}{
		{
			name:      "known legacy code EMPTY_OUTPUT",
			code:      "EMPTY_OUTPUT",
			message:   "Test",
			line:      0,
			severity:  "error",
			wantKnown: true,
		},
		{
			name:      "unknown legacy code",
			code:      "CUSTOM_ERROR",
			message:   "Test",
			line:      0,
			severity:  "error",
			wantKnown: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verr := NewLegacyValidationError(tt.code, tt.message, tt.line, tt.severity)

			if verr.GetCode() != tt.code {
				t.Errorf("GetCode() = %v, want %v", verr.GetCode(), tt.code)
			}

			hasStructuredCode := verr.Code != nil
			if hasStructuredCode != tt.wantKnown {
				t.Errorf("Has structured code = %v, want %v", hasStructuredCode, tt.wantKnown)
			}
		})
	}
}
