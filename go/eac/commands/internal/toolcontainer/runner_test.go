package toolcontainer

import (
	"context"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"gopkg.in/yaml.v3"
)

func TestContainerToolConfig_FullImage(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.ContainerToolConfig
		expected string
	}{
		{
			name:     "nil config",
			cfg:      nil,
			expected: "",
		},
		{
			name:     "empty image",
			cfg:      &config.ContainerToolConfig{},
			expected: "",
		},
		{
			name: "image with default tag",
			cfg: &config.ContainerToolConfig{
				Image: "ghcr.io/ready-to-release/drawio-cli",
			},
			expected: "ghcr.io/ready-to-release/drawio-cli:latest",
		},
		{
			name: "image with custom tag",
			cfg: &config.ContainerToolConfig{
				Image: "ghcr.io/ready-to-release/drawio-cli",
				Tag:   "v1.0.0",
			},
			expected: "ghcr.io/ready-to-release/drawio-cli:v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.FullImage()
			if result != tt.expected {
				t.Errorf("FullImage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestContainerToolConfig_GetWorkdir(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.ContainerToolConfig
		expected string
	}{
		{
			name:     "nil config",
			cfg:      nil,
			expected: "/workspace",
		},
		{
			name:     "empty workdir",
			cfg:      &config.ContainerToolConfig{},
			expected: "/workspace",
		},
		{
			name: "custom workdir",
			cfg: &config.ContainerToolConfig{
				Workdir: "/docs",
			},
			expected: "/docs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.GetWorkdir()
			if result != tt.expected {
				t.Errorf("GetWorkdir() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestContainerToolConfig_GetLocalBinding(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name     string
		cfg      *config.ContainerToolConfig
		expected bool
	}{
		{
			name:     "nil config",
			cfg:      nil,
			expected: true,
		},
		{
			name:     "nil local_binding (default true)",
			cfg:      &config.ContainerToolConfig{},
			expected: true,
		},
		{
			name: "explicit true",
			cfg: &config.ContainerToolConfig{
				LocalBinding: &trueVal,
			},
			expected: true,
		},
		{
			name: "explicit false",
			cfg: &config.ContainerToolConfig{
				LocalBinding: &falseVal,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.GetLocalBinding()
			if result != tt.expected {
				t.Errorf("GetLocalBinding() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainerToolConfig_ShouldBuildLocally(t *testing.T) {
	falseVal := false

	tests := []struct {
		name            string
		cfg             *config.ContainerToolConfig
		containerMode   bool
		expected        bool
	}{
		{
			name:          "nil config",
			cfg:           nil,
			containerMode: false,
			expected:      false,
		},
		{
			name: "no dockerfile",
			cfg: &config.ContainerToolConfig{
				Image: "some/image",
			},
			containerMode: false,
			expected:      false,
		},
		{
			name: "has dockerfile, local mode, default binding",
			cfg: &config.ContainerToolConfig{
				Dockerfile: "containers/test/Dockerfile",
			},
			containerMode: false,
			expected:      true,
		},
		{
			name: "has dockerfile, local mode, binding false",
			cfg: &config.ContainerToolConfig{
				Dockerfile:   "containers/test/Dockerfile",
				LocalBinding: &falseVal,
			},
			containerMode: false,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Can't test container mode here without modifying env vars
			// which would affect parallel tests
			if tt.containerMode {
				t.Skip("Container mode test requires env var modification")
			}

			result := tt.cfg.ShouldBuildLocally()
			if result != tt.expected {
				t.Errorf("ShouldBuildLocally() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMockRunner(t *testing.T) {
	t.Run("default success", func(t *testing.T) {
		mock := NewMockRunner(ModeLocal)

		result := mock.Run(context.Background(), &RunConfig{
			Command: "test",
			Args:    []string{"arg1", "arg2"},
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		if result.Error != nil {
			t.Errorf("expected no error, got %v", result.Error)
		}
		if mock.Mode() != ModeLocal {
			t.Errorf("expected mode %s, got %s", ModeLocal, mock.Mode())
		}
	})

	t.Run("custom run function", func(t *testing.T) {
		mock := NewMockRunner(ModeContainer).WithRunFunc(func(ctx context.Context, cfg *RunConfig) *RunResult {
			return &RunResult{ExitCode: 42}
		})

		result := mock.Run(context.Background(), &RunConfig{})

		if result.ExitCode != 42 {
			t.Errorf("expected exit code 42, got %d", result.ExitCode)
		}
	})

	t.Run("run history", func(t *testing.T) {
		mock := NewMockRunner(ModeLocal)

		mock.Run(context.Background(), &RunConfig{Command: "first"})
		mock.Run(context.Background(), &RunConfig{Command: "second"})
		mock.Run(context.Background(), &RunConfig{Command: "third"})

		history := mock.RunHistory()
		if len(history) != 3 {
			t.Errorf("expected 3 runs in history, got %d", len(history))
		}

		last := mock.LastRun()
		if last == nil || last.Command != "third" {
			t.Errorf("expected last run command to be 'third'")
		}
	})
}

func TestContainersConfig_UnmarshalYAML(t *testing.T) {
	yamlInput := `
base_images:
  python: "3.12"
  node: "25"
drawio-cli:
  dockerfile: containers/drawio-cli/Dockerfile
  image: ghcr.io/ready-to-release/drawio-cli
  workdir: /docs
mermaid-cli:
  dockerfile: containers/mermaid-cli/Dockerfile
  local_binding: false
`

	// Test parsing
	var cfg config.ContainersConfig
	if err := cfg.UnmarshalYAML(mustParseYAMLNode(t, yamlInput)); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	// Verify base_images
	if cfg.BaseImages["python"] != "3.12" {
		t.Errorf("expected python=3.12, got %s", cfg.BaseImages["python"])
	}
	if cfg.BaseImages["node"] != "25" {
		t.Errorf("expected node=25, got %s", cfg.BaseImages["node"])
	}

	// Verify tools
	if len(cfg.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(cfg.Tools))
	}

	drawio := cfg.Tools["drawio-cli"]
	if drawio == nil {
		t.Fatal("expected drawio-cli tool")
	}
	if drawio.Dockerfile != "containers/drawio-cli/Dockerfile" {
		t.Errorf("unexpected dockerfile: %s", drawio.Dockerfile)
	}
	if drawio.Workdir != "/docs" {
		t.Errorf("unexpected workdir: %s", drawio.Workdir)
	}

	mermaid := cfg.Tools["mermaid-cli"]
	if mermaid == nil {
		t.Fatal("expected mermaid-cli tool")
	}
	if mermaid.GetLocalBinding() {
		t.Error("expected local_binding=false")
	}
}

func mustParseYAMLNode(t *testing.T, input string) *yaml.Node {
	t.Helper()

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(input), &node); err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}

	// The root node is a document, we want the content
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return node.Content[0]
	}
	return &node
}
