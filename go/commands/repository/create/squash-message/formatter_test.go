package squashmessage

import (
	"strings"
	"testing"
)

func TestFormatSquashMessage_BasicFormat(t *testing.T) {
	input := `{
		"type": "feat",
		"scope": "auth",
		"subject": "add OAuth2 support",
		"body": "Implemented OAuth2 authentication flow with PKCE.",
		"auditor_summary": "New authentication mechanism added."
	}`

	result, err := FormatSquashMessage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(result, "feat(auth): add OAuth2 support\n") {
		t.Errorf("expected header 'feat(auth): add OAuth2 support', got first line: %s", strings.SplitN(result, "\n", 2)[0])
	}

	if !strings.Contains(result, "Auditor-Summary: New authentication mechanism added.") {
		t.Error("expected auditor summary in output")
	}

	if !strings.Contains(result, "Implemented OAuth2 authentication flow with PKCE.") {
		t.Error("expected body in output")
	}
}

func TestFormatSquashMessage_WithBreakingChange(t *testing.T) {
	input := `{
		"type": "feat",
		"scope": "api",
		"subject": "change response format",
		"body": "Migrated to JSON:API response format.",
		"breaking": true,
		"breaking_description": "Response structure changed from flat to JSON:API envelope"
	}`

	result, err := FormatSquashMessage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "BREAKING CHANGE: Response structure changed from flat to JSON:API envelope") {
		t.Error("expected breaking change notice in output")
	}
}

func TestFormatSquashMessage_WithChanges(t *testing.T) {
	input := `{
		"type": "refactor",
		"scope": "core",
		"subject": "restructure modules",
		"body": "Split large modules into focused packages.",
		"changes": [
			{"type": "refactor", "scope": "config", "description": "extracted config loading"},
			{"type": "feat", "description": "added validation helpers"}
		]
	}`

	result, err := FormatSquashMessage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "- refactor(config): extracted config loading") {
		t.Error("expected scoped change entry in output")
	}

	if !strings.Contains(result, "- feat: added validation helpers") {
		t.Error("expected unscoped change entry in output")
	}
}

func TestFormatSquashMessage_WithModules(t *testing.T) {
	input := `{
		"type": "feat",
		"scope": "multi",
		"subject": "update dependencies",
		"body": "Updated across multiple modules.",
		"modules": ["core", "cli", "adapters"]
	}`

	result, err := FormatSquashMessage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Modules: core, cli, adapters") {
		t.Error("expected modules list in output")
	}
}

func TestFormatSquashMessage_WithPRNumber(t *testing.T) {
	input := `{
		"type": "fix",
		"scope": "build",
		"subject": "fix race condition",
		"body": "Fixed concurrent map access.",
		"pr_number": 42
	}`

	result, err := FormatSquashMessage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "PR: #42") {
		t.Error("expected PR reference in output")
	}
}

func TestFormatSquashMessage_AuditorSummaryFallback(t *testing.T) {
	input := `{
		"type": "fix",
		"scope": "core",
		"subject": "fix parsing bug",
		"body": "Fixed a parsing issue in the config loader. Additional details follow."
	}`

	result, err := FormatSquashMessage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Auditor-Summary:") {
		t.Error("expected auditor summary fallback from body")
	}
}

func TestFormatSquashMessage_InvalidJSON(t *testing.T) {
	_, err := FormatSquashMessage("not json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFormatSquashMessage_EndsWithNewline(t *testing.T) {
	input := `{
		"type": "chore",
		"scope": "deps",
		"subject": "bump versions",
		"body": "Updated all dependencies."
	}`

	result, err := FormatSquashMessage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(result, "\n") {
		t.Error("expected result to end with newline")
	}

	if strings.HasSuffix(result, "\n\n") {
		t.Error("expected result to end with single newline, not double")
	}
}
