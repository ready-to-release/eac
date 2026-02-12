//go:build L1 && ov
// +build L1,ov

package design

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// HashDSLContent
// ---------------------------------------------------------------------------

func TestHashDSLContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "empty content",
			content: "",
		},
		{
			name:    "simple content",
			content: "workspace { }",
		},
		{
			name:    "multiline content",
			content: "workspace {\n  model {\n  }\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashDSLContent(tt.content)

			// Hash must be exactly 8 hex characters
			if len(got) != 8 {
				t.Errorf("HashDSLContent() length = %d, want 8", len(got))
			}

			// Verify it only contains hex characters
			for _, c := range got {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("HashDSLContent() contains non-hex character %q in %q", string(c), got)
					break
				}
			}
		})
	}
}

func TestHashDSLContent_Deterministic(t *testing.T) {
	content := "workspace {\n  model {\n    person user \"A user\"\n  }\n}"

	hash1 := HashDSLContent(content)
	hash2 := HashDSLContent(content)

	if hash1 != hash2 {
		t.Errorf("HashDSLContent() not deterministic: %q != %q", hash1, hash2)
	}
}

func TestHashDSLContent_DifferentContent(t *testing.T) {
	hash1 := HashDSLContent("workspace { model { } }")
	hash2 := HashDSLContent("workspace { views { } }")

	if hash1 == hash2 {
		t.Errorf("HashDSLContent() same hash for different content: %q", hash1)
	}
}

func TestHashDSLContent_LineEndingNormalization(t *testing.T) {
	tests := []struct {
		name    string
		contentA string
		contentB string
	}{
		{
			name:     "LF vs CRLF",
			contentA: "line1\nline2\nline3",
			contentB: "line1\r\nline2\r\nline3",
		},
		{
			name:     "LF vs CR",
			contentA: "line1\nline2\nline3",
			contentB: "line1\rline2\rline3",
		},
		{
			name:     "CRLF vs CR",
			contentA: "line1\r\nline2\r\nline3",
			contentB: "line1\rline2\rline3",
		},
		{
			name:     "mixed line endings",
			contentA: "line1\nline2\nline3\n",
			contentB: "line1\r\nline2\rline3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashA := HashDSLContent(tt.contentA)
			hashB := HashDSLContent(tt.contentB)

			if hashA != hashB {
				t.Errorf("HashDSLContent() not equal for normalized line endings: %q != %q", hashA, hashB)
			}
		})
	}
}

func TestHashDSLContent_MatchesSHA256Prefix(t *testing.T) {
	content := "workspace { model { } }"
	// Manually compute expected hash with same normalization logic
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	h := sha256.Sum256([]byte(normalized))
	expected := fmt.Sprintf("%x", h)[:8]

	got := HashDSLContent(content)

	if got != expected {
		t.Errorf("HashDSLContent() = %q, want %q", got, expected)
	}
}

// ---------------------------------------------------------------------------
// extractViewKey
// ---------------------------------------------------------------------------

func TestExtractViewKey(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "simple view key",
			filename: "structurizr-SystemContext.svg",
			want:     "SystemContext",
		},
		{
			name:     "view key with number suffix",
			filename: "structurizr-SystemContext-1.svg",
			want:     "SystemContext",
		},
		{
			name:     "view key with large number suffix",
			filename: "structurizr-Containers-42.svg",
			want:     "Containers",
		},
		{
			name:     "container view",
			filename: "structurizr-ContainerView.svg",
			want:     "ContainerView",
		},
		{
			name:     "deployment view",
			filename: "structurizr-Deployment.svg",
			want:     "Deployment",
		},
		{
			name:     "dynamic view",
			filename: "structurizr-DynamicView.svg",
			want:     "DynamicView",
		},
		{
			name:     "view key with dashes",
			filename: "structurizr-my-custom-view.svg",
			want:     "my-custom-view",
		},
		{
			name:     "view key with dashes and number suffix",
			filename: "structurizr-my-custom-view-1.svg",
			want:     "my-custom-view",
		},
		{
			name:     "no structurizr prefix falls back to trim",
			filename: "custom-output.svg",
			want:     "custom-output",
		},
		{
			name:     "non-svg file returns empty from regex, fallback removes extension",
			filename: "structurizr-Test.png",
			want:     "Test.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractViewKey(tt.filename)
			if got != tt.want {
				t.Errorf("extractViewKey(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestExtractViewKey_FallbackBehavior(t *testing.T) {
	// When the regex doesn't match, the fallback strips ".svg" and "structurizr-" prefix
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "no prefix no extension",
			filename: "SomeFile.svg",
			want:     "SomeFile",
		},
		{
			name:     "structurizr prefix different extension",
			filename: "structurizr-View.xml",
			want:     "View.xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractViewKey(tt.filename)
			if got != tt.want {
				t.Errorf("extractViewKey(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseViewKeysFromDSL
// ---------------------------------------------------------------------------

func TestParseViewKeysFromDSL(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "empty content",
			content: "",
			want:    []string{},
		},
		{
			name: "single systemContext view",
			content: `workspace {
    views {
        systemContext softwareSystem "SystemContextView"
    }
}`,
			want: []string{"SystemContextView"},
		},
		{
			name: "single container view",
			content: `workspace {
    views {
        container softwareSystem "ContainerView"
    }
}`,
			want: []string{"ContainerView"},
		},
		{
			name: "single component view",
			content: `workspace {
    views {
        component container1 "ComponentView"
    }
}`,
			want: []string{"ComponentView"},
		},
		{
			name: "dynamic view",
			content: `workspace {
    views {
        dynamic softwareSystem "DynamicView"
    }
}`,
			want: []string{"DynamicView"},
		},
		{
			name: "deployment view captures first quoted string",
			content: `workspace {
    views {
        deployment softwareSystem "Live" "DeploymentView"
    }
}`,
			// The regex captures the first quoted string after type + identifier,
			// so for deployment views with environment + key, it captures the environment name.
			want: []string{"Live"},
		},
		{
			name: "deployment view single quoted arg",
			content: `workspace {
    views {
        deployment softwareSystem "DeploymentView"
    }
}`,
			want: []string{"DeploymentView"},
		},
		{
			name: "custom view",
			content: `workspace {
    views {
        custom element "MyCustomView"
    }
}`,
			want: []string{"MyCustomView"},
		},
		{
			name: "filtered view",
			content: `workspace {
    views {
        filtered baseView "FilteredView"
    }
}`,
			want: []string{"FilteredView"},
		},
		{
			name: "multiple views",
			content: `workspace {
    views {
        systemContext system "ContextView"
        container system "ContainerView"
        component backend "ComponentView"
        dynamic system "DynamicFlow"
    }
}`,
			want: []string{"ContextView", "ContainerView", "ComponentView", "DynamicFlow"},
		},
		{
			name:    "no views section",
			content: "workspace {\n  model {\n    person user\n  }\n}",
			want:    []string{},
		},
		{
			name: "view with spaces in key",
			content: `workspace {
    views {
        systemContext system "System Context Overview"
    }
}`,
			want: []string{"System Context Overview"},
		},
		{
			name: "indented views",
			content: `workspace {
    views {
        systemContext system "View1"
            container system "View2"
    }
}`,
			want: []string{"View1", "View2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseViewKeysFromDSL(tt.content)

			if len(got) != len(tt.want) {
				t.Fatalf("ParseViewKeysFromDSL() returned %d keys, want %d\ngot:  %v\nwant: %v",
					len(got), len(tt.want), got, tt.want)
			}

			for i, key := range got {
				if key != tt.want[i] {
					t.Errorf("ParseViewKeysFromDSL()[%d] = %q, want %q", i, key, tt.want[i])
				}
			}
		})
	}
}

func TestParseViewKeysFromDSL_IgnoresComments(t *testing.T) {
	// Lines that look like view definitions but are inside comments or
	// don't match the exact pattern should not be picked up.
	content := `workspace {
    // systemContext system "CommentedOut"
    model {
        person user "A User"
    }
}`

	got := ParseViewKeysFromDSL(content)
	if len(got) != 0 {
		t.Errorf("ParseViewKeysFromDSL() picked up commented view, got: %v", got)
	}
}

func TestParseViewKeysFromDSL_UnknownViewType(t *testing.T) {
	// Only known view types should match
	content := `workspace {
    views {
        landscape system "LandscapeView"
    }
}`

	got := ParseViewKeysFromDSL(content)
	if len(got) != 0 {
		t.Errorf("ParseViewKeysFromDSL() matched unknown view type, got: %v", got)
	}
}

// ---------------------------------------------------------------------------
// NewExporter / NewExporterWithOutput
// ---------------------------------------------------------------------------

func TestNewExporter(t *testing.T) {
	exporter, err := NewExporter()
	if err != nil {
		t.Fatalf("NewExporter() returned error: %v", err)
	}
	if exporter == nil {
		t.Fatal("NewExporter() returned nil")
	}

	impl, ok := exporter.(*StructurizrExporterImpl)
	if !ok {
		t.Fatal("NewExporter() did not return *StructurizrExporterImpl")
	}

	if impl.OutputDir != "" {
		t.Errorf("NewExporter().OutputDir = %q, want empty string", impl.OutputDir)
	}
}

func TestNewExporterWithOutput(t *testing.T) {
	tests := []struct {
		name      string
		outputDir string
	}{
		{
			name:      "custom directory",
			outputDir: "/tmp/structurizr-export",
		},
		{
			name:      "empty directory",
			outputDir: "",
		},
		{
			name:      "relative directory",
			outputDir: "out/diagrams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter, err := NewExporterWithOutput(tt.outputDir)
			if err != nil {
				t.Fatalf("NewExporterWithOutput(%q) returned error: %v", tt.outputDir, err)
			}
			if exporter == nil {
				t.Fatal("NewExporterWithOutput() returned nil")
			}

			impl, ok := exporter.(*StructurizrExporterImpl)
			if !ok {
				t.Fatal("NewExporterWithOutput() did not return *StructurizrExporterImpl")
			}

			if impl.OutputDir != tt.outputDir {
				t.Errorf("NewExporterWithOutput().OutputDir = %q, want %q", impl.OutputDir, tt.outputDir)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetPlantUMLImage
// ---------------------------------------------------------------------------

func TestGetPlantUMLImage(t *testing.T) {
	image := GetPlantUMLImage()

	if image == "" {
		t.Error("GetPlantUMLImage() returned empty string")
	}

	// The default should contain "plantuml"
	if !strings.Contains(image, "plantuml") {
		t.Errorf("GetPlantUMLImage() = %q, expected to contain 'plantuml'", image)
	}
}

func TestDefaultPlantUMLImage_Constant(t *testing.T) {
	if DefaultPlantUMLImage == "" {
		t.Error("DefaultPlantUMLImage constant is empty")
	}

	if !strings.Contains(DefaultPlantUMLImage, "plantuml") {
		t.Errorf("DefaultPlantUMLImage = %q, expected to contain 'plantuml'", DefaultPlantUMLImage)
	}

	if !strings.Contains(DefaultPlantUMLImage, ":") {
		t.Errorf("DefaultPlantUMLImage = %q, expected to contain ':' (image:tag format)", DefaultPlantUMLImage)
	}
}

// ---------------------------------------------------------------------------
// ExportedView struct
// ---------------------------------------------------------------------------

func TestExportedView_Fields(t *testing.T) {
	view := ExportedView{
		ViewKey: "SystemContext",
		SVGPath: "/tmp/export/mymodule_SystemContext_abc12345.svg",
		Module:  "mymodule",
		DSLHash: "abc12345",
	}

	if view.ViewKey != "SystemContext" {
		t.Errorf("ViewKey = %q, want %q", view.ViewKey, "SystemContext")
	}
	if view.SVGPath != "/tmp/export/mymodule_SystemContext_abc12345.svg" {
		t.Errorf("SVGPath = %q, want expected path", view.SVGPath)
	}
	if view.Module != "mymodule" {
		t.Errorf("Module = %q, want %q", view.Module, "mymodule")
	}
	if view.DSLHash != "abc12345" {
		t.Errorf("DSLHash = %q, want %q", view.DSLHash, "abc12345")
	}
}

// ---------------------------------------------------------------------------
// ExportResult struct
// ---------------------------------------------------------------------------

func TestExportResult_WithError(t *testing.T) {
	result := ExportResult{
		Module:        "failing-module",
		WorkspacePath: "/path/to/workspace.dsl",
		DSLHash:       "deadbeef",
		Views:         nil,
		ExportDir:     "/tmp/output",
		ExecutionTime: 100 * time.Millisecond,
		Error:         fmt.Errorf("docker not running"),
	}

	if result.Error == nil {
		t.Error("ExportResult.Error should not be nil")
	}
	if !strings.Contains(result.Error.Error(), "docker not running") {
		t.Errorf("ExportResult.Error = %q, want to contain 'docker not running'", result.Error.Error())
	}
}

func TestExportResult_Success(t *testing.T) {
	result := ExportResult{
		Module:        "mymodule",
		WorkspacePath: "/path/to/workspace.dsl",
		DSLHash:       "abc12345",
		Views: []ExportedView{
			{ViewKey: "SystemContext", SVGPath: "/out/mymodule_SystemContext_abc12345.svg", Module: "mymodule", DSLHash: "abc12345"},
			{ViewKey: "Containers", SVGPath: "/out/mymodule_Containers_abc12345.svg", Module: "mymodule", DSLHash: "abc12345"},
		},
		ExportDir:     "/out",
		ExecutionTime: 500 * time.Millisecond,
		Error:         nil,
	}

	if result.Error != nil {
		t.Errorf("ExportResult.Error = %v, want nil", result.Error)
	}
	if len(result.Views) != 2 {
		t.Errorf("ExportResult.Views has %d items, want 2", len(result.Views))
	}
}

// ---------------------------------------------------------------------------
// ExportSummary struct
// ---------------------------------------------------------------------------

func TestExportSummary_Aggregation(t *testing.T) {
	summary := ExportSummary{
		TotalModules:   3,
		SuccessModules: 2,
		FailedModules:  1,
		TotalViews:     5,
		Results: []ExportResult{
			{Module: "a", Views: []ExportedView{{ViewKey: "V1"}, {ViewKey: "V2"}}},
			{Module: "b", Views: []ExportedView{{ViewKey: "V1"}, {ViewKey: "V2"}, {ViewKey: "V3"}}},
			{Module: "c", Error: fmt.Errorf("failed")},
		},
		ExecutionTime: 2 * time.Second,
	}

	if summary.TotalModules != 3 {
		t.Errorf("TotalModules = %d, want 3", summary.TotalModules)
	}
	if summary.SuccessModules != 2 {
		t.Errorf("SuccessModules = %d, want 2", summary.SuccessModules)
	}
	if summary.FailedModules != 1 {
		t.Errorf("FailedModules = %d, want 1", summary.FailedModules)
	}
	if summary.TotalViews != 5 {
		t.Errorf("TotalViews = %d, want 5", summary.TotalViews)
	}
	if len(summary.Results) != 3 {
		t.Errorf("Results length = %d, want 3", len(summary.Results))
	}
}

// ---------------------------------------------------------------------------
// viewKeyPattern regex
// ---------------------------------------------------------------------------

func TestViewKeyPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKey  string
		wantMatch bool
	}{
		{
			name:      "standard key",
			input:     "structurizr-SystemContext.svg",
			wantKey:   "SystemContext",
			wantMatch: true,
		},
		{
			name:      "key with number",
			input:     "structurizr-SystemContext-1.svg",
			wantKey:   "SystemContext",
			wantMatch: true,
		},
		{
			name:      "key with large number",
			input:     "structurizr-ContainerView-123.svg",
			wantKey:   "ContainerView",
			wantMatch: true,
		},
		{
			name:      "hyphenated key no number",
			input:     "structurizr-my-view.svg",
			wantKey:   "my-view",
			wantMatch: true,
		},
		{
			name:      "hyphenated key with number",
			input:     "structurizr-my-view-2.svg",
			wantKey:   "my-view",
			wantMatch: true,
		},
		{
			name:      "no structurizr prefix",
			input:     "other-file.svg",
			wantKey:   "",
			wantMatch: false,
		},
		{
			name:      "wrong extension",
			input:     "structurizr-View.png",
			wantKey:   "",
			wantMatch: false,
		},
		{
			name:      "no extension",
			input:     "structurizr-View",
			wantKey:   "",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := viewKeyPattern.FindStringSubmatch(tt.input)
			if tt.wantMatch {
				if len(matches) < 2 {
					t.Fatalf("viewKeyPattern did not match %q", tt.input)
				}
				if matches[1] != tt.wantKey {
					t.Errorf("viewKeyPattern captured %q, want %q", matches[1], tt.wantKey)
				}
			} else {
				if len(matches) >= 2 {
					t.Errorf("viewKeyPattern should not match %q, but captured %q", tt.input, matches[1])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// viewDefinitionPattern regex
// ---------------------------------------------------------------------------

func TestViewDefinitionPattern(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  string
		wantKey   string
		wantMatch bool
	}{
		{
			name:      "systemContext view",
			input:     `    systemContext softwareSystem "Overview"`,
			wantType:  "systemContext",
			wantKey:   "Overview",
			wantMatch: true,
		},
		{
			name:      "container view",
			input:     `    container softwareSystem "ContainerView"`,
			wantType:  "container",
			wantKey:   "ContainerView",
			wantMatch: true,
		},
		{
			name:      "component view",
			input:     `    component backend "Backend Components"`,
			wantType:  "component",
			wantKey:   "Backend Components",
			wantMatch: true,
		},
		{
			name:      "dynamic view",
			input:     `    dynamic softwareSystem "Request Flow"`,
			wantType:  "dynamic",
			wantKey:   "Request Flow",
			wantMatch: true,
		},
		{
			name:      "deployment view captures first quoted string",
			input:     `    deployment environment "Live" "Production Deployment"`,
			wantType:  "deployment",
			wantKey:   "Live",
			wantMatch: true,
		},
		{
			name:      "deployment view single arg",
			input:     `    deployment environment "DeploymentView"`,
			wantType:  "deployment",
			wantKey:   "DeploymentView",
			wantMatch: true,
		},
		{
			name:      "custom view",
			input:     `    custom element "Custom Diagram"`,
			wantType:  "custom",
			wantKey:   "Custom Diagram",
			wantMatch: true,
		},
		{
			name:      "filtered view",
			input:     `    filtered baseView "Filtered"`,
			wantType:  "filtered",
			wantKey:   "Filtered",
			wantMatch: true,
		},
		{
			name:      "unknown view type",
			input:     `    landscape system "LandscapeView"`,
			wantMatch: false,
		},
		{
			name:      "no quotes",
			input:     `    systemContext system NoQuotes`,
			wantMatch: false,
		},
		{
			name:      "empty line",
			input:     "",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := viewDefinitionPattern.FindStringSubmatch(tt.input)
			if tt.wantMatch {
				if len(matches) < 3 {
					t.Fatalf("viewDefinitionPattern did not match %q", tt.input)
				}
				if matches[1] != tt.wantType {
					t.Errorf("viewDefinitionPattern type = %q, want %q", matches[1], tt.wantType)
				}
				if matches[2] != tt.wantKey {
					t.Errorf("viewDefinitionPattern key = %q, want %q", matches[2], tt.wantKey)
				}
			} else {
				if len(matches) >= 3 {
					t.Errorf("viewDefinitionPattern should not match %q, but captured type=%q key=%q",
						tt.input, matches[1], matches[2])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StructurizrExporter interface satisfaction
// ---------------------------------------------------------------------------

func TestStructurizrExporterImpl_ImplementsInterface(t *testing.T) {
	var _ StructurizrExporter = &StructurizrExporterImpl{}
}
