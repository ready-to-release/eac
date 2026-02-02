package tool

import (
	"testing"
)

func TestNewHandlerToolBridge(t *testing.T) {
	bridge := NewHandlerToolBridge()
	if bridge == nil {
		t.Fatal("NewHandlerToolBridge() returned nil")
	}
}

func TestToolContext_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ctx     *ToolContext
		wantErr bool
	}{
		{
			name: "valid context",
			ctx: &ToolContext{
				WorkspaceRoot: "/workspace",
				OutputDir:     "/output",
			},
			wantErr: false,
		},
		{
			name: "missing workspace root",
			ctx: &ToolContext{
				OutputDir: "/output",
			},
			wantErr: true,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ctx.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubstitutePlaceholders(t *testing.T) {
	tests := []struct {
		name string
		s    string
		ctx  *ToolContext
		want string
	}{
		{
			name: "workspace placeholder",
			s:    "{workspace}/docs",
			ctx: &ToolContext{
				WorkspaceRoot: "/repo",
			},
			want: "/repo/docs",
		},
		{
			name: "multiple placeholders",
			s:    "{workspace}/staging/{output}",
			ctx: &ToolContext{
				WorkspaceRoot: "/repo",
				OutputDir:     "out",
			},
			want: "/repo/staging/out",
		},
		{
			name: "staging placeholder",
			s:    "{staging}/content",
			ctx: &ToolContext{
				StagingDir: "/cache/staging",
			},
			want: "/cache/staging/content",
		},
		{
			name: "config placeholder",
			s:    "-f {config}",
			ctx: &ToolContext{
				ConfigPath: "/output/mkdocs.yml",
			},
			want: "-f /output/mkdocs.yml",
		},
		{
			name: "custom variable",
			s:    "--theme {theme}",
			ctx: &ToolContext{
				Variables: map[string]string{
					"theme": "dark",
				},
			},
			want: "--theme dark",
		},
		{
			name: "no placeholders",
			s:    "mkdocs build",
			ctx: &ToolContext{
				WorkspaceRoot: "/repo",
			},
			want: "mkdocs build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := substitutePlaceholders(tt.s, tt.ctx)
			if got != tt.want {
				t.Errorf("substitutePlaceholders(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestBuildContainerConfig(t *testing.T) {
	bridge := NewHandlerToolBridge()

	tool := &ToolDefinition{
		ID:        "site-render-tool",
		Type:      ToolTypeContainer,
		LocalPath: "containers/site-render-tool",
		Command:   []string{"mkdocs", "build", "-f", "{config}", "--site-dir", "{output}/site"},
		WorkDir:   "/docs",
		Mounts: []MountConfig{
			{Source: "{workspace}", Target: "/docs"},
			{Source: "{staging}", Target: "/staging", ReadOnly: true},
			{Source: "{output}", Target: "/output"},
		},
		Resources: &ToolResources{
			CPUs:   4,
			Memory: "4g",
		},
	}

	tc := &ToolContext{
		WorkspaceRoot: "/repo",
		StagingDir:    "/cache/staging",
		OutputDir:     "/out",
		ConfigPath:    "/out/mkdocs.yml",
	}

	config := bridge.buildContainerConfig(tool, tc)

	// Verify image
	if config.Image != "site-render-tool:local" {
		t.Errorf("Image = %q, want %q", config.Image, "site-render-tool:local")
	}

	// Verify command substitution
	expectedCmd := []string{"mkdocs", "build", "-f", "/out/mkdocs.yml", "--site-dir", "/out/site"}
	if len(config.Command) != len(expectedCmd) {
		t.Fatalf("Command length = %d, want %d", len(config.Command), len(expectedCmd))
	}
	for i, cmd := range config.Command {
		if cmd != expectedCmd[i] {
			t.Errorf("Command[%d] = %q, want %q", i, cmd, expectedCmd[i])
		}
	}

	// Verify mounts
	if len(config.Mounts) != 3 {
		t.Fatalf("Mounts length = %d, want 3", len(config.Mounts))
	}
	if config.Mounts[0].Source != "/repo" {
		t.Errorf("Mount[0].Source = %q, want %q", config.Mounts[0].Source, "/repo")
	}
	if config.Mounts[1].Source != "/cache/staging" {
		t.Errorf("Mount[1].Source = %q, want %q", config.Mounts[1].Source, "/cache/staging")
	}
	if !config.Mounts[1].ReadOnly {
		t.Error("Mount[1] should be ReadOnly")
	}

	// Verify resources
	if config.Resources == nil {
		t.Fatal("Resources should not be nil")
	}
	if config.Resources.Memory != "4g" {
		t.Errorf("Memory = %q, want %q", config.Resources.Memory, "4g")
	}
	if config.Resources.CPUs != 4 {
		t.Errorf("CPUs = %f, want 4", config.Resources.CPUs)
	}
}

func TestBuildContainerConfig_ExternalImage(t *testing.T) {
	bridge := NewHandlerToolBridge()

	tool := &ToolDefinition{
		ID:    "external-tool",
		Type:  ToolTypeContainer,
		Image: "ghcr.io/org/image",
		Tag:   "v1.2.3",
	}

	tc := &ToolContext{
		WorkspaceRoot: "/repo",
	}

	config := bridge.buildContainerConfig(tool, tc)

	if config.Image != "ghcr.io/org/image:v1.2.3" {
		t.Errorf("Image = %q, want %q", config.Image, "ghcr.io/org/image:v1.2.3")
	}
}

func TestBuildContainerConfig_Environment(t *testing.T) {
	bridge := NewHandlerToolBridge()

	tool := &ToolDefinition{
		ID:   "env-tool",
		Type: ToolTypeContainer,
		Env: map[string]string{
			"ENABLE_PDF_EXPORT": "true",
			"THEME":             "dark",
		},
		LocalPath: "containers/test",
	}

	tc := &ToolContext{
		WorkspaceRoot: "/repo",
	}

	config := bridge.buildContainerConfig(tool, tc)

	// Check that both env vars are present
	if config.Env["ENABLE_PDF_EXPORT"] != "true" {
		t.Error("Missing or incorrect ENABLE_PDF_EXPORT env var")
	}
	if config.Env["THEME"] != "dark" {
		t.Error("Missing or incorrect THEME env var")
	}
}

func TestBuildContainerConfig_WithWeight(t *testing.T) {
	bridge := NewHandlerToolBridge()

	tool := &ToolDefinition{
		ID:        "pdf-tool",
		Type:      ToolTypeContainer,
		LocalPath: "containers/pdf-render-tool",
		Resources: &ToolResources{
			CPUs:    1,      // Base CPU
			Memory:  "2g",   // Base memory
			ShmSize: "512m", // Base shm
		},
	}

	// Test with weight=4 (like pdf-render)
	tc := &ToolContext{
		WorkspaceRoot: "/repo",
		Weight:        4,
	}

	config := bridge.buildContainerConfig(tool, tc)

	if config.Resources == nil {
		t.Fatal("Resources should not be nil")
	}

	// CPUs should be scaled: 1 * 4 = 4
	if config.Resources.CPUs != 4 {
		t.Errorf("CPUs = %f, want 4 (1 * weight=4)", config.Resources.CPUs)
	}

	// Memory should be scaled: 2g * 4 = 8g
	if config.Resources.Memory != "8g" {
		t.Errorf("Memory = %q, want %q (2g * weight=4)", config.Resources.Memory, "8g")
	}

	// ShmSize should be scaled: 512m * 4 = 2048m
	if config.Resources.ShmSize != "2048m" {
		t.Errorf("ShmSize = %q, want %q (512m * weight=4)", config.Resources.ShmSize, "2048m")
	}
}

func TestBuildContainerConfig_NoWeight(t *testing.T) {
	bridge := NewHandlerToolBridge()

	tool := &ToolDefinition{
		ID:        "simple-tool",
		Type:      ToolTypeContainer,
		LocalPath: "containers/simple",
		Resources: &ToolResources{
			CPUs:   2,
			Memory: "4g",
		},
	}

	// Test with no weight (or weight=0)
	tc := &ToolContext{
		WorkspaceRoot: "/repo",
		Weight:        0, // Should default to 1
	}

	config := bridge.buildContainerConfig(tool, tc)

	if config.Resources == nil {
		t.Fatal("Resources should not be nil")
	}

	// CPUs should remain 2 (no scaling when weight <= 1)
	if config.Resources.CPUs != 2 {
		t.Errorf("CPUs = %f, want 2 (no scaling)", config.Resources.CPUs)
	}

	// Memory should remain 4g (no scaling when weight <= 1)
	if config.Resources.Memory != "4g" {
		t.Errorf("Memory = %q, want %q (no scaling)", config.Resources.Memory, "4g")
	}
}

func TestBuildContainerConfig_ExplicitCPULimit(t *testing.T) {
	bridge := NewHandlerToolBridge()

	// Test case: tool with explicit CPU limit (like pdf-render-tool)
	// Tool has cpus: 4, component has weight: 4
	// Should NOT multiply: 4 CPUs used directly (not 4 * 4 = 16)
	tool := &ToolDefinition{
		ID:        "pdf-render-tool",
		Type:      ToolTypeContainer,
		LocalPath: "containers/pdf-render-tool",
		Resources: &ToolResources{
			CPUs:   4,    // Explicit limit
			Memory: "8g", // Will be scaled by weight
		},
	}

	tc := &ToolContext{
		WorkspaceRoot: "/repo",
		Weight:        4, // Component weight from pdf-render type
	}

	config := bridge.buildContainerConfig(tool, tc)

	if config.Resources == nil {
		t.Fatal("Resources should not be nil")
	}

	// CPUs should be 4 (explicit limit, not scaled)
	if config.Resources.CPUs != 4 {
		t.Errorf("CPUs = %f, want 4 (explicit limit, not multiplied by weight)", config.Resources.CPUs)
	}

	// Memory should NOT be scaled (>= 4g is explicit limit): 8g stays 8g
	if config.Resources.Memory != "8g" {
		t.Errorf("Memory = %q, want %q (explicit limit >= 4g, not scaled)", config.Resources.Memory, "8g")
	}
}

func TestShouldScaleMemory(t *testing.T) {
	tests := []struct {
		name string
		mem  string
		want bool
	}{
		{
			name: "base unit - 2g",
			mem:  "2g",
			want: true,
		},
		{
			name: "base unit - 1g",
			mem:  "1g",
			want: true,
		},
		{
			name: "base unit - 512m",
			mem:  "512m",
			want: true,
		},
		{
			name: "explicit limit - 4g",
			mem:  "4g",
			want: false, // >= 4g is explicit
		},
		{
			name: "explicit limit - 8g",
			mem:  "8g",
			want: false, // >= 4g is explicit
		},
		{
			name: "explicit limit - 16g",
			mem:  "16g",
			want: false,
		},
		{
			name: "empty",
			mem:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldScaleMemory(tt.mem)
			if got != tt.want {
				t.Errorf("shouldScaleMemory(%q) = %v, want %v", tt.mem, got, tt.want)
			}
		})
	}
}

func TestParseMemoryToBytes(t *testing.T) {
	tests := []struct {
		name string
		mem  string
		want int64
	}{
		{
			name: "gigabytes",
			mem:  "8g",
			want: 8 * 1024 * 1024 * 1024,
		},
		{
			name: "megabytes",
			mem:  "512m",
			want: 512 * 1024 * 1024,
		},
		{
			name: "4gb boundary",
			mem:  "4g",
			want: 4 * 1024 * 1024 * 1024,
		},
		{
			name: "just under 4gb",
			mem:  "3999m",
			want: 3999 * 1024 * 1024,
		},
		{
			name: "empty",
			mem:  "",
			want: -1,
		},
		{
			name: "invalid",
			mem:  "invalid",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMemoryToBytes(tt.mem)
			if got != tt.want {
				t.Errorf("parseMemoryToBytes(%q) = %d, want %d", tt.mem, got, tt.want)
			}
		})
	}
}

func TestScaleMemory(t *testing.T) {
	tests := []struct {
		name   string
		mem    string
		weight int
		want   string
	}{
		{
			name:   "scale gigabytes",
			mem:    "2g",
			weight: 4,
			want:   "8g",
		},
		{
			name:   "scale megabytes",
			mem:    "512m",
			weight: 4,
			want:   "2048m",
		},
		{
			name:   "no scaling when weight 1",
			mem:    "4g",
			weight: 1,
			want:   "4g",
		},
		{
			name:   "no scaling when weight 0",
			mem:    "4g",
			weight: 0,
			want:   "4g",
		},
		{
			name:   "empty memory",
			mem:    "",
			weight: 4,
			want:   "",
		},
		{
			name:   "scale bytes",
			mem:    "1024",
			weight: 2,
			want:   "2048",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scaleMemory(tt.mem, tt.weight)
			if got != tt.want {
				t.Errorf("scaleMemory(%q, %d) = %q, want %q", tt.mem, tt.weight, got, tt.want)
			}
		})
	}
}
