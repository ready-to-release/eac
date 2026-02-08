package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	containerport "github.com/ready-to-release/eac/contracts/container-runtime/0.1.0"
)

// TestIsRetryableError_WindowsFileLock tests detection of Windows file lock errors.
func TestIsRetryableError_WindowsFileLock(t *testing.T) {
	tests := []struct {
		name     string
		stderr   string
		expected bool
	}{
		{
			name:     "Windows file access error",
			stderr:   "The process cannot access the file because it is being used by another process",
			expected: true,
		},
		{
			name:     "Windows delete error",
			stderr:   "Cannot delete file: access denied",
			expected: true,
		},
		{
			name:     "Node.js EBUSY error",
			stderr:   "Error: EBUSY: resource busy or locked",
			expected: true,
		},
		{
			name:     "Unix text file busy",
			stderr:   "text file busy",
			expected: true,
		},
		{
			name:     "resource temporarily unavailable",
			stderr:   "resource temporarily unavailable",
			expected: true,
		},
		{
			name:     "normal error output",
			stderr:   "Command failed with exit code 1",
			expected: false,
		},
		{
			name:     "empty stderr",
			stderr:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &containerport.ContainerResult{
				ExitCode: 1,
				Stderr:   []byte(tt.stderr),
			}
			got := isRetryableError(nil, result)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestIsRetryableError_DockerErrors tests detection of transient Docker errors.
func TestIsRetryableError_DockerErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "container already exists",
			err:      &testError{msg: "container already exists: conflict"},
			expected: true,
		},
		{
			name:     "network not found",
			err:      &testError{msg: "network bridge not found"},
			expected: true,
		},
		{
			name:     "unable to remove",
			err:      &testError{msg: "unable to remove container"},
			expected: true,
		},
		{
			name:     "generic error",
			err:      &testError{msg: "some other error"},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableError(tt.err, nil)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestIsRetryableError_NilResult tests that nil result with nil error is not retryable.
func TestIsRetryableError_NilResult(t *testing.T) {
	got := isRetryableError(nil, nil)
	assert.False(t, got)
}

// TestIsRetryableError_SuccessResult tests that success is not retryable.
func TestIsRetryableError_SuccessResult(t *testing.T) {
	result := &containerport.ContainerResult{
		ExitCode: 0,
		Stderr:   []byte("some output"),
	}
	got := isRetryableError(nil, result)
	assert.False(t, got)
}

// testError is a simple error implementation for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
