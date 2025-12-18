package internal

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/serve"
	"github.com/stretchr/testify/assert"
)

func TestOneOffDockerRunner_Creation(t *testing.T) {
	// Test with mock client
	mockClient := new(serve.MockDockerClient)
	mockClient.On("Close").Return(nil)

	runner := NewOneOffDockerRunnerWithClient(mockClient)
	assert.NotNil(t, runner)

	err := runner.Close()
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestStripDockerLogHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "JSON object",
			input:    []byte("\x00\x00\x00\x00\x00\x00\x00\x14{\"key\":\"value\"}"),
			expected: []byte("{\"key\":\"value\"}"),
		},
		{
			name:     "JSON array",
			input:    []byte("\x00\x00\x00\x00\x00\x00\x00\x10[\"item1\",\"item2\"]"),
			expected: []byte("[\"item1\",\"item2\"]"),
		},
		{
			name:     "Already clean JSON",
			input:    []byte("{\"key\":\"value\"}"),
			expected: []byte("{\"key\":\"value\"}"),
		},
		{
			name:     "JSON with trailing text",
			input:    []byte("{\"key\":\"value\"}\nsome trailing text"),
			expected: []byte("{\"key\":\"value\"}"),
		},
		{
			name:     "JSON with prefix and suffix",
			input:    []byte("prefix text {\"key\":\"value\"} suffix text"),
			expected: []byte("{\"key\":\"value\"}"),
		},
		{
			name:     "Empty input",
			input:    []byte(""),
			expected: []byte(""),
		},
		{
			name:     "No JSON found",
			input:    []byte("some error message"),
			expected: []byte("some error message"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripDockerLogHeaders(tt.input)
			assert.Equal(t, string(tt.expected), string(result))
		})
	}
}

func TestOneOffDockerRunner_WithMockClient(t *testing.T) {
	mockClient := new(serve.MockDockerClient)

	// Create runner with mock
	runner := NewOneOffDockerRunnerWithClient(mockClient)
	assert.NotNil(t, runner)

	// Test Close
	mockClient.On("Close").Return(nil).Once()
	err := runner.Close()
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}
