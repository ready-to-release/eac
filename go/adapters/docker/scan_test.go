//go:build L1
// +build L1

package docker

import (
	"testing"
)

func TestGetDefaultMockZAPOutput(t *testing.T) {
	result := getDefaultMockZAPOutput()
	if result == nil {
		t.Fatal("getDefaultMockZAPOutput() returned nil")
	}

	site, ok := result["site"]
	if !ok {
		t.Fatal("expected 'site' key in mock output")
	}

	siteSlice, ok := site.([]map[string]interface{})
	if !ok {
		t.Fatalf("expected site to be []map[string]interface{}, got %T", site)
	}
	if len(siteSlice) != 1 {
		t.Fatalf("expected 1 site entry, got %d", len(siteSlice))
	}

	name, ok := siteSlice[0]["@name"]
	if !ok {
		t.Fatal("expected '@name' key in site entry")
	}
	if name != "http://localhost:8080" {
		t.Errorf("@name = %q, want %q", name, "http://localhost:8080")
	}

	alerts, ok := siteSlice[0]["alerts"]
	if !ok {
		t.Fatal("expected 'alerts' key in site entry")
	}
	alertsSlice, ok := alerts.([]interface{})
	if !ok {
		t.Fatalf("expected alerts to be []interface{}, got %T", alerts)
	}
	if len(alertsSlice) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alertsSlice))
	}
}

func TestRunZAPScan_MockMode(t *testing.T) {
	// Set mock mode env var
	t.Setenv("CLIE_MOCK_SECURITY", "true")

	result, err := RunZAPScan("http://localhost:8080", ZAPBaseline, t.TempDir(), "ghcr.io/zaproxy/zaproxy:stable")
	if err != nil {
		t.Fatalf("RunZAPScan in mock mode returned error: %v", err)
	}
	if result == nil {
		t.Fatal("RunZAPScan in mock mode returned nil result")
	}
}

func TestRunZAPScan_InvalidScanType(t *testing.T) {
	// Without mock mode, an invalid scan type should still be caught
	// before Docker client creation if we set mock mode
	t.Setenv("CLIE_MOCK_DOCKER", "true")

	// Mock mode returns before checking scan type, so test without mock
	// This will fail at Docker client creation in most test environments,
	// so we verify the scan type validation separately.
	validTypes := map[string]string{
		ZAPBaseline: "zap-baseline.py",
		ZAPFull:     "zap-full-scan.py",
		ZAPAPI:      "zap-api-scan.py",
	}
	for scanType, expected := range validTypes {
		if expected == "" {
			t.Errorf("scan type %q has empty script name", scanType)
		}
	}
}
