package scoring

import (
	"testing"
)

func TestCalculateBaseLikelihood(t *testing.T) {
	tests := []struct {
		name     string
		critical int
		high     int
		medium   int
		low      int
		want     int
	}{
		{
			name:     "no findings",
			critical: 0,
			high:     0,
			medium:   0,
			low:      0,
			want:     1,
		},
		{
			name:     "only low findings",
			critical: 0,
			high:     0,
			medium:   0,
			low:      5,
			want:     2,
		},
		{
			name:     "medium findings",
			critical: 0,
			high:     0,
			medium:   3,
			low:      0,
			want:     3,
		},
		{
			name:     "high findings",
			critical: 0,
			high:     2,
			medium:   0,
			low:      0,
			want:     4,
		},
		{
			name:     "critical findings",
			critical: 1,
			high:     0,
			medium:   0,
			low:      0,
			want:     5,
		},
		{
			name:     "mixed findings",
			critical: 0,
			high:     1,
			medium:   2,
			low:      3,
			want:     4, // High takes precedence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateBaseLikelihood(tt.critical, tt.high, tt.medium, tt.low)
			if got != tt.want {
				t.Errorf("CalculateBaseLikelihood() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateRiskScore(t *testing.T) {
	tests := []struct {
		name       string
		likelihood int
		impact     int
		want       int
	}{
		{
			name:       "low risk",
			likelihood: 1,
			impact:     1,
			want:       1,
		},
		{
			name:       "medium risk",
			likelihood: 3,
			impact:     3,
			want:       9,
		},
		{
			name:       "high risk",
			likelihood: 4,
			impact:     4,
			want:       16,
		},
		{
			name:       "critical risk",
			likelihood: 5,
			impact:     5,
			want:       25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateRiskScore(tt.likelihood, tt.impact)
			if got != tt.want {
				t.Errorf("CalculateRiskScore() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDetermineRiskBand(t *testing.T) {
	// Thresholds: 0-5 Low, 6-11 Medium, 12-19 High, 20+ Critical
	tests := []struct {
		name  string
		score int
		want  RiskBand
	}{
		{name: "score 1 is Low", score: 1, want: RiskLow},
		{name: "score 5 is Low", score: 5, want: RiskLow},
		{name: "score 6 is Medium", score: 6, want: RiskMedium},
		{name: "score 11 is Medium", score: 11, want: RiskMedium},
		{name: "score 12 is High", score: 12, want: RiskHigh},
		{name: "score 19 is High", score: 19, want: RiskHigh},
		{name: "score 20 is Critical", score: 20, want: RiskCritical},
		{name: "score 25 is Critical", score: 25, want: RiskCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineRiskBand(tt.score)
			if got != tt.want {
				t.Errorf("DetermineRiskBand(%d) = %s, want %s", tt.score, got, tt.want)
			}
		})
	}
}

func TestComputeRiskScore(t *testing.T) {
	tests := []struct {
		name       string
		module     string
		likelihood int
		impact     int
		reasoning  string
		wantBand   RiskBand
	}{
		{
			name:       "low risk module",
			module:     "test-module",
			likelihood: 1,
			impact:     2,
			reasoning:  "No vulnerabilities found",
			wantBand:   RiskLow,
		},
		{
			name:       "critical risk module",
			module:     "critical-module",
			likelihood: 5,
			impact:     5,
			reasoning:  "Critical vulnerabilities with active exploits",
			wantBand:   RiskCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := ComputeRiskScore(tt.module, tt.likelihood, tt.impact, tt.reasoning)

			if rs == nil {
				t.Fatal("ComputeRiskScore returned nil")
			}

			if rs.Module != tt.module {
				t.Errorf("Module = %s, want %s", rs.Module, tt.module)
			}

			if rs.Likelihood != tt.likelihood {
				t.Errorf("Likelihood = %d, want %d", rs.Likelihood, tt.likelihood)
			}

			if rs.Impact != tt.impact {
				t.Errorf("Impact = %d, want %d", rs.Impact, tt.impact)
			}

			expectedScore := tt.likelihood * tt.impact
			if rs.Score != expectedScore {
				t.Errorf("Score = %d, want %d", rs.Score, expectedScore)
			}

			if rs.Band != tt.wantBand {
				t.Errorf("Band = %s, want %s", rs.Band, tt.wantBand)
			}

			if rs.Reasoning != tt.reasoning {
				t.Errorf("Reasoning = %s, want %s", rs.Reasoning, tt.reasoning)
			}
		})
	}
}

func TestGetDefaultImpact(t *testing.T) {
	tests := []struct {
		name       string
		moduleType string
		want       int
	}{
		{name: "service type", moduleType: "service", want: 4},
		{name: "library type", moduleType: "library", want: 3},
		{name: "unknown type", moduleType: "unknown", want: 3},
		{name: "empty type", moduleType: "", want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDefaultImpact(tt.moduleType)
			if got != tt.want {
				t.Errorf("GetDefaultImpact(%s) = %d, want %d", tt.moduleType, got, tt.want)
			}
		})
	}
}

func TestRiskBandString(t *testing.T) {
	tests := []struct {
		band RiskBand
		want string
	}{
		{RiskLow, "Low"},
		{RiskMedium, "Medium"},
		{RiskHigh, "High"},
		{RiskCritical, "Critical"},
	}

	for _, tt := range tests {
		t.Run(string(tt.band), func(t *testing.T) {
			if string(tt.band) != tt.want {
				t.Errorf("RiskBand string = %s, want %s", string(tt.band), tt.want)
			}
		})
	}
}

func TestFormatRiskScore(t *testing.T) {
	rs := &RiskScore{
		Module:     "test",
		Likelihood: 3,
		Impact:     4,
		Score:      12,
		Band:       RiskHigh,
		Reasoning:  "Test reasoning",
	}

	formatted := FormatRiskScore(rs)
	if formatted == "" {
		t.Error("FormatRiskScore returned empty string")
	}

	// Should contain the score
	if !contains(formatted, "12") {
		t.Errorf("FormatRiskScore should contain score '12', got: %s", formatted)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
