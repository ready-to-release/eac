package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorCode_String(t *testing.T) {
	tests := []struct {
		name string
		code ErrorCode
		want string
	}{
		{
			name: "returns code for structure error",
			code: ErrEmptyOutput,
			want: "EMPTY_OUTPUT",
		},
		{
			name: "returns code for semantic error",
			code: ErrMissingFeature,
			want: "MISSING_FEATURE",
		},
		{
			name: "returns code for format error",
			code: ErrInvalidFeatureNaming,
			want: "INVALID_FEATURE_NAMING",
		},
		{
			name: "returns code for constraint error",
			code: ErrTooManyRules,
			want: "TOO_MANY_RULES",
		},
		{
			name: "returns code for corruption error",
			code: ErrForbiddenPrefix,
			want: "FORBIDDEN_PREFIX",
		},
		{
			name: "returns empty string for zero-value ErrorCode",
			code: ErrorCode{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.code.String())
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantFound bool
		wantCode  string
	}{
		{
			name:      "finds EMPTY_OUTPUT",
			code:      "EMPTY_OUTPUT",
			wantFound: true,
			wantCode:  "EMPTY_OUTPUT",
		},
		{
			name:      "finds MISSING_FEATURE",
			code:      "MISSING_FEATURE",
			wantFound: true,
			wantCode:  "MISSING_FEATURE",
		},
		{
			name:      "finds OSCAL_INVALID_DOCUMENT",
			code:      "OSCAL_INVALID_DOCUMENT",
			wantFound: true,
			wantCode:  "OSCAL_INVALID_DOCUMENT",
		},
		{
			name:      "finds EMPTY_MESSAGE (commit error)",
			code:      "EMPTY_MESSAGE",
			wantFound: true,
			wantCode:  "EMPTY_MESSAGE",
		},
		{
			name:      "finds UNKNOWN_TAG (warning)",
			code:      "UNKNOWN_TAG",
			wantFound: true,
			wantCode:  "UNKNOWN_TAG",
		},
		{
			name:      "returns false for nonexistent code",
			code:      "NONEXISTENT_CODE",
			wantFound: false,
		},
		{
			name:      "returns false for empty string",
			code:      "",
			wantFound: false,
		},
		{
			name:      "is case sensitive - lowercase fails",
			code:      "empty_output",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ec, found := GetErrorCode(tt.code)

			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				require.NotNil(t, ec)
				assert.Equal(t, tt.wantCode, ec.Code)
			} else {
				assert.Nil(t, ec)
			}
		})
	}
}

func TestMustGetErrorCode(t *testing.T) {
	t.Run("returns error code for valid code", func(t *testing.T) {
		ec := MustGetErrorCode("EMPTY_OUTPUT")
		require.NotNil(t, ec)
		assert.Equal(t, "EMPTY_OUTPUT", ec.Code)
		assert.Equal(t, CategoryStructure, ec.Category)
	})

	t.Run("returns error code for OSCAL code", func(t *testing.T) {
		ec := MustGetErrorCode("OSCAL_MISSING_UUID")
		require.NotNil(t, ec)
		assert.Equal(t, "OSCAL_MISSING_UUID", ec.Code)
		assert.Equal(t, CategorySemantic, ec.Category)
	})

	t.Run("panics for nonexistent code", func(t *testing.T) {
		assert.Panics(t, func() {
			MustGetErrorCode("NONEXISTENT_CODE")
		})
	})

	t.Run("panics for empty string", func(t *testing.T) {
		assert.Panics(t, func() {
			MustGetErrorCode("")
		})
	})
}

func TestGetAllErrorCodes(t *testing.T) {
	t.Run("returns all registered codes", func(t *testing.T) {
		all := GetAllErrorCodes()

		// Verify it is non-empty
		assert.Greater(t, len(all), 0)

		// Spot-check known codes from different categories
		spotChecks := []string{
			"EMPTY_OUTPUT",
			"MISSING_FEATURE",
			"INVALID_FEATURE_NAMING",
			"TOO_MANY_RULES",
			"FORBIDDEN_PREFIX",
			"MISSING_WORKSPACE",
			"OSCAL_INVALID_DOCUMENT",
			"EMPTY_MESSAGE",
			"UNKNOWN_TAG",
		}

		for _, code := range spotChecks {
			_, exists := all[code]
			assert.True(t, exists, "expected code %q in registry", code)
		}
	})

	t.Run("returns a copy not the original", func(t *testing.T) {
		all := GetAllErrorCodes()
		originalLen := len(all)

		// Mutate the returned map
		all["INJECTED_CODE"] = &ErrorCode{Code: "INJECTED_CODE"}

		// Original registry should be unaffected
		fresh := GetAllErrorCodes()
		assert.Equal(t, originalLen, len(fresh))
		_, exists := fresh["INJECTED_CODE"]
		assert.False(t, exists, "mutation should not affect registry")
	})
}

func TestGetErrorCodesByCategory(t *testing.T) {
	tests := []struct {
		name           string
		category       ErrorCategory
		expectNonEmpty bool
		expectContains string
	}{
		{
			name:           "structure category",
			category:       CategoryStructure,
			expectNonEmpty: true,
			expectContains: "EMPTY_OUTPUT",
		},
		{
			name:           "semantic category",
			category:       CategorySemantic,
			expectNonEmpty: true,
			expectContains: "MISSING_FEATURE",
		},
		{
			name:           "format category",
			category:       CategoryFormat,
			expectNonEmpty: true,
			expectContains: "INVALID_FEATURE_NAMING",
		},
		{
			name:           "constraint category",
			category:       CategoryConstraint,
			expectNonEmpty: true,
			expectContains: "TOO_MANY_RULES",
		},
		{
			name:           "corruption category",
			category:       CategoryCorruption,
			expectNonEmpty: true,
			expectContains: "FORBIDDEN_PREFIX",
		},
		{
			name:           "nonexistent category returns empty",
			category:       ErrorCategory("nonexistent"),
			expectNonEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorCodesByCategory(tt.category)

			if tt.expectNonEmpty {
				assert.Greater(t, len(result), 0)

				// Verify all results have the right category
				for _, ec := range result {
					assert.Equal(t, tt.category, ec.Category,
						"code %q should have category %q", ec.Code, tt.category)
				}

				// Verify expected code is in the result
				found := false
				for _, ec := range result {
					if ec.Code == tt.expectContains {
						found = true
						break
					}
				}
				assert.True(t, found, "expected %q in results", tt.expectContains)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestGetErrorCodesBySeverity(t *testing.T) {
	tests := []struct {
		name           string
		severity       ErrorSeverity
		expectNonEmpty bool
		expectContains string
	}{
		{
			name:           "error severity",
			severity:       SeverityError,
			expectNonEmpty: true,
			expectContains: "EMPTY_OUTPUT",
		},
		{
			name:           "warning severity",
			severity:       SeverityWarning,
			expectNonEmpty: true,
			expectContains: "UNKNOWN_TAG",
		},
		{
			name:           "nonexistent severity returns empty",
			severity:       ErrorSeverity("critical"),
			expectNonEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorCodesBySeverity(tt.severity)

			if tt.expectNonEmpty {
				assert.Greater(t, len(result), 0)

				// Verify all results have the right severity
				for _, ec := range result {
					assert.Equal(t, tt.severity, ec.Severity,
						"code %q should have severity %q", ec.Code, tt.severity)
				}

				// Verify expected code is in the result
				found := false
				for _, ec := range result {
					if ec.Code == tt.expectContains {
						found = true
						break
					}
				}
				assert.True(t, found, "expected %q in results", tt.expectContains)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestErrorCodeRegistry_Integrity(t *testing.T) {
	t.Run("every registered code key matches the ErrorCode.Code field", func(t *testing.T) {
		all := GetAllErrorCodes()
		for key, ec := range all {
			assert.Equal(t, key, ec.Code,
				"registry key %q does not match ErrorCode.Code %q", key, ec.Code)
		}
	})

	t.Run("every registered code has a non-empty description", func(t *testing.T) {
		all := GetAllErrorCodes()
		for key, ec := range all {
			assert.NotEmpty(t, ec.Description,
				"error code %q has empty description", key)
		}
	})

	t.Run("every registered code has a valid category", func(t *testing.T) {
		validCategories := map[ErrorCategory]bool{
			CategoryStructure:  true,
			CategorySemantic:   true,
			CategoryFormat:     true,
			CategoryConstraint: true,
			CategoryCorruption: true,
		}

		all := GetAllErrorCodes()
		for key, ec := range all {
			assert.True(t, validCategories[ec.Category],
				"error code %q has invalid category %q", key, ec.Category)
		}
	})

	t.Run("every registered code has a valid severity", func(t *testing.T) {
		validSeverities := map[ErrorSeverity]bool{
			SeverityError:   true,
			SeverityWarning: true,
		}

		all := GetAllErrorCodes()
		for key, ec := range all {
			assert.True(t, validSeverities[ec.Severity],
				"error code %q has invalid severity %q", key, ec.Severity)
		}
	})

	t.Run("warning severity codes exist", func(t *testing.T) {
		warnings := GetErrorCodesBySeverity(SeverityWarning)
		assert.Greater(t, len(warnings), 0, "expected at least one warning-level code")
	})

	t.Run("non-retriable codes exist", func(t *testing.T) {
		all := GetAllErrorCodes()
		nonRetriable := 0
		for _, ec := range all {
			if !ec.Retriable {
				nonRetriable++
			}
		}
		assert.Greater(t, nonRetriable, 0, "expected at least one non-retriable code")
	})
}

func TestErrorCodeProperties(t *testing.T) {
	t.Run("ErrEmptyOutput is non-retriable structure error", func(t *testing.T) {
		assert.Equal(t, "EMPTY_OUTPUT", ErrEmptyOutput.Code)
		assert.Equal(t, CategoryStructure, ErrEmptyOutput.Category)
		assert.False(t, ErrEmptyOutput.Retriable)
		assert.Equal(t, SeverityError, ErrEmptyOutput.Severity)
	})

	t.Run("ErrInvalidJSON is retriable structure error", func(t *testing.T) {
		assert.Equal(t, "INVALID_JSON", ErrInvalidJSON.Code)
		assert.Equal(t, CategoryStructure, ErrInvalidJSON.Category)
		assert.True(t, ErrInvalidJSON.Retriable)
		assert.Equal(t, SeverityError, ErrInvalidJSON.Severity)
	})

	t.Run("ErrUnknownTag is warning severity", func(t *testing.T) {
		assert.Equal(t, "UNKNOWN_TAG", ErrUnknownTag.Code)
		assert.Equal(t, SeverityWarning, ErrUnknownTag.Severity)
		assert.True(t, ErrUnknownTag.Retriable)
	})

	t.Run("ErrSetupError is non-retriable", func(t *testing.T) {
		assert.Equal(t, "SETUP_ERROR", ErrSetupError.Code)
		assert.False(t, ErrSetupError.Retriable)
		assert.Equal(t, SeverityError, ErrSetupError.Severity)
	})

	t.Run("ErrQuickValidationFailed is warning and non-retriable", func(t *testing.T) {
		assert.Equal(t, "QUICK_VALIDATION_FAILED", ErrQuickValidationFailed.Code)
		assert.False(t, ErrQuickValidationFailed.Retriable)
		assert.Equal(t, SeverityWarning, ErrQuickValidationFailed.Severity)
	})

	t.Run("ErrOSCALInvalidControlID is warning severity", func(t *testing.T) {
		assert.Equal(t, "OSCAL_INVALID_CONTROL_ID", ErrOSCALInvalidControlID.Code)
		assert.Equal(t, SeverityWarning, ErrOSCALInvalidControlID.Severity)
		assert.Equal(t, CategoryFormat, ErrOSCALInvalidControlID.Category)
	})
}
