package release

import (
	"testing"
	"time"
)

func TestGenerateCalverTag(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		time   time.Time
		want   string
	}{
		{
			name:   "standard docs tag",
			prefix: "docs",
			time:   time.Date(2025, 12, 14, 16, 30, 0, 0, time.UTC),
			want:   "docs/2025.1214.1630",
		},
		{
			name:   "eac-core tag",
			prefix: "eac-core",
			time:   time.Date(2025, 1, 5, 9, 5, 0, 0, time.UTC),
			want:   "eac-core/2025.0105.0905",
		},
		{
			name:   "midnight",
			prefix: "test",
			time:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			want:   "test/2024.0601.0000",
		},
		{
			name:   "end of day",
			prefix: "test",
			time:   time.Date(2024, 12, 31, 23, 59, 0, 0, time.UTC),
			want:   "test/2024.1231.2359",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCalverTag(tt.prefix, tt.time)
			if got != tt.want {
				t.Errorf("GenerateCalverTag(%q, %v) = %q, want %q", tt.prefix, tt.time, got, tt.want)
			}
		})
	}
}

func TestParseCalverTag(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		wantPrefix string
		wantTime   time.Time
		wantErr    bool
	}{
		{
			name:       "valid docs tag",
			tag:        "docs/2025.1214.1630",
			wantPrefix: "docs",
			wantTime:   time.Date(2025, 12, 14, 16, 30, 0, 0, time.UTC),
			wantErr:    false,
		},
		{
			name:       "valid eac-core tag",
			tag:        "eac-core/2025.0105.0905",
			wantPrefix: "eac-core",
			wantTime:   time.Date(2025, 1, 5, 9, 5, 0, 0, time.UTC),
			wantErr:    false,
		},
		{
			name:       "midnight",
			tag:        "test/2024.0601.0000",
			wantPrefix: "test",
			wantTime:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			wantErr:    false,
		},
		{
			name:    "missing prefix",
			tag:     "2025.1214.1630",
			wantErr: true,
		},
		{
			name:    "invalid version format - missing component",
			tag:     "docs/2025.1214",
			wantErr: true,
		},
		{
			name:    "invalid version format - extra component",
			tag:     "docs/2025.1214.1630.0",
			wantErr: true,
		},
		{
			name:    "invalid MMDD length",
			tag:     "docs/2025.121.1630",
			wantErr: true,
		},
		{
			name:    "invalid HHMM length",
			tag:     "docs/2025.1214.163",
			wantErr: true,
		},
		{
			name:    "empty tag",
			tag:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, parsedTime, err := ParseCalverTag(tt.tag)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseCalverTag(%q) expected error, got nil", tt.tag)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseCalverTag(%q) unexpected error: %v", tt.tag, err)
				return
			}

			if prefix != tt.wantPrefix {
				t.Errorf("ParseCalverTag(%q) prefix = %q, want %q", tt.tag, prefix, tt.wantPrefix)
			}

			if !parsedTime.Equal(tt.wantTime) {
				t.Errorf("ParseCalverTag(%q) time = %v, want %v", tt.tag, parsedTime, tt.wantTime)
			}
		})
	}
}

func TestIsValidCalverVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"2025.1214.1630", true},
		{"2024.0101.0000", true},
		{"2025.1231.2359", true},
		{"24.1214.1630", false},     // year not 4 digits
		{"2025.121.1630", false},    // MMDD not 4 digits
		{"2025.1214.163", false},    // HHMM not 4 digits
		{"2025.12141630", false},    // missing dots
		{"2025.1214", false},        // missing HHMM
		{"2025.1214.1630.0", false}, // extra component
		{"", false},                 // empty
		{"invalid", false},          // not a version
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := IsValidCalverVersion(tt.version)
			if got != tt.want {
				t.Errorf("IsValidCalverVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestGenerateAndParse_Roundtrip(t *testing.T) {
	// Test that Generate -> Parse produces the same values
	testTimes := []time.Time{
		time.Date(2025, 12, 14, 16, 30, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 6, 15, 12, 45, 0, 0, time.UTC),
	}

	prefixes := []string{"docs", "eac-core", "test-module"}

	for _, prefix := range prefixes {
		for _, tm := range testTimes {
			tag := GenerateCalverTag(prefix, tm)
			parsedPrefix, parsedTime, err := ParseCalverTag(tag)
			if err != nil {
				t.Errorf("Roundtrip failed for %s/%v: parse error %v", prefix, tm, err)
				continue
			}

			if parsedPrefix != prefix {
				t.Errorf("Roundtrip prefix mismatch: got %q, want %q", parsedPrefix, prefix)
			}

			// Compare times truncated to minute (calver doesn't store seconds)
			expectedTime := tm.Truncate(time.Minute)
			if !parsedTime.Equal(expectedTime) {
				t.Errorf("Roundtrip time mismatch: got %v, want %v", parsedTime, expectedTime)
			}
		}
	}
}

func TestIsValidCalverVersion_EdgeCases(t *testing.T) {
	// Test edge cases for date validation
	tests := []struct {
		version string
		want    bool
		desc    string
	}{
		{"2025.0229.1200", false, "invalid date Feb 29 in non-leap year"},
		{"2024.0229.1200", true, "valid date Feb 29 in leap year"},
		{"2025.1332.1200", false, "invalid month 13"},
		{"2025.0032.1200", false, "invalid month 00"},
		{"2025.0100.1200", false, "invalid day 00"},
		{"2025.0132.1200", false, "invalid day 32"},
		{"2025.0115.2500", false, "invalid hour 25"},
		{"2025.0115.1260", false, "invalid minute 60"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := IsValidCalverVersion(tt.version)
			if got != tt.want {
				t.Errorf("IsValidCalverVersion(%q) = %v, want %v (%s)", tt.version, got, tt.want, tt.desc)
			}
		})
	}
}
