//go:build L0
// +build L0

package logger

import (
	"bytes"
	"context"
	"os"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "default config",
			config: Config{
				Console: true,
			},
			wantErr: false,
		},
		{
			name: "with file output",
			config: Config{
				File:    "/tmp/test.log",
				Console: true,
			},
			wantErr: false,
		},
		{
			name: "json output",
			config: Config{
				Console: true,
				JSON:    true,
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: Config{
				Level:   "invalid",
				Console: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, logger)
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
		wantErr  bool
	}{
		{"trace", zerolog.TraceLevel, false},
		{"debug", zerolog.DebugLevel, false},
		{"info", zerolog.InfoLevel, false},
		{"warn", zerolog.WarnLevel, false},
		{"warning", zerolog.WarnLevel, false},
		{"error", zerolog.ErrorLevel, false},
		{"fatal", zerolog.FatalLevel, false},
		{"panic", zerolog.PanicLevel, false},
		{"WARN", zerolog.WarnLevel, false},
		{"invalid", zerolog.InfoLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := parseLevel(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, level)
			}
		})
	}
}

func TestDetectEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "default development",
			envVars:  map[string]string{},
			expected: "development",
		},
		{
			name:     "CI environment",
			envVars:  map[string]string{"CI": "true"},
			expected: "ci",
		},
		{
			name:     "GitHub Actions",
			envVars:  map[string]string{"GITHUB_ACTIONS": "true"},
			expected: "ci",
		},
		{
			name:     "production via R2R_ENV",
			envVars:  map[string]string{"R2R_ENV": "production"},
			expected: "production",
		},
		{
			name:     "production via ENV",
			envVars:  map[string]string{"ENV": "production"},
			expected: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment values and create cleanup function
			originalValues := make(map[string]string)
			keysToUnset := make([]string, 0)
			
			// Set test environment variables
			for k, v := range tt.envVars {
				if original, exists := os.LookupEnv(k); exists {
					originalValues[k] = original
				} else {
					keysToUnset = append(keysToUnset, k)
				}
				os.Setenv(k, v)
			}
			
			// Also need to clear CI-related vars for clean test environment
			ciVars := []string{"CI", "GITHUB_ACTIONS", "R2R_ENV", "ENV"}
			ciOriginals := make(map[string]string)
			ciToUnset := make([]string, 0)
			
			for _, k := range ciVars {
				if _, exists := tt.envVars[k]; !exists { // Don't interfere with vars set by test
					if original, exists := os.LookupEnv(k); exists {
						ciOriginals[k] = original
						os.Unsetenv(k) // Clear for clean test
					} else {
						ciToUnset = append(ciToUnset, k)
					}
				}
			}

			// Test function
			result := detectEnvironment()
			assert.Equal(t, tt.expected, result)

			// Cleanup: restore original environment
			for k, v := range originalValues {
				os.Setenv(k, v)
			}
			for _, k := range keysToUnset {
				os.Unsetenv(k)
			}
			
			// Restore CI environment
			for k, v := range ciOriginals {
				os.Setenv(k, v)
			}
			for _, k := range ciToUnset {
				os.Unsetenv(k)
			}
		})
	}
}

func TestWithContext(t *testing.T) {
	logger, err := New(Config{Console: false})
	require.NoError(t, err)

	// Create context with values
	ctx := context.Background()
	ctx = ContextWithOperationID(ctx, "test-op-123")
	ctx = ContextWithCommand(ctx, "test-command")
	ctx = ContextWithUser(ctx, "test-user")
	ctx = ContextWithComponent(ctx, "test-component")

	// Create logger with context
	var buf bytes.Buffer
	contextLogger := logger.WithContext(ctx)
	contextLogger.Logger = contextLogger.Logger.Output(&buf)

	// Log a message
	contextLogger.Info().Msg("test message")

	// Check output contains context fields
	output := buf.String()
	assert.Contains(t, output, "test-op-123")
	assert.Contains(t, output, "test-command")
	assert.Contains(t, output, "test-user")
	assert.Contains(t, output, "test-component")
}

func TestWithFields(t *testing.T) {
	logger, err := New(Config{Console: false})
	require.NoError(t, err)

	var buf bytes.Buffer
	logger.Logger = logger.Logger.Output(&buf)

	// Test WithField
	fieldLogger := logger.WithField("key1", "value1")
	fieldLogger.Info().Msg("single field test")
	assert.Contains(t, buf.String(), "key1")
	assert.Contains(t, buf.String(), "value1")

	// Test WithFields
	buf.Reset()
	fieldsLogger := logger.WithFields(map[string]interface{}{
		"key2": "value2",
		"key3": 123,
		"key4": true,
	})
	fieldsLogger.Info().Msg("multiple fields test")
	output := buf.String()
	assert.Contains(t, output, "key2")
	assert.Contains(t, output, "value2")
	assert.Contains(t, output, "key3")
	assert.Contains(t, output, "123")
	assert.Contains(t, output, "key4")
	assert.Contains(t, output, "true")
}

func TestSetLevel(t *testing.T) {
	logger, err := New(Config{Console: false, Level: "info"})
	require.NoError(t, err)

	var buf bytes.Buffer
	logger.Logger = logger.Logger.Output(&buf)

	// Debug should not be logged at info level
	logger.Debug().Msg("debug message")
	assert.Empty(t, buf.String())

	// Change level to debug
	err = logger.SetLevel("debug")
	require.NoError(t, err)

	// Now debug should be logged
	logger.Debug().Msg("debug message after level change")
	assert.Contains(t, buf.String(), "debug message after level change")
}

func TestLevelWriter(t *testing.T) {
	var buf bytes.Buffer
	lw := &LevelWriter{Writer: &buf, level: zerolog.InfoLevel}

	// Write info level - should pass through
	n, err := lw.WriteLevel(zerolog.InfoLevel, []byte("info message"))
	assert.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, "info message", buf.String())

	// Write debug level - should be filtered
	buf.Reset()
	n, err = lw.WriteLevel(zerolog.DebugLevel, []byte("debug message"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n) // Returns length even when filtered
	assert.Empty(t, buf.String())
}

func TestPackageLevelFunctions(t *testing.T) {
	// Initialize with a buffer we can check
	var buf bytes.Buffer
	once = sync.Once{} // Reset the once
	defaultLogger = nil

	// Create a custom logger that writes to our buffer
	logger, err := New(Config{Console: false, Level: "debug"})
	require.NoError(t, err)
	logger.Logger = logger.Logger.Output(&buf)
	defaultLogger = logger

	// Test package-level functions
	Debug("debug message")
	assert.Contains(t, buf.String(), "debug message")

	buf.Reset()
	Info("info message")
	assert.Contains(t, buf.String(), "info message")

	buf.Reset()
	Warn("warn message")
	assert.Contains(t, buf.String(), "warn message")

	buf.Reset()
	Error("error message")
	assert.Contains(t, buf.String(), "error message")

	// Test WithField
	buf.Reset()
	WithField("test", "value").Info().Msg("with field")
	output := buf.String()
	assert.Contains(t, output, "test")
	assert.Contains(t, output, "value")

	// Test WithFields
	buf.Reset()
	WithFields(map[string]interface{}{
		"field1": "value1",
		"field2": 42,
	}).Info().Msg("with fields")
	output = buf.String()
	assert.Contains(t, output, "field1")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "field2")
	assert.Contains(t, output, "42")
}

func TestCreateFileWriter(t *testing.T) {
	// Test creating file in temp directory
	tmpFile := "/tmp/test-logger.log"
	defer os.Remove(tmpFile)

	writer, err := createFileWriter(tmpFile)
	require.NoError(t, err)
	assert.NotNil(t, writer)

	// Test writing to file
	file, ok := writer.(*os.File)
	require.True(t, ok)
	_, err = file.WriteString("test log entry\n")
	assert.NoError(t, err)
	file.Close()

	// Verify file contents
	contents, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "test log entry\n", string(contents))
}

func TestConsoleWriterModes(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectColored bool
		expectJSON    bool
	}{
		{
			name:          "development mode",
			config:        Config{Environment: "development"},
			expectColored: true,
			expectJSON:    false,
		},
		{
			name:          "CI mode",
			config:        Config{Environment: "ci"},
			expectColored: false,
			expectJSON:    true,
		},
		{
			name:          "production mode",
			config:        Config{Environment: "production"},
			expectColored: false,
			expectJSON:    true,
		},
		{
			name:          "JSON requested",
			config:        Config{JSON: true},
			expectColored: false,
			expectJSON:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := createConsoleWriter(tt.config, zerolog.InfoLevel)
			assert.NotNil(t, writer)

			// Check if it's wrapped in LevelWriter
			lw, ok := writer.(*LevelWriter)
			require.True(t, ok)

			// Check the underlying writer type
			switch w := lw.Writer.(type) {
			case zerolog.ConsoleWriter:
				assert.True(t, tt.expectColored, "expected colored console writer")
			case *os.File:
				assert.True(t, tt.expectJSON, "expected JSON output to stdout")
			default:
				t.Fatalf("unexpected writer type: %T", w)
			}
		})
	}
}
