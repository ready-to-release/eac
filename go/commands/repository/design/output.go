// Package design provides architecture documentation commands using Structurizr DSL.
//
// This file contains output utilities that handle both user-facing console output
// and structured logging for the design command and its subcommands.
package design

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

// Output handles both user-facing output and logging.
// It provides a unified interface for console output and file logging.
type Output struct {
}

// NewOutput creates a new Output instance.
func NewOutput() *Output {
	return &Output{}
}

// Info logs and prints an info message.
// The message is both logged and printed to stdout.
func (o *Output) Info(msg string) {
	log.Info(msg)
	// Keep fmt.Println for this wrapper's output methods
	fmt.Println(msg)
}

// Infof logs and prints a formatted info message.
func (o *Output) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	o.Info(msg)
}

// Error logs and prints an error message.
// The message is both logged and printed to stderr.
func (o *Output) Error(msg string) {
	log.Errorf("%s", msg)
	// Keep fmt.Fprintln for this wrapper's output methods
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
	log.Debugf("%s", msg)
}

// Debugf logs a formatted debug message (no console output).
func (o *Output) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	o.Debug(msg)
}

// Warn logs and prints a warning message.
func (o *Output) Warn(msg string) {
	log.Warnf("%s", msg)
	// Keep fmt.Fprintf for this wrapper's output methods
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
	// For progress messages, we print to console but don't log to file
	// This is intentional - progress indicators are ephemeral
	fmt.Println(msg)
}

// Progressf prints a formatted progress message (console only).
func (o *Output) Progressf(format string, args ...interface{}) {
	// For progress messages, we print to console but don't log to file
	// This is intentional - progress indicators are ephemeral
	fmt.Printf(format+"\n", args...)
}
