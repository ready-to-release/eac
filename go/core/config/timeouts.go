// Package config provides timeout configuration for the EAC CLI.
package config

import (
	"time"
)

// Duration is a wrapper around time.Duration that supports YAML unmarshaling
// from Go-style duration strings (e.g., "30s", "5m", "2h").
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler for Duration.
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// MarshalYAML implements yaml.Marshaler for Duration.
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// D returns the underlying time.Duration value.
func (d Duration) D() time.Duration {
	return time.Duration(d)
}

// TimeoutConfig holds all configurable timeouts for the CLI.
// Loaded from contracts/core/0.1.0/defaults/timeouts.yml
// with optional user overrides from .eac/timeouts.yml.
type TimeoutConfig struct {
	Docker         DockerTimeouts         `yaml:"docker"`
	HTTP           HTTPTimeouts           `yaml:"http"`
	Filesystem     FilesystemTimeouts     `yaml:"filesystem"`
	CI             CITimeouts             `yaml:"ci"`
	LongOperations LongOperationTimeouts  `yaml:"long_operations"`
	TUI            TUITimeouts            `yaml:"tui"`
	TUILayout      TUILayoutConfig        `yaml:"tui_layout"`
	Scheduling     SchedulingTimeouts     `yaml:"scheduling"`
}

// DockerTimeouts holds Docker-related timeouts.
type DockerTimeouts struct {
	Query          Duration `yaml:"query"`           // Quick daemon queries
	ImagePull      Duration `yaml:"image_pull"`      // Image pull operations
	ImageBuild     Duration `yaml:"image_build"`     // Image build operations
	ContainerStart Duration `yaml:"container_start"` // Container startup
	ContainerStop  Duration `yaml:"container_stop"`  // Graceful stop
	ContainerExec  Duration `yaml:"container_exec"`  // Container execution
	Validation     Duration `yaml:"validation"`      // Validation operations
}

// HTTPTimeouts holds HTTP/network timeouts.
type HTTPTimeouts struct {
	Request  Duration `yaml:"request"`  // Standard requests
	Download Duration `yaml:"download"` // File downloads
	Dial     Duration `yaml:"dial"`     // Connection establishment
	Idle     Duration `yaml:"idle"`     // Idle connections
}

// FilesystemTimeouts holds file system timeouts.
type FilesystemTimeouts struct {
	LockAcquire   Duration `yaml:"lock_acquire"`   // File lock acquisition
	StaleLock     Duration `yaml:"stale_lock"`     // Stale lock threshold
	WatchDebounce Duration `yaml:"watch_debounce"` // File watcher debounce
}

// CITimeouts holds CI/pipeline timeouts.
type CITimeouts struct {
	DispatchSettle     Duration `yaml:"dispatch_settle"`     // Post-dispatch settle time
	PollInterval       Duration `yaml:"poll_interval"`       // Status polling interval
	WorkflowCompletion Duration `yaml:"workflow_completion"` // Default completion timeout
	RecentRunWindow    Duration `yaml:"recent_run_window"`   // Recent run window
}

// LongOperationTimeouts holds long operation timeouts.
type LongOperationTimeouts struct {
	SecurityScan   Duration `yaml:"security_scan"`    // Security scan timeout
	Build          Duration `yaml:"build"`            // Build timeout
	Test           Duration `yaml:"test"`             // Test timeout
	EvidenceMaxAge Duration `yaml:"evidence_max_age"` // Evidence validity period
	WorkerTimeout  Duration `yaml:"worker_timeout"`   // Kill workers that hang
}

// TUITimeouts holds TUI-related timeouts.
type TUITimeouts struct {
	AutoScrollResume Duration `yaml:"auto_scroll_resume"` // Auto-scroll resume delay
	ExitCountdown    Duration `yaml:"exit_countdown"`     // Exit countdown duration
	PortReservation  Duration `yaml:"port_reservation"`   // Port reservation TTL
	FreezeCountdown  Duration `yaml:"freeze_countdown"`   // Extended countdown when Freeze clicked
	MinDisplayTime   Duration `yaml:"min_display_time"`   // Minimum time to show completion state
	MetricsInterval  Duration `yaml:"metrics_interval"`   // CPU/memory metrics update interval
}

// TUILayoutConfig holds TUI layout configuration.
type TUILayoutConfig struct {
	MaxTabs           int `yaml:"max_tabs"`            // Maximum visible tabs before scrolling
	DefaultColumns    int `yaml:"default_columns"`     // Default number of tab columns
	MinColumns        int `yaml:"min_columns"`         // Minimum tab columns
	MaxColumns        int `yaml:"max_columns"`         // Maximum tab columns
	BufferSizePane    int `yaml:"buffer_size_pane"`    // Buffer size for each pane
	BufferSizeResults int `yaml:"buffer_size_results"` // Buffer size for results
	BufferSizeUoW     int `yaml:"buffer_size_uow"`     // Buffer size per UoW
}

// SchedulingTimeouts holds scheduling-related timeouts.
type SchedulingTimeouts struct {
	CapacityRecalc Duration `yaml:"capacity_recalc"`  // Capacity recalculation interval
}

// DefaultTimeoutConfig returns the default timeout configuration.
// These values match the current hardcoded values in the codebase.
func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		Docker: DockerTimeouts{
			Query:          Duration(5 * time.Second),
			ImagePull:      Duration(5 * time.Minute),
			ImageBuild:     Duration(30 * time.Minute),
			ContainerStart: Duration(60 * time.Second),
			ContainerStop:  Duration(5 * time.Second),
			ContainerExec:  Duration(30 * time.Minute),
			Validation:     Duration(30 * time.Second),
		},
		HTTP: HTTPTimeouts{
			Request:  Duration(30 * time.Second),
			Download: Duration(5 * time.Minute),
			Dial:     Duration(10 * time.Second),
			Idle:     Duration(60 * time.Second),
		},
		Filesystem: FilesystemTimeouts{
			LockAcquire:   Duration(30 * time.Second),
			StaleLock:     Duration(5 * time.Minute),
			WatchDebounce: Duration(500 * time.Millisecond),
		},
		CI: CITimeouts{
			DispatchSettle:     Duration(2 * time.Second),
			PollInterval:       Duration(10 * time.Second),
			WorkflowCompletion: Duration(5 * time.Minute),
			RecentRunWindow:    Duration(2 * time.Hour),
		},
		LongOperations: LongOperationTimeouts{
			SecurityScan:   Duration(30 * time.Minute),
			Build:          Duration(30 * time.Minute),
			Test:           Duration(30 * time.Minute),
			EvidenceMaxAge: Duration(24 * time.Hour),
			WorkerTimeout:  Duration(3 * time.Minute),
		},
		TUI: TUITimeouts{
			AutoScrollResume: Duration(8 * time.Second),
			ExitCountdown:    Duration(10 * time.Second),
			PortReservation:  Duration(30 * time.Second),
			FreezeCountdown:  Duration(120 * time.Second),
			MinDisplayTime:   Duration(1500 * time.Millisecond),
			MetricsInterval:  Duration(500 * time.Millisecond),
		},
		TUILayout: TUILayoutConfig{
			MaxTabs:           36,
			DefaultColumns:    4,
			MinColumns:        2,
			MaxColumns:        6,
			BufferSizePane:    500,
			BufferSizeResults: 100,
			BufferSizeUoW:     200,
		},
		Scheduling: SchedulingTimeouts{
			CapacityRecalc: Duration(2 * time.Second),
		},
	}
}

// MergeTimeoutConfigs merges user overrides into defaults.
// Non-zero values in override replace the corresponding default values.
func MergeTimeoutConfigs(defaults, override *TimeoutConfig) *TimeoutConfig {
	if override == nil {
		return defaults
	}

	result := *defaults

	// Docker
	if override.Docker.Query != 0 {
		result.Docker.Query = override.Docker.Query
	}
	if override.Docker.ImagePull != 0 {
		result.Docker.ImagePull = override.Docker.ImagePull
	}
	if override.Docker.ImageBuild != 0 {
		result.Docker.ImageBuild = override.Docker.ImageBuild
	}
	if override.Docker.ContainerStart != 0 {
		result.Docker.ContainerStart = override.Docker.ContainerStart
	}
	if override.Docker.ContainerStop != 0 {
		result.Docker.ContainerStop = override.Docker.ContainerStop
	}
	if override.Docker.ContainerExec != 0 {
		result.Docker.ContainerExec = override.Docker.ContainerExec
	}
	if override.Docker.Validation != 0 {
		result.Docker.Validation = override.Docker.Validation
	}

	// HTTP
	if override.HTTP.Request != 0 {
		result.HTTP.Request = override.HTTP.Request
	}
	if override.HTTP.Download != 0 {
		result.HTTP.Download = override.HTTP.Download
	}
	if override.HTTP.Dial != 0 {
		result.HTTP.Dial = override.HTTP.Dial
	}
	if override.HTTP.Idle != 0 {
		result.HTTP.Idle = override.HTTP.Idle
	}

	// Filesystem
	if override.Filesystem.LockAcquire != 0 {
		result.Filesystem.LockAcquire = override.Filesystem.LockAcquire
	}
	if override.Filesystem.StaleLock != 0 {
		result.Filesystem.StaleLock = override.Filesystem.StaleLock
	}
	if override.Filesystem.WatchDebounce != 0 {
		result.Filesystem.WatchDebounce = override.Filesystem.WatchDebounce
	}

	// CI
	if override.CI.DispatchSettle != 0 {
		result.CI.DispatchSettle = override.CI.DispatchSettle
	}
	if override.CI.PollInterval != 0 {
		result.CI.PollInterval = override.CI.PollInterval
	}
	if override.CI.WorkflowCompletion != 0 {
		result.CI.WorkflowCompletion = override.CI.WorkflowCompletion
	}
	if override.CI.RecentRunWindow != 0 {
		result.CI.RecentRunWindow = override.CI.RecentRunWindow
	}

	// LongOperations
	if override.LongOperations.SecurityScan != 0 {
		result.LongOperations.SecurityScan = override.LongOperations.SecurityScan
	}
	if override.LongOperations.Build != 0 {
		result.LongOperations.Build = override.LongOperations.Build
	}
	if override.LongOperations.Test != 0 {
		result.LongOperations.Test = override.LongOperations.Test
	}
	if override.LongOperations.EvidenceMaxAge != 0 {
		result.LongOperations.EvidenceMaxAge = override.LongOperations.EvidenceMaxAge
	}
	if override.LongOperations.WorkerTimeout != 0 {
		result.LongOperations.WorkerTimeout = override.LongOperations.WorkerTimeout
	}

	// TUI
	if override.TUI.AutoScrollResume != 0 {
		result.TUI.AutoScrollResume = override.TUI.AutoScrollResume
	}
	if override.TUI.ExitCountdown != 0 {
		result.TUI.ExitCountdown = override.TUI.ExitCountdown
	}
	if override.TUI.PortReservation != 0 {
		result.TUI.PortReservation = override.TUI.PortReservation
	}
	if override.TUI.FreezeCountdown != 0 {
		result.TUI.FreezeCountdown = override.TUI.FreezeCountdown
	}
	if override.TUI.MinDisplayTime != 0 {
		result.TUI.MinDisplayTime = override.TUI.MinDisplayTime
	}
	if override.TUI.MetricsInterval != 0 {
		result.TUI.MetricsInterval = override.TUI.MetricsInterval
	}

	// TUILayout
	if override.TUILayout.MaxTabs != 0 {
		result.TUILayout.MaxTabs = override.TUILayout.MaxTabs
	}
	if override.TUILayout.DefaultColumns != 0 {
		result.TUILayout.DefaultColumns = override.TUILayout.DefaultColumns
	}
	if override.TUILayout.MinColumns != 0 {
		result.TUILayout.MinColumns = override.TUILayout.MinColumns
	}
	if override.TUILayout.MaxColumns != 0 {
		result.TUILayout.MaxColumns = override.TUILayout.MaxColumns
	}
	if override.TUILayout.BufferSizePane != 0 {
		result.TUILayout.BufferSizePane = override.TUILayout.BufferSizePane
	}
	if override.TUILayout.BufferSizeResults != 0 {
		result.TUILayout.BufferSizeResults = override.TUILayout.BufferSizeResults
	}
	if override.TUILayout.BufferSizeUoW != 0 {
		result.TUILayout.BufferSizeUoW = override.TUILayout.BufferSizeUoW
	}

	// Scheduling
	if override.Scheduling.CapacityRecalc != 0 {
		result.Scheduling.CapacityRecalc = override.Scheduling.CapacityRecalc
	}
	return &result
}
