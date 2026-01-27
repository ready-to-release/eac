package tool

import (
	"testing"
)

func TestToolDefinition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tool    ToolDefinition
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty tool",
			tool:    ToolDefinition{},
			wantErr: true,
			errMsg:  "tool ID is required",
		},
		{
			name: "system tool without binary",
			tool: ToolDefinition{
				ID:   "test-tool",
				Type: ToolTypeSystem,
			},
			wantErr: true,
			errMsg:  "system tool \"test-tool\" requires binary",
		},
		{
			name: "valid system tool",
			tool: ToolDefinition{
				ID:     "go-build",
				Type:   ToolTypeSystem,
				Binary: "go",
				Args:   []string{"build", "./..."},
			},
			wantErr: false,
		},
		{
			name: "container tool without image",
			tool: ToolDefinition{
				ID:   "trivy",
				Type: ToolTypeContainer,
			},
			wantErr: true,
			errMsg:  "container tool \"trivy\" requires image or container reference",
		},
		{
			name: "valid container tool with image",
			tool: ToolDefinition{
				ID:    "trivy",
				Type:  ToolTypeContainer,
				Image: "ghcr.io/aquasecurity/trivy",
				Tag:   "0.68.2",
			},
			wantErr: false,
		},
		{
			name: "valid container tool with container ref",
			tool: ToolDefinition{
				ID:        "trivy",
				Type:      ToolTypeContainer,
				Container: "trivy",
			},
			wantErr: false,
		},
		{
			name: "tool without type",
			tool: ToolDefinition{
				ID:     "test-tool",
				Binary: "test",
			},
			wantErr: true,
			errMsg:  "tool \"test-tool\" requires type (system or container)",
		},
		{
			name: "invalid mount - missing source",
			tool: ToolDefinition{
				ID:    "test",
				Type:  ToolTypeContainer,
				Image: "test",
				Mounts: []MountConfig{
					{Target: "/app"},
				},
			},
			wantErr: true,
			errMsg:  "tool \"test\" mount[0]: mount source is required",
		},
		{
			name: "invalid mount - missing target",
			tool: ToolDefinition{
				ID:    "test",
				Type:  ToolTypeContainer,
				Image: "test",
				Mounts: []MountConfig{
					{Source: "/workspace"},
				},
			},
			wantErr: true,
			errMsg:  "tool \"test\" mount[0]: mount target is required",
		},
		{
			name: "valid container with mounts",
			tool: ToolDefinition{
				ID:    "test",
				Type:  ToolTypeContainer,
				Image: "test",
				Mounts: []MountConfig{
					{Source: "{workspace}", Target: "/app"},
					{Source: "{output}", Target: "/out", ReadOnly: true},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tool.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("error message mismatch:\n  got:  %s\n  want: %s", err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestToolDefinition_FullImage(t *testing.T) {
	tests := []struct {
		name  string
		tool  ToolDefinition
		want  string
	}{
		{
			name:  "no image",
			tool:  ToolDefinition{},
			want:  "",
		},
		{
			name: "image without tag",
			tool: ToolDefinition{
				Image: "alpine",
			},
			want: "alpine:latest",
		},
		{
			name: "image with tag",
			tool: ToolDefinition{
				Image: "ghcr.io/aquasecurity/trivy",
				Tag:   "0.68.2",
			},
			want: "ghcr.io/aquasecurity/trivy:0.68.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tool.FullImage()
			if got != tt.want {
				t.Errorf("FullImage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToolDefinition_Clone(t *testing.T) {
	original := &ToolDefinition{
		ID:          "test",
		Description: "Test tool",
		Type:        ToolTypeContainer,
		Image:       "test-image",
		Tag:         "v1",
		Args:        []string{"--verbose"},
		Command:     []string{"run"},
		Entrypoint:  []string{"/bin/sh"},
		Env:         map[string]string{"FOO": "bar"},
		Mounts: []MountConfig{
			{Source: "/src", Target: "/app"},
		},
		Requirements: []string{"docker"},
	}

	clone := original.Clone()

	// Modify original to verify deep copy
	original.ID = "modified"
	original.Args[0] = "modified"
	original.Command[0] = "modified"
	original.Env["FOO"] = "modified"
	original.Mounts[0].Source = "modified"
	original.Requirements[0] = "modified"

	// Verify clone wasn't affected
	if clone.ID != "test" {
		t.Error("Clone ID was modified")
	}
	if clone.Args[0] != "--verbose" {
		t.Error("Clone Args was modified")
	}
	if clone.Command[0] != "run" {
		t.Error("Clone Command was modified")
	}
	if clone.Env["FOO"] != "bar" {
		t.Error("Clone Env was modified")
	}
	if clone.Mounts[0].Source != "/src" {
		t.Error("Clone Mounts was modified")
	}
	if clone.Requirements[0] != "docker" {
		t.Error("Clone Requirements was modified")
	}
}

func TestMountConfig_ResolvePlaceholders(t *testing.T) {
	mount := MountConfig{
		Source:   "{workspace}/src",
		Target:   "/app/{module}",
		ReadOnly: true,
	}

	placeholders := map[string]string{
		"{workspace}": "/home/user/project",
		"{module}":    "mymodule",
	}

	resolved := mount.ResolvePlaceholders(placeholders)

	if resolved.Source != "/home/user/project/src" {
		t.Errorf("Source = %q, want %q", resolved.Source, "/home/user/project/src")
	}
	if resolved.Target != "/app/mymodule" {
		t.Errorf("Target = %q, want %q", resolved.Target, "/app/mymodule")
	}
	if !resolved.ReadOnly {
		t.Error("ReadOnly should be true")
	}

	// Verify original wasn't modified
	if mount.Source != "{workspace}/src" {
		t.Error("Original mount was modified")
	}
}

func TestToolAssignment_GetToolID(t *testing.T) {
	assignment := &ToolAssignment{
		Builder: "go-system",
		Linter:  "golangci-lint",
		Scanner: "trivy-vuln",
		Tester:  "go-test",
		Server:  "docs-serve",
	}

	tests := []struct {
		op   OperationType
		want string
	}{
		{OperationBuild, "go-system"},
		{OperationLint, "golangci-lint"},
		{OperationScan, "trivy-vuln"},
		{OperationTest, "go-test"},
		{OperationServe, "docs-serve"},
		{OperationType("unknown"), ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.op), func(t *testing.T) {
			got := assignment.GetToolID(tt.op)
			if got != tt.want {
				t.Errorf("GetToolID(%s) = %q, want %q", tt.op, got, tt.want)
			}
		})
	}
}

func TestToolAssignment_GetToolIDs_MultipleTools(t *testing.T) {
	assignment := &ToolAssignment{
		Linter:   "primary-linter",
		Linters:  []string{"linter-1", "linter-2", "linter-3"},
		Scanner:  "primary-scanner",
		Scanners: []string{"scanner-1", "scanner-2"},
	}

	// Multiple linters should take precedence
	linters := assignment.GetToolIDs(OperationLint)
	if len(linters) != 3 {
		t.Errorf("expected 3 linters, got %d", len(linters))
	}

	// Multiple scanners should take precedence
	scanners := assignment.GetToolIDs(OperationScan)
	if len(scanners) != 2 {
		t.Errorf("expected 2 scanners, got %d", len(scanners))
	}
}

func TestToolAssignment_GetToolIDs_SingleTool(t *testing.T) {
	assignment := &ToolAssignment{
		Linter: "single-linter",
	}

	linters := assignment.GetToolIDs(OperationLint)
	if len(linters) != 1 || linters[0] != "single-linter" {
		t.Errorf("expected single-linter, got %v", linters)
	}
}

func TestToolConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &ToolConfig{
			Tools: map[string]*ToolDefinition{
				"go-system": {
					ID:     "go-system",
					Type:   ToolTypeSystem,
					Binary: "go",
				},
			},
			ComponentTools: map[string]*ToolAssignment{
				"go": {Builder: "go-system"},
			},
		}

		errs := config.Validate()
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
	})

	t.Run("invalid tool reference", func(t *testing.T) {
		config := &ToolConfig{
			Tools: map[string]*ToolDefinition{
				"go-system": {
					ID:     "go-system",
					Type:   ToolTypeSystem,
					Binary: "go",
				},
			},
			ComponentTools: map[string]*ToolAssignment{
				"go": {Builder: "nonexistent-tool"},
			},
		}

		errs := config.Validate()
		if len(errs) != 1 {
			t.Errorf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("invalid environment tool reference", func(t *testing.T) {
		config := &ToolConfig{
			Tools: map[string]*ToolDefinition{
				"go-system": {
					ID:     "go-system",
					Type:   ToolTypeSystem,
					Binary: "go",
				},
			},
			Environments: map[string]*EnvironmentConfig{
				"ci": {
					ComponentTools: map[string]*ToolAssignment{
						"go": {Builder: "nonexistent"},
					},
				},
			},
		}

		errs := config.Validate()
		if len(errs) != 1 {
			t.Errorf("expected 1 error, got %d", len(errs))
		}
	})
}

func TestExecutionResult_Success(t *testing.T) {
	success := &ExecutionResult{ExitCode: 0}
	if !success.Success() {
		t.Error("ExitCode 0 should be success")
	}

	failure := &ExecutionResult{ExitCode: 1}
	if failure.Success() {
		t.Error("ExitCode 1 should not be success")
	}
}

func TestExecutionResult_Output(t *testing.T) {
	t.Run("prefers stdout", func(t *testing.T) {
		result := &ExecutionResult{
			Stdout: []byte("stdout content"),
			Stderr: []byte("stderr content"),
		}
		if string(result.Output()) != "stdout content" {
			t.Error("should prefer stdout")
		}
	})

	t.Run("fallback to stderr", func(t *testing.T) {
		result := &ExecutionResult{
			Stderr: []byte("stderr content"),
		}
		if string(result.Output()) != "stderr content" {
			t.Error("should fallback to stderr")
		}
	})
}

func TestAllOperations(t *testing.T) {
	ops := AllOperations()
	expected := []OperationType{
		OperationBuild,
		OperationLint,
		OperationScan,
		OperationTest,
		OperationServe,
	}

	if len(ops) != len(expected) {
		t.Errorf("expected %d operations, got %d", len(expected), len(ops))
	}

	for i, op := range expected {
		if ops[i] != op {
			t.Errorf("operation[%d] = %s, want %s", i, ops[i], op)
		}
	}
}
