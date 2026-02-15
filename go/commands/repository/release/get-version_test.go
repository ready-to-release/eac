package release

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionInfo_Struct(t *testing.T) {
	info := VersionInfo{
		Module:      "clie",
		Version:     "1.2.3",
		Tag:         "clie/1.2.3",
		Date:        "2025-12-14",
		VersionType: "semver",
	}

	if info.Module != "clie" {
		t.Errorf("Module = %q, want %q", info.Module, "clie")
	}
	if info.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", info.Version, "1.2.3")
	}
	if info.Tag != "clie/1.2.3" {
		t.Errorf("Tag = %q, want %q", info.Tag, "clie/1.2.3")
	}
	if info.VersionType != "semver" {
		t.Errorf("VersionType = %q, want %q", info.VersionType, "semver")
	}
}

func TestVersionInfo_JSONMarshal(t *testing.T) {
	info := VersionInfo{
		Module:      "docs",
		Version:     "2025.1214.1630",
		Tag:         "docs/2025.1214.1630",
		VersionType: "calver",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded VersionInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Module != info.Module {
		t.Errorf("Module = %q, want %q", decoded.Module, info.Module)
	}
	if decoded.Version != info.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, info.Version)
	}
	if decoded.VersionType != info.VersionType {
		t.Errorf("VersionType = %q, want %q", decoded.VersionType, info.VersionType)
	}
}

func TestVersionInfo_OmitEmptyDate(t *testing.T) {
	info := VersionInfo{
		Module:      "test",
		Version:     "1.0.0",
		Tag:         "test/1.0.0",
		VersionType: "semver",
		// Date is empty
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Empty date should be omitted from JSON
	jsonStr := string(data)
	if strings.Contains(jsonStr, "date") {
		t.Error("empty date should be omitted from JSON")
	}
}

func TestVersionInfo_WithDate(t *testing.T) {
	info := VersionInfo{
		Module:      "test",
		Version:     "1.0.0",
		Tag:         "test/1.0.0",
		Date:        "2025-01-15",
		VersionType: "semver",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, "2025-01-15") {
		t.Error("date should be present in JSON")
	}
}
