//go:build L1

package orchestrator

import (
	"testing"
)

func TestParseMeminfo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{
			name: "standard meminfo",
			input: `MemTotal:        6039696 kB
MemFree:          123456 kB
MemAvailable:    4567890 kB
Buffers:          234567 kB
`,
			expected: 6039696 * 1024,
		},
		{
			name: "large memory",
			input: `MemTotal:       16384000 kB
MemFree:         8000000 kB
`,
			expected: 16384000 * 1024,
		},
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
		{
			name: "no MemTotal line",
			input: `MemFree:          123456 kB
MemAvailable:    4567890 kB
`,
			expected: 0,
		},
		{
			name:     "malformed line",
			input:    "MemTotal: notanumber kB",
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseMeminfo(tc.input)
			if result != tc.expected {
				t.Errorf("parseMeminfo() = %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestParseDockerInfoMemory(t *testing.T) {
	// Pre-calculate float-based expected values (5.762 GiB in bytes, accounting for float precision)
	expected5762GiB := uint64(6186900389)

	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{
			name: "standard GiB format",
			input: `Client:
 Version:           24.0.7
Server:
 Server Version:    24.0.7
 Total Memory: 5.762GiB
 CPUs: 8
`,
			expected: expected5762GiB,
		},
		{
			name: "MiB format",
			input: `Total Memory: 5762MiB
`,
			expected: 5762 * 1024 * 1024,
		},
		{
			name: "GB format",
			input: `Total Memory: 6GB
`,
			expected: 6 * 1024 * 1024 * 1024,
		},
		{
			name:     "no memory line",
			input:    `Server Version: 24.0.7`,
			expected: 0,
		},
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseDockerInfoMemory(tc.input)
			if result != tc.expected {
				t.Errorf("parseDockerInfoMemory() = %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestParseDockerMemoryValue(t *testing.T) {
	// Pre-calculate float-based expected values (5.762 GiB in bytes, accounting for float precision)
	expected5762GiB := uint64(6186900389)

	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{"GiB", "5.762GiB", expected5762GiB},
		{"MiB", "5762MiB", 5762 * 1024 * 1024},
		{"TiB", "1TiB", 1024 * 1024 * 1024 * 1024},
		{"KiB", "1024KiB", 1024 * 1024},
		{"GB", "6GB", 6 * 1024 * 1024 * 1024},
		{"MB", "4096MB", 4096 * 1024 * 1024},
		{"G short", "6G", 6 * 1024 * 1024 * 1024},
		{"M short", "4096M", 4096 * 1024 * 1024},
		{"empty", "", 0},
		{"invalid", "notamemory", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseDockerMemoryValue(tc.input)
			if result != tc.expected {
				t.Errorf("parseDockerMemoryValue(%q) = %d, want %d", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGetEffectiveMemoryBytes(t *testing.T) {
	// This is an integration test that depends on the actual environment.
	// It should return a non-zero value if Docker is running or .wslconfig exists,
	// or 0 if neither is available (which is fine for the fallback behavior).
	mem := GetEffectiveMemoryBytes()
	t.Logf("GetEffectiveMemoryBytes() = %d bytes (%.2f GB)", mem, float64(mem)/(1024*1024*1024))

	// We can't assert a specific value, but we can verify it's reasonable if non-zero
	if mem > 0 {
		// Should be at least 1GB and less than 1TB
		if mem < 1*1024*1024*1024 {
			t.Errorf("Memory seems too low: %d bytes", mem)
		}
		if mem > 1024*1024*1024*1024 {
			t.Errorf("Memory seems too high: %d bytes", mem)
		}
	}
}

func TestGetDockerMemoryBytes(t *testing.T) {
	// Integration test - only runs if Docker is available
	mem := GetDockerMemoryBytes()
	t.Logf("GetDockerMemoryBytes() = %d bytes (%.2f GB)", mem, float64(mem)/(1024*1024*1024))

	// If Docker is available, memory should be reasonable
	if mem > 0 {
		if mem < 1*1024*1024*1024 {
			t.Errorf("Docker memory seems too low: %d bytes", mem)
		}
		if mem > 1024*1024*1024*1024 {
			t.Errorf("Docker memory seems too high: %d bytes", mem)
		}
	}
}

func TestGetWSLMemoryBytes(t *testing.T) {
	// Integration test - only runs on Windows with WSL
	mem := GetWSLMemoryBytes()
	t.Logf("GetWSLMemoryBytes() = %d bytes (%.2f GB)", mem, float64(mem)/(1024*1024*1024))

	// If WSL is available, memory should be reasonable
	if mem > 0 {
		if mem < 1*1024*1024*1024 {
			t.Errorf("WSL memory seems too low: %d bytes", mem)
		}
		if mem > 1024*1024*1024*1024 {
			t.Errorf("WSL memory seems too high: %d bytes", mem)
		}
	}
}

func TestParseMeminfoStats(t *testing.T) {
	input := `MemTotal:        6039696 kB
MemFree:          123456 kB
MemAvailable:    4567890 kB
Buffers:          234567 kB
`
	stats := parseMeminfoStats(input)

	expectedTotal := uint64(6039696 * 1024)
	expectedAvail := uint64(4567890 * 1024)

	if stats.TotalBytes != expectedTotal {
		t.Errorf("TotalBytes = %d, want %d", stats.TotalBytes, expectedTotal)
	}
	if stats.AvailableBytes != expectedAvail {
		t.Errorf("AvailableBytes = %d, want %d", stats.AvailableBytes, expectedAvail)
	}
	if stats.UsedBytes != expectedTotal-expectedAvail {
		t.Errorf("UsedBytes = %d, want %d", stats.UsedBytes, expectedTotal-expectedAvail)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0MB"},
		{500 * 1024 * 1024, "500MB"},
		{1024 * 1024 * 1024, "1.00GB"},
		{uint64(1.5 * 1024 * 1024 * 1024), "1.50GB"},
		{6 * 1024 * 1024 * 1024, "6.00GB"},
	}

	for _, tc := range tests {
		result := FormatBytes(tc.bytes)
		if result != tc.expected {
			t.Errorf("FormatBytes(%d) = %q, want %q", tc.bytes, result, tc.expected)
		}
	}
}

func TestGetEffectiveCPUs(t *testing.T) {
	cpus := GetEffectiveCPUs()
	dockerCPUs := GetDockerCPUs()
	t.Logf("GetEffectiveCPUs() = %d, GetDockerCPUs() = %d", cpus, dockerCPUs)

	if cpus < 1 {
		t.Errorf("GetEffectiveCPUs() = %d, want >= 1", cpus)
	}
}

func TestGetMemoryStats(t *testing.T) {
	// Integration test
	stats := GetMemoryStats()
	t.Logf("GetMemoryStats() = total=%s used=%s avail=%s (%.1f%%)",
		FormatBytes(stats.TotalBytes), FormatBytes(stats.UsedBytes),
		FormatBytes(stats.AvailableBytes), stats.UsedPercent)

	// Should have some memory info
	if stats.TotalBytes > 0 {
		if stats.UsedPercent < 0 || stats.UsedPercent > 100 {
			t.Errorf("UsedPercent out of range: %.1f", stats.UsedPercent)
		}
	}
}
