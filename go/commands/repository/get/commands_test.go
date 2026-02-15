//go:build L0 && ov

package get

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommands_ReturnsAllCommands(t *testing.T) {
	cmds := Commands()
	assert.NotEmpty(t, cmds, "Commands() should return registered commands")

	// Verify command count matches expected (update when new commands added)
	assert.GreaterOrEqual(t, len(cmds), 30, "Expected at least 30 get subcommands")

	// Verify each command has required metadata
	for _, cmd := range cmds {
		name := cmd.Name()
		assert.NotEmpty(t, name, "Each command must have a name")
		meta := cmd.Metadata()
		assert.NotEmpty(t, meta.Short, "Command %q must have a short description", name)
	}
}

func TestCommands_NoDuplicateNames(t *testing.T) {
	cmds := Commands()
	seen := make(map[string]bool)
	for _, cmd := range cmds {
		name := cmd.Name()
		assert.False(t, seen[name], "Duplicate command name: %s", name)
		seen[name] = true
	}
}

func TestArtifactBuildModes(t *testing.T) {
	modes := ArtifactBuildModes{
		Default: []string{"binary-linux-amd64"},
		All:     []string{"binary-linux-amd64", "binary-darwin-arm64"},
	}
	assert.Len(t, modes.Default, 1)
	assert.Len(t, modes.All, 2)
}

func TestModuleCIStatus_StructFields(t *testing.T) {
	status := ModuleCIStatus{
		Reason: "files_changed_since_abc123",
	}
	assert.Equal(t, "files_changed_since_abc123", status.Reason)
}

func TestFormatTriggerReason(t *testing.T) {
	tests := []struct {
		name   string
		reason string
		want   string
	}{
		{"dependency", "dependency eac needs CI", "eac changed"},
		{"files changed", "files_changed_since_abc1234", "files changed"},
		{"no history", "no_ci_history", "no previous CI"},
		{"query failed", "query_failed", "CI query failed"},
		{"unknown", "other_reason", "other_reason"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTriggerReason(tt.reason)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Note: TestFormatPlatformName and TestBinarySize tests are in binary-sizes_test.go
