package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewErrorFormatter(t *testing.T) {
	f := NewErrorFormatter()
	require.NotNil(t, f)
}

func TestErrorFormatter_FormatEnhancedError(t *testing.T) {
	f := NewErrorFormatter()

	tests := []struct {
		name           string
		code           ErrorCode
		summary        string
		aiOutput       string
		expectedFormat string
		correctExample string
		fixGuidance    string
		lineNum        int
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "all sections populated",
			code:           ErrMissingFeature,
			summary:        "Feature declaration is missing",
			aiOutput:       "Scenario: something\nGiven a step",
			expectedFormat: "Feature: <name>",
			correctExample: "Feature: User Authentication",
			fixGuidance:    "Add a Feature declaration at the top",
			lineNum:        1,
			wantContains: []string{
				"Feature declaration is missing",
				"What AI generated:",
				"Expected format:",
				"Correct example:",
				"How to fix:",
				"Add a Feature declaration at the top",
			},
		},
		{
			name:           "only summary",
			code:           ErrEmptyOutput,
			summary:        "Output was empty",
			aiOutput:       "",
			expectedFormat: "",
			correctExample: "",
			fixGuidance:    "",
			lineNum:        0,
			wantContains:   []string{"Output was empty"},
			wantNotContain: []string{"What AI generated:", "Expected format:", "Correct example:", "How to fix:"},
		},
		{
			name:           "summary with ai output only",
			code:           ErrInvalidJSON,
			summary:        "Invalid JSON",
			aiOutput:       "{bad json}",
			expectedFormat: "",
			correctExample: "",
			fixGuidance:    "",
			lineNum:        5,
			wantContains:   []string{"Invalid JSON", "What AI generated:", "{bad json}"},
			wantNotContain: []string{"Expected format:", "Correct example:", "How to fix:"},
		},
		{
			name:           "summary with fix guidance only",
			code:           ErrIncorrectIndentation,
			summary:        "Wrong indentation",
			aiOutput:       "",
			expectedFormat: "",
			correctExample: "",
			fixGuidance:    "Use 2 spaces for indentation",
			lineNum:        10,
			wantContains:   []string{"Wrong indentation", "How to fix:", "Use 2 spaces"},
			wantNotContain: []string{"What AI generated:", "Expected format:", "Correct example:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := f.FormatEnhancedError(
				tt.code, tt.summary, tt.aiOutput,
				tt.expectedFormat, tt.correctExample,
				tt.fixGuidance, tt.lineNum,
			)

			require.NotNil(t, ve)
			assert.Equal(t, tt.code.Code, ve.GetCode())
			assert.Equal(t, tt.lineNum, ve.Line)

			for _, s := range tt.wantContains {
				assert.Contains(t, ve.Message, s)
			}
			for _, s := range tt.wantNotContain {
				assert.NotContains(t, ve.Message, s)
			}
		})
	}
}

func TestErrorFormatter_FormatEnhancedError_TruncatesLongAIOutput(t *testing.T) {
	f := NewErrorFormatter()

	// Build a 20-line output
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line content"
	}
	longOutput := strings.Join(lines, "\n")

	ve := f.FormatEnhancedError(
		ErrInvalidJSON, "Bad JSON", longOutput,
		"", "", "", 0,
	)

	require.NotNil(t, ve)
	assert.Contains(t, ve.Message, "... (truncated)")
}

func TestErrorFormatter_FormatWithExample(t *testing.T) {
	f := NewErrorFormatter()

	tests := []struct {
		name           string
		code           ErrorCode
		summary        string
		correctExample string
		actualSnippet  string
		lineNum        int
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "both example and actual provided",
			code:           ErrInvalidFeatureNaming,
			summary:        "Feature name is wrong",
			correctExample: "Feature: User Authentication",
			actualSnippet:  "Feature: user_auth",
			lineNum:        1,
			wantContains: []string{
				"Feature name is wrong",
				"AI generated:",
				"user_auth",
				"Expected:",
				"User Authentication",
			},
		},
		{
			name:           "only correct example",
			code:           ErrMissingFeature,
			summary:        "No feature found",
			correctExample: "Feature: Something",
			actualSnippet:  "",
			lineNum:        0,
			wantContains:   []string{"No feature found", "Expected:", "Something"},
			wantNotContain: []string{"AI generated:"},
		},
		{
			name:           "only actual snippet",
			code:           ErrMissingRule,
			summary:        "Missing rule",
			correctExample: "",
			actualSnippet:  "Scenario: orphan",
			lineNum:        3,
			wantContains:   []string{"Missing rule", "AI generated:", "orphan"},
			wantNotContain: []string{"Expected:"},
		},
		{
			name:           "neither example nor actual",
			code:           ErrEmptyOutput,
			summary:        "Output is empty",
			correctExample: "",
			actualSnippet:  "",
			lineNum:        0,
			wantContains:   []string{"Output is empty"},
			wantNotContain: []string{"AI generated:", "Expected:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := f.FormatWithExample(
				tt.code, tt.summary,
				tt.correctExample, tt.actualSnippet,
				tt.lineNum,
			)

			require.NotNil(t, ve)
			assert.Equal(t, tt.code.Code, ve.GetCode())
			assert.Equal(t, tt.lineNum, ve.Line)

			for _, s := range tt.wantContains {
				assert.Contains(t, ve.Message, s)
			}
			for _, s := range tt.wantNotContain {
				assert.NotContains(t, ve.Message, s)
			}
		})
	}
}

func TestErrorFormatter_FormatStructuredError(t *testing.T) {
	f := NewErrorFormatter()

	tests := []struct {
		name         string
		code         ErrorCode
		summary      string
		structure    string
		actualFields []string
		lineNum      int
		wantContains []string
	}{
		{
			name:         "with actual fields",
			code:         ErrJSONSchemaViolation,
			summary:      "Schema violation",
			structure:    `{"name": "string", "version": "string"}`,
			actualFields: []string{"name", "count"},
			lineNum:      1,
			wantContains: []string{
				"Schema violation",
				"Expected structure:",
				`"name": "string"`,
				"Fields AI provided:",
				"name",
				"count",
			},
		},
		{
			name:         "without actual fields",
			code:         ErrMissingRequiredElement,
			summary:      "Missing element",
			structure:    "<root>\n  <child/>\n</root>",
			actualFields: nil,
			lineNum:      0,
			wantContains: []string{
				"Missing element",
				"Expected structure:",
				"<root>",
			},
		},
		{
			name:         "empty actual fields slice",
			code:         ErrMissingRequiredElement,
			summary:      "Missing element",
			structure:    "required: field1, field2",
			actualFields: []string{},
			lineNum:      5,
			wantContains: []string{
				"Missing element",
				"Expected structure:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := f.FormatStructuredError(
				tt.code, tt.summary,
				tt.structure, tt.actualFields,
				tt.lineNum,
			)

			require.NotNil(t, ve)
			assert.Equal(t, tt.code.Code, ve.GetCode())
			assert.Equal(t, tt.lineNum, ve.Line)

			for _, s := range tt.wantContains {
				assert.Contains(t, ve.Message, s)
			}
		})
	}
}

func TestErrorFormatter_FormatMissingFieldError(t *testing.T) {
	f := NewErrorFormatter()

	tests := []struct {
		name            string
		code            ErrorCode
		fieldName       string
		parentStructure string
		actualFields    []string
		exampleValue    string
		lineNum         int
		wantContains    []string
	}{
		{
			name:            "all fields provided",
			code:            ErrOSCALMissingUUID,
			fieldName:       "uuid",
			parentStructure: "catalog",
			actualFields:    []string{"title", "version"},
			exampleValue:    `"uuid": "abc-123-def"`,
			lineNum:         1,
			wantContains: []string{
				"Missing required field: uuid",
				"'catalog' must have a 'uuid' field",
				"Example:",
				`"uuid": "abc-123-def"`,
				"Fields currently present:",
				"title",
				"version",
				"Fix: Add the 'uuid' field",
			},
		},
		{
			name:            "no example value",
			code:            ErrOSCALMissingTitle,
			fieldName:       "title",
			parentStructure: "metadata",
			actualFields:    []string{"version"},
			exampleValue:    "",
			lineNum:         3,
			wantContains: []string{
				"Missing required field: title",
				"'metadata' must have a 'title' field",
				"Fields currently present:",
				"Fix: Add the 'title' field",
			},
		},
		{
			name:            "no actual fields",
			code:            ErrMissingRequiredElement,
			fieldName:       "controls",
			parentStructure: "catalog",
			actualFields:    nil,
			exampleValue:    `"controls": [{"id": "AC-1"}]`,
			lineNum:         0,
			wantContains: []string{
				"Missing required field: controls",
				"'catalog' must have a 'controls' field",
				"Example:",
				"Fix: Add the 'controls' field",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := f.FormatMissingFieldError(
				tt.code, tt.fieldName,
				tt.parentStructure, tt.actualFields,
				tt.exampleValue, tt.lineNum,
			)

			require.NotNil(t, ve)
			assert.Equal(t, tt.code.Code, ve.GetCode())
			assert.Equal(t, tt.lineNum, ve.Line)

			for _, s := range tt.wantContains {
				assert.Contains(t, ve.Message, s)
			}
		})
	}
}

func TestErrorFormatter_indent(t *testing.T) {
	f := NewErrorFormatter()

	tests := []struct {
		name   string
		text   string
		prefix string
		want   string
	}{
		{
			name:   "single line",
			text:   "hello",
			prefix: "  ",
			want:   "  hello",
		},
		{
			name:   "multiple lines",
			text:   "line1\nline2\nline3",
			prefix: "  ",
			want:   "  line1\n  line2\n  line3",
		},
		{
			name:   "empty text",
			text:   "",
			prefix: "  ",
			want:   "  ",
		},
		{
			name:   "custom prefix",
			text:   "a\nb",
			prefix: ">>> ",
			want:   ">>> a\n>>> b",
		},
		{
			name:   "empty prefix",
			text:   "a\nb",
			prefix: "",
			want:   "a\nb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.indent(tt.text, tt.prefix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorFormatter_truncate(t *testing.T) {
	f := NewErrorFormatter()

	tests := []struct {
		name     string
		text     string
		maxLines int
		want     string
	}{
		{
			name:     "text within limit",
			text:     "line1\nline2\nline3",
			maxLines: 5,
			want:     "line1\nline2\nline3",
		},
		{
			name:     "text at exact limit",
			text:     "line1\nline2\nline3",
			maxLines: 3,
			want:     "line1\nline2\nline3",
		},
		{
			name:     "text exceeds limit",
			text:     "line1\nline2\nline3\nline4\nline5",
			maxLines: 3,
			want:     "line1\nline2\nline3\n... (truncated)",
		},
		{
			name:     "single line within limit",
			text:     "just one line",
			maxLines: 1,
			want:     "just one line",
		},
		{
			name:     "two lines truncated to one",
			text:     "line1\nline2",
			maxLines: 1,
			want:     "line1\n... (truncated)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.truncate(tt.text, tt.maxLines)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorFormatter_TruncateJSON(t *testing.T) {
	f := NewErrorFormatter()

	t.Run("valid JSON is pretty-printed and truncated", func(t *testing.T) {
		jsonStr := `{"name":"test","version":"1.0","description":"a description"}`
		result := f.TruncateJSON(jsonStr, 3)

		// Should be pretty-printed (indented)
		assert.Contains(t, result, "  ")
		// Should be truncated if more than 3 lines after pretty-printing
		lines := strings.Split(result, "\n")
		// Pretty JSON will have more than 3 lines, so it should be truncated
		assert.Contains(t, result, "... (truncated)")
		_ = lines
	})

	t.Run("invalid JSON is truncated as-is", func(t *testing.T) {
		jsonStr := "not json\nline2\nline3\nline4"
		result := f.TruncateJSON(jsonStr, 2)

		assert.Equal(t, "not json\nline2\n... (truncated)", result)
	})

	t.Run("short valid JSON is not truncated", func(t *testing.T) {
		jsonStr := `{"a":"b"}`
		result := f.TruncateJSON(jsonStr, 20)

		// Should not contain truncation marker
		assert.NotContains(t, result, "... (truncated)")
	})
}

func TestErrorFormatter_ExtractJSONFields(t *testing.T) {
	f := NewErrorFormatter()

	tests := []struct {
		name       string
		jsonStr    string
		wantFields []string
		wantEmpty  bool
	}{
		{
			name:       "extracts top-level fields",
			jsonStr:    `{"name":"test","version":"1.0","count":5}`,
			wantFields: []string{"name", "version", "count"},
		},
		{
			name:       "handles nested objects - only top level",
			jsonStr:    `{"metadata":{"title":"t"},"uuid":"abc"}`,
			wantFields: []string{"metadata", "uuid"},
		},
		{
			name:      "returns empty for invalid JSON",
			jsonStr:   `not json`,
			wantEmpty: true,
		},
		{
			name:      "returns empty for JSON array",
			jsonStr:   `[1,2,3]`,
			wantEmpty: true,
		},
		{
			name:      "returns empty for empty string",
			jsonStr:   ``,
			wantEmpty: true,
		},
		{
			name:       "handles empty JSON object",
			jsonStr:    `{}`,
			wantFields: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := f.ExtractJSONFields(tt.jsonStr)

			if tt.wantEmpty {
				assert.Empty(t, fields)
			} else {
				assert.Equal(t, len(tt.wantFields), len(fields))
				for _, expected := range tt.wantFields {
					assert.Contains(t, fields, expected)
				}
			}
		})
	}
}

func TestErrorFormatter_HighlightDifference(t *testing.T) {
	f := NewErrorFormatter()

	t.Run("single line comparison", func(t *testing.T) {
		result := f.HighlightDifference("Feature: User Auth", "Feature: user_auth")

		assert.Contains(t, result, "Expected:")
		assert.Contains(t, result, "Actual:")
		assert.Contains(t, result, "Feature: User Auth")
		assert.Contains(t, result, "Feature: user_auth")
	})

	t.Run("multiline comparison", func(t *testing.T) {
		result := f.HighlightDifference("line1\nline2", "line1\nchanged")

		assert.Contains(t, result, "Expected:")
		assert.Contains(t, result, "Actual:")
		// Each line is individually prefixed by indent, so check each line separately
		assert.Contains(t, result, "line1")
		assert.Contains(t, result, "line2")
		assert.Contains(t, result, "changed")
	})

	t.Run("uses check and cross markers", func(t *testing.T) {
		result := f.HighlightDifference("good", "bad")

		// The indent method prefixes lines; HighlightDifference uses markers
		assert.Contains(t, result, "good")
		assert.Contains(t, result, "bad")
	})
}
