package config

import (
	"regexp"
	"strings"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadTestConfig is a test helper that loads the real contract config.
func loadTestConfig(t *testing.T) *TestingConfig {
	t.Helper()
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadTestingConfig(repoRoot, t.TempDir())
	require.NoError(t, err)
	return cfg
}

// newEmptyConfig creates an empty TestingConfig for edge case testing.
func newEmptyConfig() *TestingConfig {
	return &TestingConfig{
		suites:           make(map[string]*core.SuiteDefinition),
		suiteOrder:       []string{},
		tags:             make(map[string]*core.TagDefinition),
		tagTypes:         make(map[string]core.TagType),
		skipReasons:      []core.SkipReason{},
		compiledPatterns: make(map[string]*regexp.Regexp),
		tagLookup:        make(map[string]*core.TagDefinition),
	}
}

func TestLoadTestingConfig(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	// Use a temp dir as config root (no user overrides)
	configRoot := t.TempDir()

	cfg, err := LoadTestingConfig(repoRoot, configRoot)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	t.Run("loads suite definitions", func(t *testing.T) {
		suites := cfg.ListSuites()
		assert.NotEmpty(t, suites, "should have suites loaded")

		// Check for expected suites
		suite, ok := cfg.GetSuite("unit")
		assert.True(t, ok, "should have unit suite")
		assert.Equal(t, "Unit Tests", suite.Name())
		assert.False(t, suite.IsExtended(), "unit suite should not be extended")
		assert.NotEmpty(t, suite.Selectors())
	})

	t.Run("loads tag definitions", func(t *testing.T) {
		tags := cfg.ListTags()
		assert.NotEmpty(t, tags, "should have tags loaded")

		// Check for expected taxonomy tags
		tag, ok := cfg.GetTag("@L0")
		assert.True(t, ok, "should have @L0 tag")
		assert.Equal(t, "taxonomy-level", tag.Type())
	})

	t.Run("loads skip reasons", func(t *testing.T) {
		reasons := cfg.GetSkipReasons()
		assert.NotEmpty(t, reasons, "should have skip reasons loaded")

		// Find the wip reason
		var foundWIP bool
		for _, r := range reasons {
			if r.Code == "wip" {
				foundWIP = true
				assert.Equal(t, "Work In Progress", r.Name)
				break
			}
		}
		assert.True(t, foundWIP, "should have wip skip reason")
	})
}

func TestTestingConfig_GetSuite(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadTestingConfig(repoRoot, t.TempDir())
	require.NoError(t, err)

	tests := []struct {
		moniker    string
		wantOk     bool
		isExtended bool
	}{
		{"unit", true, false},
		{"integration", true, false},
		{"acceptance", true, true},
		{"production-verification", true, true},
		{"manual", true, true},
		{"nonexistent", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.moniker, func(t *testing.T) {
			suite, ok := cfg.GetSuite(tt.moniker)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.isExtended, suite.IsExtended())
				assert.NotEmpty(t, suite.Name())
				assert.NotEmpty(t, suite.Description())
			}
		})
	}
}

func TestTestingConfig_GetDefaultSuites(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadTestingConfig(repoRoot, t.TempDir())
	require.NoError(t, err)

	defaults := cfg.GetDefaultSuites()

	// Default suites should include unit and integration (non-extended)
	assert.Contains(t, defaults, "unit", "unit should be a default suite")
	assert.Contains(t, defaults, "integration", "integration should be a default suite")

	// Extended suites should not be included
	assert.NotContains(t, defaults, "acceptance", "acceptance should not be a default suite")
	assert.NotContains(t, defaults, "production-verification", "production-verification should not be a default suite")
	assert.NotContains(t, defaults, "manual", "manual should not be a default suite")
}

func TestTestingConfig_GetTagsByType(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadTestingConfig(repoRoot, t.TempDir())
	require.NoError(t, err)

	t.Run("taxonomy-level tags", func(t *testing.T) {
		tags := cfg.GetTagsByType("taxonomy-level")
		assert.NotEmpty(t, tags)
		assert.Contains(t, tags, "@L0")
		assert.Contains(t, tags, "@L1")
		assert.Contains(t, tags, "@L2")
		assert.Contains(t, tags, "@L3")
		assert.Contains(t, tags, "@L4")
	})

	t.Run("verification tags", func(t *testing.T) {
		tags := cfg.GetTagsByType("verification")
		assert.NotEmpty(t, tags)
		assert.Contains(t, tags, "@ov")
		assert.Contains(t, tags, "@iv")
		assert.Contains(t, tags, "@piv")
	})

	t.Run("execution_control tags", func(t *testing.T) {
		tags := cfg.GetTagsByType("execution_control")
		assert.NotEmpty(t, tags)
		assert.Contains(t, tags, "@skip")
		assert.Contains(t, tags, "@pending")
	})

	t.Run("nonexistent type returns empty", func(t *testing.T) {
		tags := cfg.GetTagsByType("nonexistent")
		assert.Empty(t, tags)
	})
}

func TestTestingConfig_SuiteSelectors(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadTestingConfig(repoRoot, t.TempDir())
	require.NoError(t, err)

	suite, ok := cfg.GetSuite("unit")
	require.True(t, ok)

	selectors := suite.Selectors()
	require.NotEmpty(t, selectors)

	// Unit suite should include @L0 and @L1, exclude @L2-@L4
	selector := selectors[0]
	assert.Contains(t, selector.AnyOfTags, "@L0")
	assert.Contains(t, selector.AnyOfTags, "@L1")
	assert.Contains(t, selector.ExcludeTags, "@L2")
	assert.Contains(t, selector.ExcludeTags, "@L3")
	assert.Contains(t, selector.ExcludeTags, "@L4")
}

// =============================================================================
// GetTag - Pattern-based tag matching tests
// =============================================================================

func TestTestingConfig_GetTag_PatternMatching(t *testing.T) {
	cfg := loadTestConfig(t)

	tests := []struct {
		name     string
		tag      string
		wantOk   bool
		wantType string
	}{
		// Exact match tags
		{
			name:     "exact match - L0 taxonomy level",
			tag:      "@L0",
			wantOk:   true,
			wantType: "taxonomy-level",
		},
		{
			name:     "exact match - L4 taxonomy level",
			tag:      "@L4",
			wantOk:   true,
			wantType: "taxonomy-level",
		},
		{
			name:     "exact match - HE2E taxonomy level",
			tag:      "@HE2E",
			wantOk:   true,
			wantType: "taxonomy-level",
		},
		{
			name:     "exact match - ov verification",
			tag:      "@ov",
			wantOk:   true,
			wantType: "verification",
		},
		{
			name:     "exact match - piv verification",
			tag:      "@piv",
			wantOk:   true,
			wantType: "verification",
		},
		{
			name:     "exact match - pending execution control",
			tag:      "@pending",
			wantOk:   true,
			wantType: "execution_control",
		},
		{
			name:     "exact match - Manual execution control",
			tag:      "@Manual",
			wantOk:   true,
			wantType: "execution_control",
		},
		{
			name:     "exact match - simple skip (no reason)",
			tag:      "@skip",
			wantOk:   true,
			wantType: "execution_control",
		},

		// Pattern-based tags - @deps:<name>
		{
			name:     "pattern match - deps:docker",
			tag:      "@deps:docker",
			wantOk:   true,
			wantType: "system_dependency",
		},
		{
			name:     "pattern match - deps:linux",
			tag:      "@deps:linux",
			wantOk:   true,
			wantType: "system_dependency",
		},
		{
			name:     "pattern match - deps:go-1-22",
			tag:      "@deps:go-1-22",
			wantOk:   true,
			wantType: "system_dependency",
		},

		// Pattern-based tags - @depm:<module-name>
		{
			name:     "pattern match - depm:r2r-cli",
			tag:      "@depm:r2r-cli",
			wantOk:   true,
			wantType: "module_dependency",
		},
		{
			name:     "pattern match - depm:core",
			tag:      "@depm:core",
			wantOk:   true,
			wantType: "module_dependency",
		},

		// Pattern-based tags - @env:<env-moniker>
		{
			name:     "pattern match - env:isolated-test-project",
			tag:      "@env:isolated-test-project",
			wantOk:   true,
			wantType: "environment",
		},
		{
			name:     "pattern match - env:staging",
			tag:      "@env:staging",
			wantOk:   true,
			wantType: "environment",
		},

		// Pattern-based tags - @skip:<reason>
		{
			name:     "pattern match - skip:wip",
			tag:      "@skip:wip",
			wantOk:   true,
			wantType: "execution_control",
		},
		{
			name:     "pattern match - skip:broken",
			tag:      "@skip:broken",
			wantOk:   true,
			wantType: "execution_control",
		},
		{
			name:     "pattern match - skip:flaky",
			tag:      "@skip:flaky",
			wantOk:   true,
			wantType: "execution_control",
		},

		// Pattern-based tags - @control:<control-id>
		{
			name:     "pattern match - control:ac-2",
			tag:      "@control:ac-2",
			wantOk:   true,
			wantType: "oscal_control",
		},
		{
			name:     "pattern match - control:ia-5(1)",
			tag:      "@control:ia-5(1)",
			wantOk:   true,
			wantType: "oscal_control",
		},

		// Invalid tags - should not match
		{
			name:   "nonexistent tag",
			tag:    "@nonexistent",
			wantOk: false,
		},
		{
			name:   "malformed deps pattern - uppercase",
			tag:    "@deps:Docker",
			wantOk: false,
		},
		{
			name:   "malformed deps pattern - spaces",
			tag:    "@deps:docker compose",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, ok := cfg.GetTag(tt.tag)
			assert.Equal(t, tt.wantOk, ok, "GetTag(%q) ok", tt.tag)
			if ok && tt.wantType != "" {
				assert.Equal(t, tt.wantType, tag.Type(), "GetTag(%q) type", tt.tag)
			}
		})
	}
}

// =============================================================================
// IsKnownTag tests
// =============================================================================

func TestTestingConfig_IsKnownTag(t *testing.T) {
	cfg := loadTestConfig(t)

	tests := []struct {
		name string
		tag  string
		want bool
	}{
		// Exact match tags
		{"exact - L0", "@L0", true},
		{"exact - L4", "@L4", true},
		{"exact - HE2E", "@HE2E", true},
		{"exact - ov", "@ov", true},
		{"exact - pending", "@pending", true},
		{"exact - skip", "@skip", true},
		{"exact - Manual", "@Manual", true},

		// Pattern match tags
		{"pattern - deps:docker", "@deps:docker", true},
		{"pattern - depm:r2r-cli", "@depm:r2r-cli", true},
		{"pattern - env:isolated", "@env:isolated", true},
		{"pattern - skip:wip", "@skip:wip", true},
		{"pattern - control:ac-2", "@control:ac-2", true},

		// Unknown tags
		{"unknown - random", "@random", false},
		{"unknown - invalid-format", "L0", false}, // missing @
		{"unknown - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.IsKnownTag(tt.tag)
			assert.Equal(t, tt.want, got, "IsKnownTag(%q)", tt.tag)
		})
	}
}

func TestTestingConfig_IsKnownTag_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	assert.False(t, cfg.IsKnownTag("@L0"), "empty config should not recognize @L0")
	assert.False(t, cfg.IsKnownTag("@deps:docker"), "empty config should not recognize patterns")
}

// =============================================================================
// GetTaxonomyLevelTags tests
// =============================================================================

func TestTestingConfig_GetTaxonomyLevelTags(t *testing.T) {
	cfg := loadTestConfig(t)

	tags := cfg.GetTaxonomyLevelTags()

	// Should include all taxonomy levels
	assert.Contains(t, tags, "@L0", "should include @L0")
	assert.Contains(t, tags, "@L1", "should include @L1")
	assert.Contains(t, tags, "@L2", "should include @L2")
	assert.Contains(t, tags, "@L3", "should include @L3")
	assert.Contains(t, tags, "@L4", "should include @L4")
	assert.Contains(t, tags, "@HE2E", "should include @HE2E")

	// Should NOT include non-taxonomy tags
	assert.NotContains(t, tags, "@ov", "should not include verification tags")
	assert.NotContains(t, tags, "@skip", "should not include execution control tags")
	assert.NotContains(t, tags, "@deps:<system-dependency>", "should not include pattern tags")
}

func TestTestingConfig_GetTaxonomyLevelTags_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	tags := cfg.GetTaxonomyLevelTags()
	assert.Empty(t, tags, "empty config should return empty slice")
}

// =============================================================================
// GetVerificationTags tests
// =============================================================================

func TestTestingConfig_GetVerificationTags(t *testing.T) {
	cfg := loadTestConfig(t)

	tags := cfg.GetVerificationTags()

	// Should include all verification types
	assert.Contains(t, tags, "@ov", "should include @ov (operational verification)")
	assert.Contains(t, tags, "@iv", "should include @iv (installation verification)")
	assert.Contains(t, tags, "@pv", "should include @pv (performance verification)")
	assert.Contains(t, tags, "@piv", "should include @piv (production installation verification)")
	assert.Contains(t, tags, "@ppv", "should include @ppv (production performance verification)")

	// Should NOT include non-verification tags
	assert.NotContains(t, tags, "@L0", "should not include taxonomy tags")
	assert.NotContains(t, tags, "@skip", "should not include execution control tags")
}

func TestTestingConfig_GetVerificationTags_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	tags := cfg.GetVerificationTags()
	assert.Empty(t, tags, "empty config should return empty slice")
}

// =============================================================================
// GetValidSkipReasons tests
// =============================================================================

func TestTestingConfig_GetValidSkipReasons(t *testing.T) {
	cfg := loadTestConfig(t)

	reasons := cfg.GetValidSkipReasons()

	// Should include expected skip reasons
	assert.Contains(t, reasons, "wip", "should include wip")
	assert.Contains(t, reasons, "broken", "should include broken")
	assert.Contains(t, reasons, "flaky", "should include flaky")
	assert.Contains(t, reasons, "deprecated", "should include deprecated")
	assert.Contains(t, reasons, "blocked", "should include blocked")

	// Should return strings only (no SkipReason structs)
	for _, r := range reasons {
		assert.NotEmpty(t, r, "reason code should not be empty")
		assert.IsType(t, "", r, "reason should be a string")
	}
}

func TestTestingConfig_GetValidSkipReasons_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	reasons := cfg.GetValidSkipReasons()
	assert.Empty(t, reasons, "empty config should return empty slice")
}

// =============================================================================
// ValidateSkipReason tests
// =============================================================================

func TestTestingConfig_ValidateSkipReason(t *testing.T) {
	cfg := loadTestConfig(t)

	tests := []struct {
		name       string
		code       string
		wantOk     bool
		wantName   string
		wantHasMsg bool
	}{
		{
			name:     "valid - wip",
			code:     "wip",
			wantOk:   true,
			wantName: "Work In Progress",
		},
		{
			name:     "valid - broken",
			code:     "broken",
			wantOk:   true,
			wantName: "Broken Test",
		},
		{
			name:     "valid - flaky",
			code:     "flaky",
			wantOk:   true,
			wantName: "Flaky Test",
		},
		{
			name:     "valid - deprecated",
			code:     "deprecated",
			wantOk:   true,
			wantName: "Deprecated Feature",
		},
		{
			name:     "valid - blocked",
			code:     "blocked",
			wantOk:   true,
			wantName: "Blocked",
		},
		{
			name:   "invalid - empty",
			code:   "",
			wantOk: false,
		},
		{
			name:   "invalid - unknown",
			code:   "unknown",
			wantOk: false,
		},
		{
			name:   "invalid - uppercase",
			code:   "WIP",
			wantOk: false,
		},
		{
			name:   "invalid - with prefix",
			code:   "@skip:wip",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason, ok := cfg.ValidateSkipReason(tt.code)
			assert.Equal(t, tt.wantOk, ok, "ValidateSkipReason(%q) ok", tt.code)
			if ok {
				assert.Equal(t, tt.code, reason.Code, "reason.Code")
				assert.Equal(t, tt.wantName, reason.Name, "reason.Name")
				assert.NotEmpty(t, reason.Description, "reason.Description should not be empty")
			}
		})
	}
}

func TestTestingConfig_ValidateSkipReason_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	reason, ok := cfg.ValidateSkipReason("wip")
	assert.False(t, ok, "empty config should not validate any reason")
	assert.Empty(t, reason.Code, "reason code should be empty")
}

// =============================================================================
// GetSkipTags tests
// =============================================================================

func TestTestingConfig_GetSkipTags(t *testing.T) {
	cfg := loadTestConfig(t)

	tags := cfg.GetSkipTags()

	// Tags should contain all skip reasons
	assert.Contains(t, tags, "@skip:wip", "should contain wip")
	assert.Contains(t, tags, "@skip:broken", "should contain broken")
	assert.Contains(t, tags, "@skip:flaky", "should contain flaky")
	assert.Contains(t, tags, "@skip:deprecated", "should contain deprecated")
	assert.Contains(t, tags, "@skip:blocked", "should contain blocked")
}

func TestTestingConfig_GetSkipTags_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	tags := cfg.GetSkipTags()
	assert.Empty(t, tags, "empty config should return empty slice")
}

func TestTestingConfig_GetSkipTags_Format(t *testing.T) {
	cfg := loadTestConfig(t)

	tags := cfg.GetSkipTags()
	assert.NotEmpty(t, tags, "should have tags")

	for _, tag := range tags {
		assert.True(t, strings.HasPrefix(tag, "@skip:"),
			"each tag should start with @skip:, got: %s", tag)
	}
}

// =============================================================================
// GetSkipTagsForSuite tests
// =============================================================================

func TestTestingConfig_GetSkipTagsForSuite(t *testing.T) {
	cfg := loadTestConfig(t)

	tags := cfg.GetSkipTagsForSuite()

	// Should include all skip reason tags
	assert.Contains(t, tags, "@skip:wip", "should include @skip:wip")
	assert.Contains(t, tags, "@skip:broken", "should include @skip:broken")
	assert.Contains(t, tags, "@skip:flaky", "should include @skip:flaky")
	assert.Contains(t, tags, "@skip:deprecated", "should include @skip:deprecated")
	assert.Contains(t, tags, "@skip:blocked", "should include @skip:blocked")

	// Should include @pending
	assert.Contains(t, tags, "@pending", "should include @pending")
}

func TestTestingConfig_GetSkipTagsForSuite_AlwaysIncludesPending(t *testing.T) {
	cfg := loadTestConfig(t)

	tags := cfg.GetSkipTagsForSuite()

	// @pending should always be the last element
	assert.Equal(t, "@pending", tags[len(tags)-1], "@pending should be last")
}

func TestTestingConfig_GetSkipTagsForSuite_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	tags := cfg.GetSkipTagsForSuite()

	// Even with no skip reasons, @pending should be included
	assert.Contains(t, tags, "@pending", "should always include @pending")
	assert.Len(t, tags, 1, "should only have @pending")
}

// =============================================================================
// ValidateTag tests
// =============================================================================

func TestTestingConfig_ValidateTag(t *testing.T) {
	cfg := loadTestConfig(t)

	tests := []struct {
		name    string
		tag     string
		wantErr bool
		errMsg  string
	}{
		// Valid exact match tags
		{"valid - L0", "@L0", false, ""},
		{"valid - L4", "@L4", false, ""},
		{"valid - HE2E", "@HE2E", false, ""},
		{"valid - ov", "@ov", false, ""},
		{"valid - pending", "@pending", false, ""},
		{"valid - skip", "@skip", false, ""},

		// Valid pattern match tags
		{"valid - deps:docker", "@deps:docker", false, ""},
		{"valid - depm:r2r-cli", "@depm:r2r-cli", false, ""},
		{"valid - env:staging", "@env:staging", false, ""},
		{"valid - control:ac-2", "@control:ac-2", false, ""},

		// Valid skip with valid reason
		{"valid - skip:wip", "@skip:wip", false, ""},
		{"valid - skip:broken", "@skip:broken", false, ""},
		{"valid - skip:flaky", "@skip:flaky", false, ""},

		// Invalid skip with invalid reason
		{
			name:    "invalid - skip:invalid",
			tag:     "@skip:invalid",
			wantErr: true,
			errMsg:  "invalid skip reason",
		},
		{
			name:    "invalid - skip:UPPERCASE",
			tag:     "@skip:UPPERCASE",
			wantErr: true,
			errMsg:  "invalid format", // UPPERCASE doesn't match lowercase pattern
		},

		// Malformed pattern tags
		{
			name:    "malformed - deps:Invalid",
			tag:     "@deps:Invalid",
			wantErr: true,
			errMsg:  "invalid format",
		},
		{
			name:    "malformed - deps with space",
			tag:     "@deps:docker compose",
			wantErr: true,
			errMsg:  "invalid format",
		},

		// Unknown tags (not an error by design - handled separately)
		{"unknown tag - no error", "@custom-unknown", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cfg.ValidateTag(tt.tag)
			if tt.wantErr {
				assert.Error(t, err, "ValidateTag(%q) should error", tt.tag)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "error message")
				}
			} else {
				assert.NoError(t, err, "ValidateTag(%q) should not error", tt.tag)
			}
		})
	}
}

func TestTestingConfig_ValidateTag_SkipReasonValidation(t *testing.T) {
	cfg := loadTestConfig(t)

	// Get all valid skip reasons and verify they pass validation
	validReasons := cfg.GetValidSkipReasons()
	for _, reason := range validReasons {
		tag := "@skip:" + reason
		t.Run("valid-"+reason, func(t *testing.T) {
			err := cfg.ValidateTag(tag)
			assert.NoError(t, err, "ValidateTag(%q) should not error for valid reason", tag)
		})
	}

	// Test some invalid reasons
	invalidReasons := []string{"invalid", "unknown", "xyz", "123", ""}
	for _, reason := range invalidReasons {
		tag := "@skip:" + reason
		// Skip empty string test as it changes the tag format
		if reason == "" {
			continue
		}
		t.Run("invalid-"+reason, func(t *testing.T) {
			err := cfg.ValidateTag(tag)
			assert.Error(t, err, "ValidateTag(%q) should error for invalid reason", tag)
		})
	}
}

// =============================================================================
// HasConstraint tests
// =============================================================================

func TestTestingConfig_HasConstraint(t *testing.T) {
	cfg := loadTestConfig(t)

	tests := []struct {
		name       string
		tag        string
		constraint string
		want       bool
	}{
		{
			name:       "Manual tag has mutual exclusion constraint",
			tag:        "@Manual",
			constraint: "mutually_exclusive_with_taxonomy_levels",
			want:       true,
		},
		{
			name:       "Manual tag - wrong constraint",
			tag:        "@Manual",
			constraint: "nonexistent_constraint",
			want:       false,
		},
		{
			name:       "L0 tag - no constraint",
			tag:        "@L0",
			constraint: "mutually_exclusive_with_taxonomy_levels",
			want:       false,
		},
		{
			name:       "ov tag - no constraint",
			tag:        "@ov",
			constraint: "any_constraint",
			want:       false,
		},
		{
			name:       "unknown tag - no constraint",
			tag:        "@unknown",
			constraint: "any_constraint",
			want:       false,
		},
		{
			name:       "empty tag",
			tag:        "",
			constraint: "any_constraint",
			want:       false,
		},
		{
			name:       "empty constraint",
			tag:        "@Manual",
			constraint: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.HasConstraint(tt.tag, tt.constraint)
			assert.Equal(t, tt.want, got, "HasConstraint(%q, %q)", tt.tag, tt.constraint)
		})
	}
}

func TestTestingConfig_HasConstraint_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	got := cfg.HasConstraint("@Manual", "mutually_exclusive_with_taxonomy_levels")
	assert.False(t, got, "empty config should return false")
}

// =============================================================================
// GetSuiteLTags tests
// =============================================================================

func TestTestingConfig_GetSuiteLTags(t *testing.T) {
	cfg := loadTestConfig(t)

	tests := []struct {
		name     string
		moniker  string
		wantTags []string
		wantNil  bool
	}{
		{
			name:     "unit suite - L0 and L1",
			moniker:  "unit",
			wantTags: []string{"@L0", "@L1"},
		},
		{
			name:     "integration suite - L2",
			moniker:  "integration",
			wantTags: []string{"@L2"},
		},
		{
			name:     "acceptance suite - L3",
			moniker:  "acceptance",
			wantTags: []string{"@L3"},
		},
		{
			name:    "nonexistent suite",
			moniker: "nonexistent",
			wantNil: true,
		},
		{
			name:    "empty moniker",
			moniker: "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetSuiteLTags(tt.moniker)
			if tt.wantNil {
				assert.Nil(t, got, "GetSuiteLTags(%q) should return nil", tt.moniker)
			} else {
				for _, tag := range tt.wantTags {
					assert.Contains(t, got, tag, "GetSuiteLTags(%q) should contain %s", tt.moniker, tag)
				}
			}
		})
	}
}

func TestTestingConfig_GetSuiteLTags_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	got := cfg.GetSuiteLTags("unit")
	assert.Nil(t, got, "empty config should return nil")
}

// =============================================================================
// GetLTagToSuiteMap tests
// =============================================================================

func TestTestingConfig_GetLTagToSuiteMap(t *testing.T) {
	cfg := loadTestConfig(t)

	tagMap := cfg.GetLTagToSuiteMap()

	// Should map L-tags to their suites
	assert.NotEmpty(t, tagMap, "tag map should not be empty")

	// Unit suite covers L0 and L1
	assert.Equal(t, "unit", tagMap["@L0"], "@L0 should map to unit")
	assert.Equal(t, "unit", tagMap["@L1"], "@L1 should map to unit")

	// Integration suite covers L2
	assert.Equal(t, "integration", tagMap["@L2"], "@L2 should map to integration")

	// Acceptance suite covers L3
	assert.Equal(t, "acceptance", tagMap["@L3"], "@L3 should map to acceptance")

	// Only L-tags should be in the map
	for tag := range tagMap {
		assert.True(t, strings.HasPrefix(tag, "@L"),
			"all keys should be L-tags, got: %s", tag)
	}
}

func TestTestingConfig_GetLTagToSuiteMap_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	tagMap := cfg.GetLTagToSuiteMap()
	assert.Empty(t, tagMap, "empty config should return empty map")
}

// =============================================================================
// GetSuiteForLTag tests
// =============================================================================

func TestTestingConfig_GetSuiteForLTag(t *testing.T) {
	cfg := loadTestConfig(t)

	tests := []struct {
		name     string
		ltag     string
		wantSuit string
	}{
		{"L0 -> unit", "@L0", "unit"},
		{"L1 -> unit", "@L1", "unit"},
		{"L2 -> integration", "@L2", "integration"},
		{"L3 -> acceptance", "@L3", "acceptance"},
		{"unknown tag", "@L99", ""},
		{"non-L tag", "@ov", ""},
		{"empty tag", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetSuiteForLTag(tt.ltag)
			assert.Equal(t, tt.wantSuit, got, "GetSuiteForLTag(%q)", tt.ltag)
		})
	}
}

func TestTestingConfig_GetSuiteForLTag_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	got := cfg.GetSuiteForLTag("@L0")
	assert.Empty(t, got, "empty config should return empty string")
}

// =============================================================================
// ListNonProduction tests
// =============================================================================

func TestTestingConfig_ListNonProduction(t *testing.T) {
	cfg := loadTestConfig(t)

	suites := cfg.ListNonProduction()

	// Should include non-production suites
	assert.Contains(t, suites, "unit", "should include unit")
	assert.Contains(t, suites, "integration", "should include integration")
	assert.Contains(t, suites, "acceptance", "should include acceptance")
	assert.Contains(t, suites, "manual", "should include manual")

	// Should NOT include production-verification (requires @L4)
	assert.NotContains(t, suites, "production-verification",
		"should not include production-verification (requires @L4)")
}

func TestTestingConfig_ListNonProduction_ExcludesL4Suites(t *testing.T) {
	cfg := loadTestConfig(t)

	suites := cfg.ListNonProduction()

	// Verify production-verification is excluded
	for _, suite := range suites {
		s, ok := cfg.GetSuite(suite)
		if !ok {
			continue
		}

		// Check that no suite in the list requires @L4
		for _, sel := range s.Selectors() {
			for _, tag := range sel.RequireTags {
				assert.NotEqual(t, "@L4", tag,
					"suite %s should not require @L4", suite)
			}
		}
	}
}

func TestTestingConfig_ListNonProduction_EmptyConfig(t *testing.T) {
	cfg := newEmptyConfig()

	suites := cfg.ListNonProduction()
	assert.Empty(t, suites, "empty config should return empty slice")
}

func TestTestingConfig_ListNonProduction_PreservesOrder(t *testing.T) {
	cfg := loadTestConfig(t)

	suites := cfg.ListNonProduction()
	allSuites := cfg.ListSuites()

	// Non-production suites should maintain the same relative order as all suites
	suiteIdx := make(map[string]int)
	for i, s := range allSuites {
		suiteIdx[s] = i
	}

	for i := 0; i < len(suites)-1; i++ {
		assert.Less(t, suiteIdx[suites[i]], suiteIdx[suites[i+1]],
			"suite order should be preserved: %s should come before %s",
			suites[i], suites[i+1])
	}
}
