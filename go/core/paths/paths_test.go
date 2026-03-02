package paths

import (
	"path/filepath"
	"testing"
)

// TestNewPathHelpers tests the newly added path helper functions.
func TestNewPathHelpers(t *testing.T) {
	repoRoot := "/repo"
	moniker := "test-module"

	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "BuildOutputDir",
			fn:       func() string { return BuildOutputDir(repoRoot) },
			expected: filepath.Join(repoRoot, "out", "build"),
		},
		{
			name:     "BuildLogPath",
			fn:       func() string { return BuildLogPath(repoRoot, moniker) },
			expected: filepath.Join(repoRoot, "out", "build", moniker, "build.log"),
		},
		{
			name:     "BuildTimingPath",
			fn:       func() string { return BuildTimingPath(repoRoot, moniker) },
			expected: filepath.Join(repoRoot, "out", "build", moniker, "build-timing.txt"),
		},
		{
			name:     "TestModuleDir",
			fn:       func() string { return TestModuleDir(repoRoot, moniker) },
			expected: filepath.Join(repoRoot, "out", "test", moniker),
		},
		{
			name:     "TestModuleTimingPath",
			fn:       func() string { return TestModuleTimingPath(repoRoot, moniker) },
			expected: filepath.Join(repoRoot, "out", "test", moniker, "test-timing.txt"),
		},
		{
			name:     "TestOutputDir",
			fn:       func() string { return TestOutputDir(repoRoot) },
			expected: filepath.Join(repoRoot, "out", "test"),
		},
		{
			name:     "RiskControlsPath",
			fn:       func() string { return RiskControlsPath(repoRoot) },
			expected: filepath.Join(repoRoot, "specs", ".risk-controls"),
		},
		{
			name:     "RiskCatalogPath",
			fn:       func() string { return RiskCatalogPath(repoRoot) },
			expected: filepath.Join(repoRoot, "templates", "specs", "risk-catalog", "controls.catalog.json"),
		},
		{
			name:     "TemplateSpecsPath",
			fn:       func() string { return TemplateSpecsPath(repoRoot, "risk-catalog") },
			expected: filepath.Join(repoRoot, "templates", "specs", "risk-catalog"),
		},
		{
			name:     "TemplateReportsPath",
			fn:       func() string { return TemplateReportsPath(repoRoot, "summary.md") },
			expected: filepath.Join(repoRoot, "templates", "reports", "summary.md"),
		},
		{
			name:     "SpecsFeaturePath_WithModule",
			fn:       func() string { return SpecsFeaturePath(repoRoot, moniker, "my-feature") },
			expected: filepath.Join(repoRoot, "specs", moniker, "my-feature", "specification.feature"),
		},
		{
			name:     "SpecsFeaturePath_NoModule",
			fn:       func() string { return SpecsFeaturePath(repoRoot, "", "top-level-feature") },
			expected: filepath.Join(repoRoot, "specs", "top-level-feature", "specification.feature"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("%s() = %q, expected %q", tt.name, result, tt.expected)
			}
		})
	}
}

// TestDeployPathHelpers tests the deploy-specific path helper functions.
func TestDeployPathHelpers(t *testing.T) {
	repoRoot := "/repo"
	moniker := "infra"
	env := "development"

	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "DeployOutputPath",
			fn:       func() string { return DeployOutputPath(repoRoot, moniker, env) },
			expected: filepath.Join(repoRoot, "out", "deploy", moniker, env),
		},
		{
			name:     "DeployLogPath",
			fn:       func() string { return DeployLogPath(repoRoot, moniker, env) },
			expected: filepath.Join(repoRoot, "out", "deploy", moniker, env, "deploy.log"),
		},
		{
			name:     "DeployEvidencePath",
			fn:       func() string { return DeployEvidencePath(repoRoot, moniker, env) },
			expected: filepath.Join(repoRoot, "out", "deploy", moniker, env, "deploy-evidence.json"),
		},
		{
			name:     "DeployOutputPath_Production",
			fn:       func() string { return DeployOutputPath(repoRoot, "app", "production") },
			expected: filepath.Join(repoRoot, "out", "deploy", "app", "production"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("%s() = %q, expected %q", tt.name, result, tt.expected)
			}
		})
	}
}

// TestCachePathHelpers tests the cache path helper functions.
func TestCachePathHelpers(t *testing.T) {
	repoRoot := "/repo"

	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "CacheRootPath",
			fn:       func() string { return CacheRootPath(repoRoot) },
			expected: filepath.Join(repoRoot, ".cache", "eac"),
		},
		{
			name:     "BuildCachePath",
			fn:       func() string { return BuildCachePath(repoRoot) },
			expected: filepath.Join(repoRoot, ".cache", "eac", "build"),
		},
		{
			name:     "FileHashCachePath",
			fn:       func() string { return FileHashCachePath(repoRoot, "my-book") },
			expected: filepath.Join(repoRoot, ".cache", "eac", "build", "hashes", "my-book.json"),
		},
		{
			name:     "PDFScreenshotsCachePath",
			fn:       func() string { return PDFScreenshotsCachePath(repoRoot) },
			expected: filepath.Join(repoRoot, ".cache", "eac", "pdf-screenshots"),
		},
		{
			name:     "PDFScreenshotsDirPath",
			fn:       func() string { return PDFScreenshotsDirPath(repoRoot, "abc123") },
			expected: filepath.Join(repoRoot, ".cache", "eac", "pdf-screenshots", "abc123"),
		},
		{
			name:     "StagingCachePath",
			fn:       func() string { return StagingCachePath(repoRoot) },
			expected: filepath.Join(repoRoot, ".cache", "eac", "staging"),
		},
		{
			name:     "BookStagingCachePath",
			fn:       func() string { return BookStagingCachePath(repoRoot, "docs:site", "my-book") },
			expected: filepath.Join(repoRoot, ".cache", "eac", "staging", "docs:site", "my-book"),
		},
		{
			name:     "BuildStateCachePath",
			fn:       func() string { return BuildStateCachePath(repoRoot) },
			expected: filepath.Join(repoRoot, ".cache", "eac", "build", "state"),
		},
		{
			name:     "IncrementalCachePath",
			fn:       func() string { return IncrementalCachePath(repoRoot) },
			expected: filepath.Join(repoRoot, ".cache", "eac", "incremental"),
		},
		{
			name:     "SemaphoreCachePath",
			fn:       func() string { return SemaphoreCachePath(repoRoot) },
			expected: filepath.Join(repoRoot, ".cache", "eac", "semaphores"),
		},
		{
			name:     "PreprocessCachePath",
			fn:       func() string { return PreprocessCachePath(repoRoot) },
			expected: filepath.Join(repoRoot, ".cache", "eac", "preprocess"),
		},
		{
			name:     "NpmWorkCachePath",
			fn:       func() string { return NpmWorkCachePath(repoRoot) },
			expected: filepath.Join(repoRoot, ".cache", "eac", "npm", "work"),
		},
		{
			name:     "NpmDownloadCachePath",
			fn:       func() string { return NpmDownloadCachePath(repoRoot) },
			expected: filepath.Join(repoRoot, ".cache", "eac", "npm", "cache"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("%s() = %q, expected %q", tt.name, result, tt.expected)
			}
		})
	}
}

// TestPathConstants validates that path constants haven't changed.
func TestPathConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"OutDir", OutDir, "out"},
		{"SpecsDir", SpecsDir, "specs"},
		{"TemplatesDir", TemplatesDir, "templates"},
		{"CLIEDir", CLIEDir, ".clie"},
		{"BuildDir", BuildDir, "build"},
		{"DeployDir", DeployDir, "deploy"},
		{"TestDir", TestDir, "test"},
		{"LogsDir", LogsDir, "logs"},
		{"RiskControlsDir", RiskControlsDir, ".risk-controls"},
		{"ReleaseDir", ReleaseDir, "release"},
		{"EACCacheRoot", EACCacheRoot, ".cache/eac"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, expected %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

// TestSanitizeForCacheName tests converting source paths to cache-friendly identifiers.
func TestSanitizeForCacheName(t *testing.T) {
	tests := []struct {
		name       string
		sourcePath string
		want       string
	}{
		{
			name:       "drawio_in_architecture",
			sourcePath: "docs/assets/architecture/modules-overview.drawio.png",
			want:       "architecture_modules-overview",
		},
		{
			name:       "drawio_in_assisted_with_numbers",
			sourcePath: "docs/assets/assisted/01-cd-model-12-stages.drawio.png",
			want:       "assisted_01-cd-model-12-stages",
		},
		{
			name:       "markdown_in_cd-model",
			sourcePath: "docs/explanation/cd-model/overview.md",
			want:       "cd-model_overview",
		},
		{
			name:       "markdown_in_eac_architecture",
			sourcePath: "docs/reference/eac/architecture/index.md",
			want:       "architecture_index",
		},
		{
			name:       "very_long_path_truncated",
			sourcePath: "docs/explanation/continuous-delivery/release-management/very-long-feature-name-that-exceeds-the-limit.md",
			want:       "release-management_very-long-feature-name-that-exceeds-the-limit",
		},
		{
			name:       "windows_path_with_backslashes",
			sourcePath: "docs\\assets\\architecture\\modules-overview.drawio.png",
			want:       "architecture_modules-overview",
		},
		{
			name:       "deeply_nested_path",
			sourcePath: "docs/reference/eac/commands/categories/build.md",
			want:       "categories_build",
		},
		{
			name:       "simple_file_at_root",
			sourcePath: "docs/index.md",
			want:       "docs_index",
		},
		{
			name:       "file_in_assets_cache",
			sourcePath: "docs/assets/cache/drawio/somefile.png",
			want:       "drawio_somefile",
		},
		{
			name:       "path_with_multiple_extensions",
			sourcePath: "docs/assets/diagram.drawio.svg.png",
			want:       "assets_diagram.drawio.svg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeForCacheName(tt.sourcePath)
			if got != tt.want {
				t.Errorf("SanitizeForCacheName(%q) = %q, want %q", tt.sourcePath, got, tt.want)
			}

			// Verify length constraint (max 64 chars)
			if len(got) > 64 {
				t.Errorf("SanitizeForCacheName(%q) returned %d chars, want <= 64", tt.sourcePath, len(got))
			}
		})
	}
}

// TestDrawioCachePath tests the traceable drawio cache path generation.
func TestDrawioCachePath(t *testing.T) {
	tests := []struct {
		name        string
		cacheRoot   string
		sourcePath  string
		contentHash string
		want        string
	}{
		{
			name:        "basic_drawio_cache_path",
			cacheRoot:   "/repo/docs/assets/cache",
			sourcePath:  "docs/assets/architecture/modules-overview.drawio.png",
			contentHash: "abc123def456789",
			want:        filepath.Join("/repo/docs/assets/cache", "drawio", "architecture_modules-overview_abc123de.png"),
		},
		{
			name:        "short_hash_preserved",
			cacheRoot:   "/cache",
			sourcePath:  "docs/assets/assisted/01-cd-model-12-stages.drawio.png",
			contentHash: "12345678",
			want:        filepath.Join("/cache", "drawio", "assisted_01-cd-model-12-stages_12345678.png"),
		},
		{
			name:        "hash_exactly_8_chars",
			cacheRoot:   "/cache",
			sourcePath:  "docs/assets/test.drawio.png",
			contentHash: "abcdefgh",
			want:        filepath.Join("/cache", "drawio", "assets_test_abcdefgh.png"),
		},
		{
			name:        "long_hash_truncated_to_8",
			cacheRoot:   "/cache",
			sourcePath:  "docs/assets/test.drawio.png",
			contentHash: "0123456789abcdef0123456789abcdef",
			want:        filepath.Join("/cache", "drawio", "assets_test_01234567.png"),
		},
		{
			name:        "empty_hash_handled",
			cacheRoot:   "/cache",
			sourcePath:  "docs/assets/test.drawio.png",
			contentHash: "",
			want:        filepath.Join("/cache", "drawio", "assets_test_.png"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DrawioCachePath(tt.cacheRoot, tt.sourcePath, tt.contentHash)
			if got != tt.want {
				t.Errorf("DrawioCachePath(%q, %q, %q) = %q, want %q",
					tt.cacheRoot, tt.sourcePath, tt.contentHash, got, tt.want)
			}
		})
	}
}

// TestMermaidCachePath tests the traceable mermaid cache path generation.
func TestMermaidCachePath(t *testing.T) {
	tests := []struct {
		name        string
		cacheRoot   string
		sourcePath  string
		blockIndex  int
		contentHash string
		want        string
	}{
		{
			name:        "basic_mermaid_cache_path",
			cacheRoot:   "/repo/docs/assets/cache",
			sourcePath:  "docs/explanation/cd-model/overview.md",
			blockIndex:  0,
			contentHash: "abc123def456789",
			want:        filepath.Join("/repo/docs/assets/cache", "mermaid", "cd-model_overview_0_abc123de.svg"),
		},
		{
			name:        "second_block_in_file",
			cacheRoot:   "/cache",
			sourcePath:  "docs/reference/eac/architecture/index.md",
			blockIndex:  1,
			contentHash: "deadbeef12345678",
			want:        filepath.Join("/cache", "mermaid", "architecture_index_1_deadbeef.svg"),
		},
		{
			name:        "high_block_index",
			cacheRoot:   "/cache",
			sourcePath:  "docs/test.md",
			blockIndex:  99,
			contentHash: "abcdefgh",
			want:        filepath.Join("/cache", "mermaid", "docs_test_99_abcdefgh.svg"),
		},
		{
			name:        "zero_block_index",
			cacheRoot:   "/cache",
			sourcePath:  "docs/overview.md",
			blockIndex:  0,
			contentHash: "12345678",
			want:        filepath.Join("/cache", "mermaid", "docs_overview_0_12345678.svg"),
		},
		{
			name:        "long_hash_truncated_to_8",
			cacheRoot:   "/cache",
			sourcePath:  "docs/test.md",
			blockIndex:  0,
			contentHash: "0123456789abcdef0123456789abcdef",
			want:        filepath.Join("/cache", "mermaid", "docs_test_0_01234567.svg"),
		},
		{
			name:        "windows_path_support",
			cacheRoot:   "C:\\projects\\cache",
			sourcePath:  "docs\\reference\\index.md",
			blockIndex:  2,
			contentHash: "wxyz1234",
			want:        filepath.Join("C:\\projects\\cache", "mermaid", "reference_index_2_wxyz1234.svg"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MermaidCachePath(tt.cacheRoot, tt.sourcePath, tt.blockIndex, tt.contentHash)
			if got != tt.want {
				t.Errorf("MermaidCachePath(%q, %q, %d, %q) = %q, want %q",
					tt.cacheRoot, tt.sourcePath, tt.blockIndex, tt.contentHash, got, tt.want)
			}
		})
	}
}

// TestTemplatePathVariadic tests the variadic template path functions.
func TestTemplatePathVariadic(t *testing.T) {
	repoRoot := "/repo"

	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "TemplateSpecsPath_NoSubpaths",
			fn:       func() string { return TemplateSpecsPath(repoRoot) },
			expected: filepath.Join(repoRoot, "templates", "specs"),
		},
		{
			name:     "TemplateSpecsPath_OneSubpath",
			fn:       func() string { return TemplateSpecsPath(repoRoot, "risk-catalog") },
			expected: filepath.Join(repoRoot, "templates", "specs", "risk-catalog"),
		},
		{
			name:     "TemplateSpecsPath_MultipleSubpaths",
			fn:       func() string { return TemplateSpecsPath(repoRoot, "risk-catalog", "controls.json") },
			expected: filepath.Join(repoRoot, "templates", "specs", "risk-catalog", "controls.json"),
		},
		{
			name:     "TemplateReportsPath_NoSubpaths",
			fn:       func() string { return TemplateReportsPath(repoRoot) },
			expected: filepath.Join(repoRoot, "templates", "reports"),
		},
		{
			name:     "TemplateReportsPath_OneSubpath",
			fn:       func() string { return TemplateReportsPath(repoRoot, "summary.md") },
			expected: filepath.Join(repoRoot, "templates", "reports", "summary.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("%s() = %q, expected %q", tt.name, result, tt.expected)
			}
		})
	}
}
