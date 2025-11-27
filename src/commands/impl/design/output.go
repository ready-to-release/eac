// Package design provides architecture documentation commands using Structurizr DSL.
//
// This file contains output utilities that handle both user-facing console output
// and structured logging for the design command and its subcommands.
package design

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/core/logging"
)

// Output handles both user-facing output and logging.
// It provides a unified interface for console output and file logging.
type Output struct {
	logger *logging.Logger
}

// NewOutput creates a new Output instance with the given logger.
// The logger can be nil, in which case only console output is performed.
func NewOutput(logger *logging.Logger) *Output {
	return &Output{logger: logger}
}

// Info logs and prints an info message.
// The message is both logged (if logger is configured) and printed to stdout.
func (o *Output) Info(msg string) {
	if o.logger != nil {
		o.logger.Info(msg)
	}
	fmt.Println(msg)
}

// Infof logs and prints a formatted info message.
func (o *Output) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	o.Info(msg)
}

// Error logs and prints an error message.
// The message is both logged (if logger is configured) and printed to stderr.
func (o *Output) Error(msg string) {
	if o.logger != nil {
		o.logger.Error(msg)
	}
	fmt.Fprintln(os.Stderr, msg)
}

// Errorf logs and prints a formatted error message.
func (o *Output) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	o.Error(msg)
}

// Debug logs a debug message (no console output).
// Debug messages are only written to the log file when debug mode is enabled.
func (o *Output) Debug(msg string) {
	if o.logger != nil {
		o.logger.Debug(msg)
	}
}

// Debugf logs a formatted debug message (no console output).
func (o *Output) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	o.Debug(msg)
}

// Warn logs and prints a warning message.
func (o *Output) Warn(msg string) {
	if o.logger != nil {
		o.logger.Warn(msg)
	}
	fmt.Fprintf(os.Stderr, "Warning: %s\n", msg)
}

// Warnf logs and prints a formatted warning message.
func (o *Output) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	o.Warn(msg)
}

// Progress prints a progress message (console only, no logging).
// Use this for user-facing progress indicators like spinners or status updates.
func (o *Output) Progress(msg string) {
	fmt.Println(msg)
}

// Progressf prints a formatted progress message (console only).
func (o *Output) Progressf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
