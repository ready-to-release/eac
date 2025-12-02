//go:build L1

package testing

import "testing"

func TestSharedTestContext_NewSharedTestContext(t *testing.T) {
	ctx := NewSharedTestContext()

	if ctx == nil {
		t.Fatal("NewSharedTestContext() returned nil")
	}

	if ctx.AffectedModules == nil {
		t.Error("AffectedModules should be initialized")
	}

	if ctx.ValidationErrors == nil {
		t.Error("ValidationErrors should be initialized")
	}

	if ctx.CreatedFiles == nil {
		t.Error("CreatedFiles should be initialized")
	}
}

func TestSharedTestContext_Reset(t *testing.T) {
	ctx := NewSharedTestContext()

	// Set some values
	ctx.CommandOutput = "test output"
	ctx.ExitCode = 1
	ctx.TestCommitMessage = "test message"
	ctx.AffectedModules = []string{"mod1", "mod2"}
	ctx.OriginalRepoRoot = "/original/root"
	ctx.IsolatedTestProjectDir = "/isolated/dir"

	ctx.Reset()

	if ctx.CommandOutput != "" {
		t.Error("CommandOutput should be empty after Reset")
	}
	if ctx.ExitCode != 0 {
		t.Error("ExitCode should be 0 after Reset")
	}
	if ctx.TestCommitMessage != "" {
		t.Error("TestCommitMessage should be empty after Reset")
	}
	if len(ctx.AffectedModules) != 0 {
		t.Error("AffectedModules should be empty after Reset")
	}

	// Isolation fields should NOT be reset
	if ctx.OriginalRepoRoot != "/original/root" {
		t.Error("OriginalRepoRoot should NOT be reset")
	}
	if ctx.IsolatedTestProjectDir != "/isolated/dir" {
		t.Error("IsolatedTestProjectDir should NOT be reset")
	}
}

func TestSharedTestContext_PointerSharing(t *testing.T) {
	// Verify that changes through one pointer are visible through another
	ctx := NewSharedTestContext()
	ptr1 := ctx
	ptr2 := ctx

	ptr1.CommandOutput = "from ptr1"

	if ptr2.CommandOutput != "from ptr1" {
		t.Error("Changes through ptr1 should be visible through ptr2")
	}

	ptr2.ExitCode = 42

	if ptr1.ExitCode != 42 {
		t.Error("Changes through ptr2 should be visible through ptr1")
	}
}

func TestSharedTestContext_HasOutput(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*SharedTestContext)
		expected bool
	}{
		{
			name:     "empty context",
			setup:    func(c *SharedTestContext) {},
			expected: false,
		},
		{
			name: "has output",
			setup: func(c *SharedTestContext) {
				c.CommandOutput = "some output"
			},
			expected: true,
		},
		{
			name: "non-zero exit code",
			setup: func(c *SharedTestContext) {
				c.ExitCode = 1
			},
			expected: true,
		},
		{
			name: "both output and exit code",
			setup: func(c *SharedTestContext) {
				c.CommandOutput = "output"
				c.ExitCode = 1
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewSharedTestContext()
			tt.setup(ctx)

			if got := ctx.HasOutput(); got != tt.expected {
				t.Errorf("HasOutput() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSharedTestContext_AddMethods(t *testing.T) {
	ctx := NewSharedTestContext()

	ctx.AddValidationError("error1")
	ctx.AddValidationError("error2")
	if len(ctx.ValidationErrors) != 2 {
		t.Errorf("ValidationErrors should have 2 items, got %d", len(ctx.ValidationErrors))
	}

	ctx.AddCreatedFile("/path/to/file1")
	ctx.AddCreatedFile("/path/to/file2")
	if len(ctx.CreatedFiles) != 2 {
		t.Errorf("CreatedFiles should have 2 items, got %d", len(ctx.CreatedFiles))
	}

	ctx.AddAffectedModule("mod1")
	if len(ctx.AffectedModules) != 1 {
		t.Errorf("AffectedModules should have 1 item, got %d", len(ctx.AffectedModules))
	}
}

func TestSharedTestContext_SetIsolation(t *testing.T) {
	ctx := NewSharedTestContext()

	ctx.SetIsolation("/original", "/isolated")

	if ctx.OriginalRepoRoot != "/original" {
		t.Error("OriginalRepoRoot not set correctly")
	}
	if ctx.IsolatedTestProjectDir != "/isolated" {
		t.Error("IsolatedTestProjectDir not set correctly")
	}
	if ctx.TestDir != "/isolated" {
		t.Error("TestDir should be set to isolated dir")
	}
}

func TestSharedTestContext_ClearIsolation(t *testing.T) {
	ctx := NewSharedTestContext()
	ctx.SetIsolation("/original", "/isolated")

	ctx.ClearIsolation()

	// Only clear the isolated dir, not the original root
	if ctx.IsolatedTestProjectDir != "" {
		t.Error("IsolatedTestProjectDir should be cleared")
	}
	if ctx.TestDir != "" {
		t.Error("TestDir should be cleared")
	}
}
