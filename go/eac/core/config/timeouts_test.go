package config

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestDurationUnmarshalYAML tests Duration YAML parsing.
func TestDurationUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"milliseconds", "500ms", 500 * time.Millisecond, false},
		{"seconds", "30s", 30 * time.Second, false},
		{"minutes", "5m", 5 * time.Minute, false},
		{"hours", "2h", 2 * time.Hour, false},
		{"combined", "1h30m", 90 * time.Minute, false},
		{"invalid", "invalid", 0, true},
		{"empty", "", 0, false}, // empty parses to 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := yaml.Unmarshal([]byte(tt.input), &d)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && d.D() != tt.want {
				t.Errorf("UnmarshalYAML() = %v, want %v", d.D(), tt.want)
			}
		})
	}
}

// TestDurationMarshalYAML tests Duration YAML serialization.
func TestDurationMarshalYAML(t *testing.T) {
	tests := []struct {
		name  string
		input Duration
		want  string
	}{
		{"milliseconds", Duration(500 * time.Millisecond), "500ms"},
		{"seconds", Duration(30 * time.Second), "30s"},
		{"minutes", Duration(5 * time.Minute), "5m0s"},
		{"hours", Duration(2 * time.Hour), "2h0m0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.MarshalYAML()
			if err != nil {
				t.Errorf("MarshalYAML() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("MarshalYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDefaultTimeoutConfig tests default timeout values.
func TestDefaultTimeoutConfig(t *testing.T) {
	cfg := DefaultTimeoutConfig()

	// Docker defaults
	if cfg.Docker.Query.D() != 5*time.Second {
		t.Errorf("Docker.Query = %v, want 5s", cfg.Docker.Query.D())
	}
	if cfg.Docker.ImagePull.D() != 5*time.Minute {
		t.Errorf("Docker.ImagePull = %v, want 5m", cfg.Docker.ImagePull.D())
	}
	if cfg.Docker.ImageBuild.D() != 30*time.Minute {
		t.Errorf("Docker.ImageBuild = %v, want 30m", cfg.Docker.ImageBuild.D())
	}
	if cfg.Docker.ContainerStart.D() != 60*time.Second {
		t.Errorf("Docker.ContainerStart = %v, want 60s", cfg.Docker.ContainerStart.D())
	}
	if cfg.Docker.ContainerStop.D() != 5*time.Second {
		t.Errorf("Docker.ContainerStop = %v, want 5s", cfg.Docker.ContainerStop.D())
	}
	if cfg.Docker.ContainerExec.D() != 30*time.Minute {
		t.Errorf("Docker.ContainerExec = %v, want 30m", cfg.Docker.ContainerExec.D())
	}
	if cfg.Docker.Validation.D() != 30*time.Second {
		t.Errorf("Docker.Validation = %v, want 30s", cfg.Docker.Validation.D())
	}

	// Filesystem defaults
	if cfg.Filesystem.LockAcquire.D() != 30*time.Second {
		t.Errorf("Filesystem.LockAcquire = %v, want 30s", cfg.Filesystem.LockAcquire.D())
	}
	if cfg.Filesystem.StaleLock.D() != 5*time.Minute {
		t.Errorf("Filesystem.StaleLock = %v, want 5m", cfg.Filesystem.StaleLock.D())
	}

	// CI defaults
	if cfg.CI.DispatchSettle.D() != 2*time.Second {
		t.Errorf("CI.DispatchSettle = %v, want 2s", cfg.CI.DispatchSettle.D())
	}
	if cfg.CI.PollInterval.D() != 10*time.Second {
		t.Errorf("CI.PollInterval = %v, want 10s", cfg.CI.PollInterval.D())
	}
	if cfg.CI.RecentRunWindow.D() != 2*time.Hour {
		t.Errorf("CI.RecentRunWindow = %v, want 2h", cfg.CI.RecentRunWindow.D())
	}

	// Long operations defaults
	if cfg.LongOperations.SecurityScan.D() != 30*time.Minute {
		t.Errorf("LongOperations.SecurityScan = %v, want 30m", cfg.LongOperations.SecurityScan.D())
	}
	if cfg.LongOperations.EvidenceMaxAge.D() != 24*time.Hour {
		t.Errorf("LongOperations.EvidenceMaxAge = %v, want 24h", cfg.LongOperations.EvidenceMaxAge.D())
	}

	// TUI defaults
	if cfg.TUI.AutoScrollResume.D() != 8*time.Second {
		t.Errorf("TUI.AutoScrollResume = %v, want 8s", cfg.TUI.AutoScrollResume.D())
	}
	if cfg.TUI.PortReservation.D() != 30*time.Second {
		t.Errorf("TUI.PortReservation = %v, want 30s", cfg.TUI.PortReservation.D())
	}

	// Scheduling defaults
	if cfg.Scheduling.CapacityRecalc.D() != 2*time.Second {
		t.Errorf("Scheduling.CapacityRecalc = %v, want 2s", cfg.Scheduling.CapacityRecalc.D())
	}
}

// TestTimeoutConfigFromYAML tests loading timeout config from YAML.
func TestTimeoutConfigFromYAML(t *testing.T) {
	yamlContent := `
docker:
  query: 10s
  image_pull: 10m
filesystem:
  lock_acquire: 60s
ci:
  poll_interval: 5s
`
	var cfg TimeoutConfig
	if err := yaml.Unmarshal([]byte(yamlContent), &cfg); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	if cfg.Docker.Query.D() != 10*time.Second {
		t.Errorf("Docker.Query = %v, want 10s", cfg.Docker.Query.D())
	}
	if cfg.Docker.ImagePull.D() != 10*time.Minute {
		t.Errorf("Docker.ImagePull = %v, want 10m", cfg.Docker.ImagePull.D())
	}
	if cfg.Filesystem.LockAcquire.D() != 60*time.Second {
		t.Errorf("Filesystem.LockAcquire = %v, want 60s", cfg.Filesystem.LockAcquire.D())
	}
	if cfg.CI.PollInterval.D() != 5*time.Second {
		t.Errorf("CI.PollInterval = %v, want 5s", cfg.CI.PollInterval.D())
	}
}

// TestMergeTimeoutConfigs tests merging user overrides with defaults.
func TestMergeTimeoutConfigs(t *testing.T) {
	defaults := DefaultTimeoutConfig()

	// User override for just one value
	userOverride := &TimeoutConfig{
		Docker: DockerTimeouts{
			Query: Duration(10 * time.Second),
		},
	}

	merged := MergeTimeoutConfigs(defaults, userOverride)

	// Overridden value should be from user
	if merged.Docker.Query.D() != 10*time.Second {
		t.Errorf("Docker.Query = %v, want 10s (user override)", merged.Docker.Query.D())
	}

	// Non-overridden values should keep defaults
	if merged.Docker.ImagePull.D() != 5*time.Minute {
		t.Errorf("Docker.ImagePull = %v, want 5m (default)", merged.Docker.ImagePull.D())
	}
	if merged.Filesystem.LockAcquire.D() != 30*time.Second {
		t.Errorf("Filesystem.LockAcquire = %v, want 30s (default)", merged.Filesystem.LockAcquire.D())
	}
}
