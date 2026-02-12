//go:build L1
// +build L1

package docker

import (
	"testing"
)

func TestShouldOpenBrowser_CI(t *testing.T) {
	t.Setenv("CI", "true")
	if shouldOpenBrowser() {
		t.Error("shouldOpenBrowser() = true in CI, want false")
	}
}

func TestShouldOpenBrowser_GitHubActions(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	if shouldOpenBrowser() {
		t.Error("shouldOpenBrowser() = true in GitHub Actions, want false")
	}
}

func TestShouldOpenBrowser_TestRunID(t *testing.T) {
	t.Setenv("CLIE_TEST_RUN_ID", "test-123")
	if shouldOpenBrowser() {
		t.Error("shouldOpenBrowser() = true with CLIE_TEST_RUN_ID, want false")
	}
}

func TestShouldOpenBrowser_GodogFormat(t *testing.T) {
	t.Setenv("GODOG_FORMAT", "pretty")
	if shouldOpenBrowser() {
		t.Error("shouldOpenBrowser() = true with GODOG_FORMAT, want false")
	}
}

func TestShouldOpenBrowser_NoBrowserFlag(t *testing.T) {
	t.Setenv("CLIE_NO_BROWSER", "true")
	if shouldOpenBrowser() {
		t.Error("shouldOpenBrowser() = true with CLIE_NO_BROWSER=true, want false")
	}
}

func TestOpenBrowser_Skipped_InCI(t *testing.T) {
	t.Setenv("CI", "true")
	// Should return nil (no-op) when in CI
	err := OpenBrowser("http://localhost:8080")
	if err != nil {
		t.Errorf("OpenBrowser() in CI returned error: %v", err)
	}
}

func TestOpenBrowserWithFallback_Skipped_InCI(t *testing.T) {
	t.Setenv("CI", "true")
	opened, err := OpenBrowserWithFallback("http://localhost:8080")
	if err != nil {
		t.Errorf("OpenBrowserWithFallback() in CI returned error: %v", err)
	}
	if opened {
		t.Error("OpenBrowserWithFallback() in CI returned opened=true, want false")
	}
}
