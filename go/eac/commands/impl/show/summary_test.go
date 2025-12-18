package show

import (
	"testing"
)

func TestFormatJobResult(t *testing.T) {
	tests := []struct {
		result string
		want   string
	}{
		{"success", ":white_check_mark:"},
		{"skipped", ":grey_question: skipped"},
		{"failure", ":x: failure"},
		{"cancelled", ":x: cancelled"},
		{"", ":grey_question: unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.result, func(t *testing.T) {
			got := formatJobResult(tt.result)
			if got != tt.want {
				t.Errorf("formatJobResult(%q) = %q, want %q", tt.result, got, tt.want)
			}
		})
	}
}

func TestFormatScanResult(t *testing.T) {
	tests := []struct {
		result string
		want   string
	}{
		{"success", ":white_check_mark:"},
		{"skipped", ":grey_question: skipped"},
		{"failure", ":warning: failure"},
		{"", ":grey_question: unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.result, func(t *testing.T) {
			got := formatScanResult(tt.result)
			if got != tt.want {
				t.Errorf("formatScanResult(%q) = %q, want %q", tt.result, got, tt.want)
			}
		})
	}
}

func TestFormatJobResult_SuccessIsCheckmark(t *testing.T) {
	result := formatJobResult("success")
	if result != ":white_check_mark:" {
		t.Errorf("success should be checkmark, got %q", result)
	}
}

func TestFormatJobResult_FailureHasX(t *testing.T) {
	result := formatJobResult("failure")
	if result[:3] != ":x:" {
		t.Errorf("failure should start with :x:, got %q", result)
	}
}

func TestFormatScanResult_FailureIsWarning(t *testing.T) {
	// Scan failures use warning emoji instead of X
	result := formatScanResult("failure")
	if result[:9] != ":warning:" {
		t.Errorf("scan failure should start with :warning:, got %q", result)
	}
}

func TestFormatJobResult_CustomError(t *testing.T) {
	customErrors := []string{
		"timeout",
		"infrastructure_failure",
		"action_required",
		"startup_failure",
	}

	for _, err := range customErrors {
		result := formatJobResult(err)
		if result[:3] != ":x:" {
			t.Errorf("formatJobResult(%q) should start with :x:, got %q", err, result)
		}
		if result != ":x: "+err {
			t.Errorf("formatJobResult(%q) = %q, want %q", err, result, ":x: "+err)
		}
	}
}

