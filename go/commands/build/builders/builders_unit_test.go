package builders

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogln_FormatsOutput(t *testing.T) {
	var buf bytes.Buffer
	Logln(&buf, "hello %s %d", "world", 42)
	output := buf.String()
	assert.Contains(t, output, "hello world 42")
}

func TestSubstituteVars_AllCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		vars     map[string]string
		expected string
	}{
		{"single var", "image:${TAG}", map[string]string{"${TAG}": "v1.0"}, "image:v1.0"},
		{"multiple vars", "${NAME}:${TAG}", map[string]string{"${NAME}": "myimage", "${TAG}": "latest"}, "myimage:latest"},
		{"no match", "plain-text", map[string]string{"${FOO}": "bar"}, "plain-text"},
		{"empty vars", "unchanged", map[string]string{}, "unchanged"},
		{"nil vars", "unchanged", nil, "unchanged"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteVars(tt.input, tt.vars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

