// Package internal provides shared utilities for work commands
package internal

import "time"

// timeOperation wraps a git operation with debug logging and timing.
// It logs the operation name and parameters before execution, runs fn,
// then logs the operation name, duration, and any error after execution.
// This eliminates the repetitive start/duration boilerplate in git_ops.go.
func timeOperation(name string, params string, fn func() error) error {
	log.Debugf("Git operation: %s | %s", name, params)

	start := time.Now()
	err := fn()

	log.Debugf("Git operation completed: %s | duration=%v err=%v", name, time.Since(start), err)

	return err
}
