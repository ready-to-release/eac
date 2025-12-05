package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenderer(t *testing.T) {
	renderer := NewRenderer("test.md")
	assert.NotNil(t, renderer)
	assert.Equal(t, "test.md", renderer.templatePath)
	assert.NotNil(t, renderer.customFuncs)
	assert.Equal(t, "zero", renderer.missingKeyMode)
}

func TestRendererWithFuncs(t *testing.T) {
	renderer := NewRenderer("test.md").
		WithFuncs(map[string]interface{}{
			"custom": func() string { return "test" },
		})

	assert.NotNil(t, renderer.customFuncs["custom"])
	assert.NotNil(t, renderer.customFuncs["join"]) // Default funcs still present
}

func TestRendererWithMissingKeyMode(t *testing.T) {
	renderer := NewRenderer("test.md").
		WithMissingKeyMode("error")

	assert.Equal(t, "error", renderer.missingKeyMode)
}

func TestRenderToString(t *testing.T) {
	// Create temporary template file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.md")

	tests := []struct {
		name         string
		template     string
		data         interface{}
		expected     string
		expectError  bool
	}{
		{
			name:     "simple template",
			template: "Hello {{ .Name }}!",
			data: map[string]interface{}{
				"Name": "World",
			},
			expected: "Hello World!",
		},
		{
			name:     "template with join function",
			template: "Items: {{ join .Items \", \" }}",
			data: map[string]interface{}{
				"Items": []string{"apple", "banana", "cherry"},
			},
			expected: "Items: apple, banana, cherry",
		},
		{
			name:     "template with comparison",
			template: "{{ if gt .Count 5 }}Many{{ else }}Few{{ end }}",
			data: map[string]interface{}{
				"Count": 10,
			},
			expected: "Many",
		},
		{
			name:     "template with math",
			template: "Total: {{ add .A .B }}",
			data: map[string]interface{}{
				"A": 5,
				"B": 3,
			},
			expected: "Total: 8",
		},
		{
			name:     "template with percent",
			template: "{{ percent .Value .Total }}%",
			data: map[string]interface{}{
				"Value": 25,
				"Total": 100,
			},
			expected: "25%",
		},
		{
			name:     "template with percentf",
			template: "{{ percentf .Value .Total 2 }}%",
			data: map[string]interface{}{
				"Value": 1,
				"Total": 3,
			},
			expected: "33.33%",
		},
		{
			name:     "template with missing key (zero mode)",
			template: "Value: {{ .Missing }}",
			data:     map[string]interface{}{},
			expected: "Value: <no value>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write template
			err := os.WriteFile(templatePath, []byte(tt.template), 0644)
			require.NoError(t, err)

			// Render
			renderer := NewRenderer(templatePath)
			result, err := renderer.RenderToString(tt.data)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRenderToFile(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.md")
	outputPath := filepath.Join(tmpDir, "output", "report.md")

	// Create template
	template := "# Report\n\nName: {{ .Name }}\nCount: {{ .Count }}"
	err := os.WriteFile(templatePath, []byte(template), 0644)
	require.NoError(t, err)

	// Render to file
	data := map[string]interface{}{
		"Name":  "Test Report",
		"Count": 42,
	}

	renderer := NewRenderer(templatePath)
	err = renderer.RenderToFile(outputPath, data)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, outputPath)

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "# Report\n\nName: Test Report\nCount: 42", string(content))
}

func TestRenderToFileCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.md")
	outputPath := filepath.Join(tmpDir, "nested", "deep", "output.md")

	// Create template
	err := os.WriteFile(templatePath, []byte("Test"), 0644)
	require.NoError(t, err)

	// Render - should create nested directories
	renderer := NewRenderer(templatePath)
	err = renderer.RenderToFile(outputPath, map[string]interface{}{})
	require.NoError(t, err)

	assert.FileExists(t, outputPath)
}

func TestRenderTemplateNotFound(t *testing.T) {
	renderer := NewRenderer("/nonexistent/template.md")
	_, err := renderer.RenderToString(map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestRenderInvalidTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "invalid.md")

	// Create invalid template (unclosed tag)
	err := os.WriteFile(templatePath, []byte("{{ .Name"), 0644)
	require.NoError(t, err)

	renderer := NewRenderer(templatePath)
	_, err = renderer.RenderToString(map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}

func TestDefaultFuncMap(t *testing.T) {
	funcs := defaultFuncMap()

	// Test string functions
	assert.NotNil(t, funcs["join"])
	assert.NotNil(t, funcs["split"])
	assert.NotNil(t, funcs["upper"])
	assert.NotNil(t, funcs["lower"])

	// Test comparison functions
	assert.NotNil(t, funcs["gt"])
	assert.NotNil(t, funcs["gte"])
	assert.NotNil(t, funcs["lt"])
	assert.NotNil(t, funcs["lte"])
	assert.NotNil(t, funcs["eq"])

	// Test math functions
	assert.NotNil(t, funcs["add"])
	assert.NotNil(t, funcs["sub"])
	assert.NotNil(t, funcs["mul"])
	assert.NotNil(t, funcs["div"])

	// Test percent functions
	assert.NotNil(t, funcs["percent"])
	assert.NotNil(t, funcs["percentf"])

	// Test utility functions
	assert.NotNil(t, funcs["riskBadge"])
	assert.NotNil(t, funcs["truncate"])
	assert.NotNil(t, funcs["first"])
	assert.NotNil(t, funcs["last"])

	// Test actual function behavior
	t.Run("string functions", func(t *testing.T) {
		joinFn := funcs["join"].(func([]string, string) string)
		assert.Equal(t, "a,b,c", joinFn([]string{"a", "b", "c"}, ","))

		upperFn := funcs["upper"].(func(string) string)
		assert.Equal(t, "HELLO", upperFn("hello"))
	})

	t.Run("comparison functions", func(t *testing.T) {
		gtFn := funcs["gt"].(func(int, int) bool)
		assert.True(t, gtFn(5, 3))
		assert.False(t, gtFn(3, 5))

		eqFn := funcs["eq"].(func(int, int) bool)
		assert.True(t, eqFn(5, 5))
		assert.False(t, eqFn(5, 3))
	})

	t.Run("math functions", func(t *testing.T) {
		addFn := funcs["add"].(func(int, int) int)
		assert.Equal(t, 8, addFn(5, 3))

		subFn := funcs["sub"].(func(int, int) int)
		assert.Equal(t, 2, subFn(5, 3))

		mulFn := funcs["mul"].(func(int, int) int)
		assert.Equal(t, 15, mulFn(5, 3))

		divFn := funcs["div"].(func(int, int) int)
		assert.Equal(t, 5, divFn(15, 3))
		assert.Equal(t, 0, divFn(15, 0)) // Division by zero returns 0
	})

	t.Run("percent functions", func(t *testing.T) {
		percentFn := funcs["percent"].(func(int, int) float64)
		assert.Equal(t, 25.0, percentFn(25, 100))
		assert.Equal(t, 0.0, percentFn(25, 0)) // Division by zero returns 0

		percentfFn := funcs["percentf"].(func(int, int, int) string)
		assert.Equal(t, "25.0", percentfFn(25, 100, 1))
		assert.Equal(t, "25.00", percentfFn(25, 100, 2))
		assert.Equal(t, "0.0", percentfFn(25, 0, 1)) // Division by zero
	})

	t.Run("utility functions", func(t *testing.T) {
		riskBadgeFn := funcs["riskBadge"].(func(string) string)
		assert.Equal(t, "🔴 Critical", riskBadgeFn("critical"))
		assert.Equal(t, "🔴 Critical", riskBadgeFn("Critical"))
		assert.Equal(t, "🟠 High", riskBadgeFn("high"))
		assert.Equal(t, "🟡 Medium", riskBadgeFn("medium"))
		assert.Equal(t, "🟢 Low", riskBadgeFn("low"))
		assert.Equal(t, "Unknown", riskBadgeFn("Unknown"))

		truncateFn := funcs["truncate"].(func(string, int) string)
		assert.Equal(t, "hello", truncateFn("hello", 10))
		assert.Equal(t, "hello", truncateFn("hello", 5))
		assert.Equal(t, "hel...", truncateFn("hello world", 6))
		assert.Equal(t, "hello wo...", truncateFn("hello world test", 11))
		assert.Equal(t, "h", truncateFn("hello", 1))

		firstFn := funcs["first"].(func(int, []string) []string)
		items := []string{"a", "b", "c", "d", "e"}
		assert.Equal(t, []string{"a", "b", "c"}, firstFn(3, items))
		assert.Equal(t, items, firstFn(10, items))
		assert.Equal(t, []string{}, firstFn(0, items))

		lastFn := funcs["last"].(func(int, []string) []string)
		assert.Equal(t, []string{"c", "d", "e"}, lastFn(3, items))
		assert.Equal(t, items, lastFn(10, items))
		assert.Equal(t, []string{}, lastFn(0, items))
	})
}

func TestNormalizeSpecPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already normalized",
			input:    "specs/module/test.feature",
			expected: "specs/module/test.feature",
		},
		{
			name:     "with ../ prefix",
			input:    "../specs/module/test.feature",
			expected: "specs/module/test.feature",
		},
		{
			name:     "with multiple ../ prefixes",
			input:    "../../specs/module/test.feature",
			expected: "specs/module/test.feature",
		},
		{
			name:     "without specs prefix",
			input:    "module/test.feature",
			expected: "specs/module/test.feature",
		},
		{
			name:     "with backslashes",
			input:    "..\\specs\\module\\test.feature",
			expected: "specs/module/test.feature",
		},
		{
			name:     "complex path",
			input:    "../../../specs/eac-commands/create/test.feature",
			expected: "specs/eac-commands/create/test.feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeSpecPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
