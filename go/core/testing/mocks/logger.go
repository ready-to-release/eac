package mocks

import "github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"

// MockLogger implements interfaces.LoggerPort for testing.
// It captures all log calls for verification in tests.
type MockLogger struct {
	DebugMessages []string
	InfoMessages  []string
	WarnMessages  []string
	ErrorMessages []string
	name          string
}

// NewMockLogger creates a new MockLogger.
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// WithName sets the logger name.
func (m *MockLogger) WithName(name string) *MockLogger {
	m.name = name
	return m
}

// Debug implements interfaces.LoggerPort.
func (m *MockLogger) Debug(msg string) {
	m.DebugMessages = append(m.DebugMessages, msg)
}

// Info implements interfaces.LoggerPort.
func (m *MockLogger) Info(msg string) {
	m.InfoMessages = append(m.InfoMessages, msg)
}

// Warn implements interfaces.LoggerPort.
func (m *MockLogger) Warn(msg string) {
	m.WarnMessages = append(m.WarnMessages, msg)
}

// Error implements interfaces.LoggerPort.
func (m *MockLogger) Error(msg string) {
	m.ErrorMessages = append(m.ErrorMessages, msg)
}

// Debugf implements interfaces.LoggerPort.
func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Debug(format) // Simplified for testing
}

// Infof implements interfaces.LoggerPort.
func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Info(format) // Simplified for testing
}

// Warnf implements interfaces.LoggerPort.
func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Warn(format) // Simplified for testing
}

// Errorf implements interfaces.LoggerPort.
func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Error(format) // Simplified for testing
}

// WithSuffix implements interfaces.LoggerPort.
func (m *MockLogger) WithSuffix(suffix string) interfaces.LoggerPort {
	child := NewMockLogger()
	if m.name != "" {
		child.name = m.name + "." + suffix
	} else {
		child.name = suffix
	}
	return child
}

// Sync implements interfaces.LoggerPort.
func (m *MockLogger) Sync() error {
	return nil
}

// Clear resets all captured messages.
func (m *MockLogger) Clear() {
	m.DebugMessages = nil
	m.InfoMessages = nil
	m.WarnMessages = nil
	m.ErrorMessages = nil
}

// Interface compliance check
var _ interfaces.LoggerPort = (*MockLogger)(nil)
