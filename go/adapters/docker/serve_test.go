//go:build L1 && ov
// +build L1,ov

package docker

import (
	"testing"
)

func TestIsPortAvailable(t *testing.T) {
	// Test that the function works without panicking
	// Note: We can't test specific ports reliably in CI since they may be in use
	t.Run("function_executes", func(t *testing.T) {
		// Just verify the function doesn't panic
		_ = IsPortAvailable(59999)
	})
}

func TestFindAvailablePort(t *testing.T) {
	t.Run("finds_port_in_range", func(t *testing.T) {
		port, err := FindAvailablePort()
		if err != nil {
			t.Skipf("No ports available in test environment: %v", err)
		}

		if port < PortRangeStart || port > PortRangeEnd {
			t.Errorf("Port %d is outside range %d-%d", port, PortRangeStart, PortRangeEnd)
		}
	})
}

func TestFindAvailablePortOrUse(t *testing.T) {
	tests := []struct {
		name          string
		preferredPort int
		wantInRange   bool
	}{
		{
			name:          "zero_prefers_auto_allocation",
			preferredPort: 0,
			wantInRange:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := FindAvailablePortOrUse(tt.preferredPort)
			if err != nil {
				t.Skipf("No ports available: %v", err)
			}

			if tt.wantInRange && (port < PortRangeStart || port > PortRangeEnd) {
				t.Errorf("Port %d is outside range %d-%d", port, PortRangeStart, PortRangeEnd)
			}
		})
	}
}
