// Package config provides timeout helper functions for type-safe access.
package config

import (
	"context"
	"sync"
	"time"

	tui "github.com/ready-to-release/eac/contracts/tui-adapter/0.1.0/interfaces"
)

// Global timeout configuration (loaded at startup).
var (
	globalTimeouts   *TimeoutConfig
	globalTimeoutsMu sync.RWMutex
)

// SetGlobalTimeouts sets the global timeout configuration.
// Called during config loading.
func SetGlobalTimeouts(cfg *TimeoutConfig) {
	globalTimeoutsMu.Lock()
	defer globalTimeoutsMu.Unlock()
	globalTimeouts = cfg
}

// Timeouts returns the global timeout configuration.
// Returns defaults if timeouts have not been explicitly loaded.
func Timeouts() *TimeoutConfig {
	globalTimeoutsMu.RLock()
	defer globalTimeoutsMu.RUnlock()
	if globalTimeouts == nil {
		return DefaultTimeoutConfig()
	}
	return globalTimeouts
}

// ============================================================================
// Docker Timeout Helpers
// ============================================================================

// DockerQueryTimeout returns the timeout for quick Docker queries.
// Use for: docker info, memory detection, daemon ping.
func DockerQueryTimeout() time.Duration {
	return Timeouts().Docker.Query.D()
}

// DockerImagePullTimeout returns the timeout for image pulls.
func DockerImagePullTimeout() time.Duration {
	return Timeouts().Docker.ImagePull.D()
}

// DockerImageBuildTimeout returns the timeout for image builds.
func DockerImageBuildTimeout() time.Duration {
	return Timeouts().Docker.ImageBuild.D()
}

// DockerContainerStartTimeout returns the timeout for container startup.
func DockerContainerStartTimeout() time.Duration {
	return Timeouts().Docker.ContainerStart.D()
}

// DockerContainerStopTimeout returns the timeout for graceful container stop.
func DockerContainerStopTimeout() time.Duration {
	return Timeouts().Docker.ContainerStop.D()
}

// DockerContainerExecTimeout returns the timeout for container execution.
func DockerContainerExecTimeout() time.Duration {
	return Timeouts().Docker.ContainerExec.D()
}

// DockerValidationTimeout returns the timeout for Docker validation.
func DockerValidationTimeout() time.Duration {
	return Timeouts().Docker.Validation.D()
}

// WithDockerQueryContext returns a context with Docker query timeout.
func WithDockerQueryContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, DockerQueryTimeout())
}

// WithDockerExecContext returns a context with container execution timeout.
func WithDockerExecContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, DockerContainerExecTimeout())
}

// ============================================================================
// HTTP Timeout Helpers
// ============================================================================

// HTTPRequestTimeout returns the timeout for standard HTTP requests.
func HTTPRequestTimeout() time.Duration {
	return Timeouts().HTTP.Request.D()
}

// HTTPDownloadTimeout returns the timeout for file downloads.
func HTTPDownloadTimeout() time.Duration {
	return Timeouts().HTTP.Download.D()
}

// HTTPDialTimeout returns the timeout for connection establishment.
func HTTPDialTimeout() time.Duration {
	return Timeouts().HTTP.Dial.D()
}

// HTTPIdleTimeout returns the timeout for idle connections.
func HTTPIdleTimeout() time.Duration {
	return Timeouts().HTTP.Idle.D()
}

// ============================================================================
// File System Timeout Helpers
// ============================================================================

// FileLockTimeout returns the timeout for file lock acquisition.
func FileLockTimeout() time.Duration {
	return Timeouts().Filesystem.LockAcquire.D()
}

// StaleLockThreshold returns the threshold for stale lock cleanup.
func StaleLockThreshold() time.Duration {
	return Timeouts().Filesystem.StaleLock.D()
}

// FileWatchDebounce returns the file watcher debounce interval.
func FileWatchDebounce() time.Duration {
	return Timeouts().Filesystem.WatchDebounce.D()
}

// WithFileLockContext returns a context with file lock timeout.
func WithFileLockContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, FileLockTimeout())
}

// ============================================================================
// CI/Pipeline Timeout Helpers
// ============================================================================

// CIDispatchSettleTime returns the time to wait after workflow dispatch.
func CIDispatchSettleTime() time.Duration {
	return Timeouts().CI.DispatchSettle.D()
}

// CIPollInterval returns the status polling interval.
func CIPollInterval() time.Duration {
	return Timeouts().CI.PollInterval.D()
}

// CIWorkflowCompletionTimeout returns the default workflow completion timeout.
func CIWorkflowCompletionTimeout() time.Duration {
	return Timeouts().CI.WorkflowCompletion.D()
}

// CIRecentRunWindow returns the window for considering runs as recent.
func CIRecentRunWindow() time.Duration {
	return Timeouts().CI.RecentRunWindow.D()
}

// ============================================================================
// Long Operation Timeout Helpers
// ============================================================================

// SecurityScanTimeout returns the timeout for security scans.
func SecurityScanTimeout() time.Duration {
	return Timeouts().LongOperations.SecurityScan.D()
}

// BuildTimeout returns the timeout for build operations.
func BuildTimeout() time.Duration {
	return Timeouts().LongOperations.Build.D()
}

// TestTimeout returns the timeout for test operations.
func TestTimeout() time.Duration {
	return Timeouts().LongOperations.Test.D()
}

// EvidenceMaxAge returns the maximum age for valid evidence.
func EvidenceMaxAge() time.Duration {
	return Timeouts().LongOperations.EvidenceMaxAge.D()
}

// WorkerTimeout returns the timeout for killing hanging workers.
func WorkerTimeout() time.Duration {
	return Timeouts().LongOperations.WorkerTimeout.D()
}

// ============================================================================
// TUI Timeout Helpers
// ============================================================================

// TUIAutoScrollResumeTimeout returns the auto-scroll resume delay.
func TUIAutoScrollResumeTimeout() time.Duration {
	return Timeouts().TUI.AutoScrollResume.D()
}

// TUIExitCountdownDuration returns the exit countdown duration.
func TUIExitCountdownDuration() time.Duration {
	return Timeouts().TUI.ExitCountdown.D()
}

// PortReservationTTL returns the port reservation TTL.
func PortReservationTTL() time.Duration {
	return Timeouts().TUI.PortReservation.D()
}

// TUIConfig returns the TUI configuration from global timeouts.
// This creates a TUIConfig struct suitable for passing to console.NewModel.
func TUIConfig() *tui.TUIConfig {
	t := Timeouts()
	return &tui.TUIConfig{
		// Timeouts
		MetricsInterval:  t.TUI.MetricsInterval.D(),
		MinDisplayTime:   t.TUI.MinDisplayTime.D(),
		ExitCountdown:    t.TUI.ExitCountdown.D(),
		FreezeCountdown:  t.TUI.FreezeCountdown.D(),
		AutoScrollResume: t.TUI.AutoScrollResume.D(),

		// Layout
		MaxTabs:           t.TUILayout.MaxTabs,
		DefaultColumns:    t.TUILayout.DefaultColumns,
		MinColumns:        t.TUILayout.MinColumns,
		MaxColumns:        t.TUILayout.MaxColumns,
		BufferSizePane:    t.TUILayout.BufferSizePane,
		BufferSizeResults: t.TUILayout.BufferSizeResults,
		BufferSizeUoW:     t.TUILayout.BufferSizeUoW,
	}
}

// ============================================================================
// Scheduling Timeout Helpers
// ============================================================================

// CapacityRecalcInterval returns the capacity recalculation interval.
func CapacityRecalcInterval() time.Duration {
	return Timeouts().Scheduling.CapacityRecalc.D()
}
