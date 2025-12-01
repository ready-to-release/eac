//go:build L0 && ov

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// HandlersConfig Loading Tests
// =============================================================================

func TestHandlersConfig_Load(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// Handlers should be loaded
	assert.NotNil(t, cfg.Handlers, "Handlers should be loaded")
	assert.NotEmpty(t, cfg.Handlers.Handlers, "Handlers should have definitions")
}

func TestHandlersConfig_LoadFromRepository(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// Verify specific handlers are loaded from handlers.yml
	expectedHandlers := []string{"go", "docker", "mkdocs", "npm", "none"}
	for _, name := range expectedHandlers {
		handler := cfg.Handlers.Get(name)
		assert.NotNil(t, handler, "should find handler: %s", name)
	}
}

func TestHandlersConfig_HandlerMapBuilt(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// Handler map should be built for fast lookup
	assert.NotNil(t, cfg.Handlers.handlerMap)
	assert.Equal(t, len(cfg.Handlers.Handlers), len(cfg.Handlers.handlerMap))
}

// =============================================================================
// Handler Get Tests
// =============================================================================

func TestHandlersConfig_GetHandler(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("get go handler", func(t *testing.T) {
		goHandler := cfg.Handlers.Get("go")
		require.NotNil(t, goHandler, "should find go handler")
		assert.Equal(t, "builtin", goHandler.Type)
		assert.Equal(t, "go", goHandler.Name)
	})

	t.Run("get docker handler", func(t *testing.T) {
		dockerHandler := cfg.Handlers.Get("docker")
		require.NotNil(t, dockerHandler, "should find docker handler")
		assert.Equal(t, "builtin", dockerHandler.Type)
	})

	t.Run("get mkdocs handler", func(t *testing.T) {
		mkdocsHandler := cfg.Handlers.Get("mkdocs")
		require.NotNil(t, mkdocsHandler, "should find mkdocs handler")
		assert.Equal(t, "docker", mkdocsHandler.Type)
	})

	t.Run("get npm handler", func(t *testing.T) {
		npmHandler := cfg.Handlers.Get("npm")
		require.NotNil(t, npmHandler, "should find npm handler")
		assert.Equal(t, "command", npmHandler.Type)
	})

	t.Run("get none handler", func(t *testing.T) {
		noneHandler := cfg.Handlers.Get("none")
		require.NotNil(t, noneHandler, "should find none handler")
		assert.Equal(t, "builtin", noneHandler.Type)
	})

	t.Run("get unknown handler returns nil", func(t *testing.T) {
		unknown := cfg.Handlers.Get("unknown-xyz")
		assert.Nil(t, unknown)
	})

	t.Run("get empty name returns nil", func(t *testing.T) {
		empty := cfg.Handlers.Get("")
		assert.Nil(t, empty)
	})
}

func TestHandlersConfig_Get_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	result := cfg.Get("go")
	assert.Nil(t, result, "Get on nil config should return nil")
}

func TestHandlersConfig_Get_NilHandlerMap(t *testing.T) {
	cfg := &HandlersConfig{
		Handlers: []Handler{
			{Name: "test", Type: "builtin"},
		},
		handlerMap: nil, // Not built yet
	}
	result := cfg.Get("test")
	assert.Nil(t, result, "Get with nil handlerMap should return nil")
}

// =============================================================================
// Cross-Compile Targets Tests
// =============================================================================

func TestHandlersConfig_GetCrossCompileTargets(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	targets := cfg.Handlers.GetCrossCompileTargets()
	assert.NotEmpty(t, targets, "should have cross-compile targets")

	t.Run("has all expected platforms", func(t *testing.T) {
		expectedPlatforms := []struct {
			os   string
			arch string
		}{
			{"linux", "amd64"},
			{"linux", "arm64"},
			{"darwin", "amd64"},
			{"darwin", "arm64"},
			{"windows", "amd64"},
		}

		for _, expected := range expectedPlatforms {
			found := false
			for _, target := range targets {
				if target.OS == expected.os && target.Arch == expected.arch {
					found = true
					break
				}
			}
			assert.True(t, found, "should have %s/%s target", expected.os, expected.arch)
		}
	})

	t.Run("windows has exe suffix", func(t *testing.T) {
		for _, target := range targets {
			if target.OS == "windows" {
				assert.Equal(t, ".exe", target.Suffix, "windows should have .exe suffix")
			}
		}
	})

	t.Run("non-windows has no suffix", func(t *testing.T) {
		for _, target := range targets {
			if target.OS != "windows" {
				assert.Empty(t, target.Suffix, "%s should have no suffix", target.OS)
			}
		}
	})
}

func TestHandlersConfig_GetCrossCompileTargets_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	targets := cfg.GetCrossCompileTargets()
	assert.NotEmpty(t, targets, "nil config should return defaults")
	assert.Len(t, targets, 5, "default should have 5 targets")
}

func TestHandlersConfig_GetCrossCompileTargets_NoGoHandler(t *testing.T) {
	cfg := &HandlersConfig{
		Handlers:   []Handler{},
		handlerMap: make(map[string]*Handler),
	}
	targets := cfg.GetCrossCompileTargets()
	assert.NotEmpty(t, targets, "missing go handler should return defaults")
}

func TestHandlersConfig_GetCrossCompileTargets_NoBuildConfig(t *testing.T) {
	cfg := &HandlersConfig{
		Handlers: []Handler{
			{Name: "go", Type: "builtin", Build: nil},
		},
	}
	cfg.buildHandlerMap()
	targets := cfg.GetCrossCompileTargets()
	assert.NotEmpty(t, targets, "nil build config should return defaults")
}

func TestHandlersConfig_GetCrossCompileTargets_EmptyConfig(t *testing.T) {
	cfg := &HandlersConfig{
		Handlers: []Handler{
			{
				Name: "go",
				Type: "builtin",
				Build: &HandlerSpec{
					Handler: "go",
					Config:  map[string]interface{}{},
				},
			},
		},
	}
	cfg.buildHandlerMap()
	targets := cfg.GetCrossCompileTargets()
	assert.NotEmpty(t, targets, "empty config should return defaults")
}

// =============================================================================
// Dispatch Rule Tests
// =============================================================================

func TestHandlersConfig_GetBuildHandler(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("uses primary build dep by default", func(t *testing.T) {
		handler := cfg.Handlers.GetBuildHandler("go-library", []string{"go_module"}, "go")
		assert.Equal(t, "go", handler)
	})

	t.Run("mkdocs dispatch rule matches documentation+container", func(t *testing.T) {
		handler := cfg.Handlers.GetBuildHandler("mkdocs-site", []string{"documentation", "serveable", "container"}, "docker")
		assert.Equal(t, "mkdocs", handler)
	})

	t.Run("empty build dep returns empty", func(t *testing.T) {
		handler := cfg.Handlers.GetBuildHandler("scripts-package", []string{}, "")
		assert.Equal(t, "", handler)
	})

	t.Run("docker build dep", func(t *testing.T) {
		handler := cfg.Handlers.GetBuildHandler("go-r2r-extension", []string{"go_module", "container"}, "docker")
		// Should match default rule with docker as primary build dep
		assert.Equal(t, "docker", handler)
	})

	t.Run("npm build dep", func(t *testing.T) {
		handler := cfg.Handlers.GetBuildHandler("vscode-ext", []string{"npm_package", "typescript"}, "npm")
		assert.Equal(t, "npm", handler)
	})
}

func TestHandlersConfig_GetBuildHandler_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	handler := cfg.GetBuildHandler("test", []string{}, "go")
	assert.Equal(t, "go", handler, "nil config should return primary build dep")
}

func TestHandlersConfig_GetBuildHandler_NilDispatch(t *testing.T) {
	cfg := &HandlersConfig{
		Handlers:   []Handler{},
		Dispatch:   nil,
		handlerMap: make(map[string]*Handler),
	}
	handler := cfg.GetBuildHandler("test", []string{}, "go")
	assert.Equal(t, "go", handler, "nil dispatch should return primary build dep")
}

func TestHandlersConfig_GetTestHandler(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("go test handler", func(t *testing.T) {
		handler := cfg.Handlers.GetTestHandler("go-library", []string{"go_module"}, "go")
		assert.Equal(t, "go", handler)
	})

	t.Run("npm test handler", func(t *testing.T) {
		handler := cfg.Handlers.GetTestHandler("vscode-ext", []string{"npm_package"}, "npm")
		assert.Equal(t, "npm", handler)
	})

	t.Run("static module no build dep", func(t *testing.T) {
		handler := cfg.Handlers.GetTestHandler("configuration", []string{}, "")
		assert.Equal(t, "", handler)
	})
}

func TestHandlersConfig_GetTestHandler_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	handler := cfg.GetTestHandler("test", []string{}, "go")
	assert.Equal(t, "go", handler, "nil config should return primary build dep")
}

// =============================================================================
// Match Condition Tests
// =============================================================================

func TestHandlersConfig_MatchesRule(t *testing.T) {
	cfg := &HandlersConfig{}

	t.Run("default rule matches everything", func(t *testing.T) {
		rule := DispatchRule{
			Match:   MatchCondition{Default: true},
			Handler: "test",
		}
		assert.True(t, cfg.MatchesRule(rule, "any-type", []string{"any"}, "any"))
	})

	t.Run("type match", func(t *testing.T) {
		rule := DispatchRule{
			Match:   MatchCondition{Type: "go-cli"},
			Handler: "test",
		}
		assert.True(t, cfg.MatchesRule(rule, "go-cli", []string{}, ""))
		assert.False(t, cfg.MatchesRule(rule, "go-library", []string{}, ""))
	})

	t.Run("build dep match", func(t *testing.T) {
		rule := DispatchRule{
			Match:   MatchCondition{BuildDep: "go"},
			Handler: "test",
		}
		assert.True(t, cfg.MatchesRule(rule, "any", []string{}, "go"))
		assert.False(t, cfg.MatchesRule(rule, "any", []string{}, "npm"))
	})

	t.Run("capabilities match - all required", func(t *testing.T) {
		rule := DispatchRule{
			Match:   MatchCondition{Capabilities: []string{"documentation", "container"}},
			Handler: "test",
		}
		assert.True(t, cfg.MatchesRule(rule, "any", []string{"documentation", "container", "extra"}, ""))
		assert.False(t, cfg.MatchesRule(rule, "any", []string{"documentation"}, ""))
		assert.False(t, cfg.MatchesRule(rule, "any", []string{"container"}, ""))
		assert.False(t, cfg.MatchesRule(rule, "any", []string{}, ""))
	})

	t.Run("combined match - type and capabilities", func(t *testing.T) {
		rule := DispatchRule{
			Match: MatchCondition{
				Type:         "mkdocs-site",
				Capabilities: []string{"documentation"},
			},
			Handler: "test",
		}
		assert.True(t, cfg.MatchesRule(rule, "mkdocs-site", []string{"documentation", "container"}, ""))
		assert.False(t, cfg.MatchesRule(rule, "other-type", []string{"documentation", "container"}, ""))
		assert.False(t, cfg.MatchesRule(rule, "mkdocs-site", []string{"container"}, ""))
	})

	t.Run("combined match - build dep and capabilities", func(t *testing.T) {
		rule := DispatchRule{
			Match: MatchCondition{
				BuildDep:     "docker",
				Capabilities: []string{"container"},
			},
			Handler: "test",
		}
		assert.True(t, cfg.MatchesRule(rule, "any", []string{"container"}, "docker"))
		assert.False(t, cfg.MatchesRule(rule, "any", []string{"container"}, "go"))
		assert.False(t, cfg.MatchesRule(rule, "any", []string{}, "docker"))
	})
}

func TestHandlersConfig_ResolveHandlerName(t *testing.T) {
	cfg := &HandlersConfig{}

	t.Run("literal handler name", func(t *testing.T) {
		result := cfg.resolveHandlerName("mkdocs", "docker")
		assert.Equal(t, "mkdocs", result)
	})

	t.Run("primary build dep placeholder", func(t *testing.T) {
		result := cfg.resolveHandlerName("{primary_build_dep}", "go")
		assert.Equal(t, "go", result)
	})

	t.Run("empty primary build dep", func(t *testing.T) {
		result := cfg.resolveHandlerName("{primary_build_dep}", "")
		assert.Equal(t, "", result)
	})
}

// =============================================================================
// UPX Platform Tests
// =============================================================================

func TestHandlersConfig_UPXPlatforms(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("get platforms", func(t *testing.T) {
		platforms := cfg.Handlers.GetUPXPlatforms()
		assert.Contains(t, platforms, "linux")
		assert.Contains(t, platforms, "windows")
	})

	t.Run("linux supported", func(t *testing.T) {
		assert.True(t, cfg.Handlers.IsUPXSupported("linux"))
	})

	t.Run("windows supported", func(t *testing.T) {
		assert.True(t, cfg.Handlers.IsUPXSupported("windows"))
	})

	t.Run("darwin not supported", func(t *testing.T) {
		assert.False(t, cfg.Handlers.IsUPXSupported("darwin"))
	})

	t.Run("unknown platform not supported", func(t *testing.T) {
		assert.False(t, cfg.Handlers.IsUPXSupported("freebsd"))
	})
}

func TestHandlersConfig_GetUPXPlatforms_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	platforms := cfg.GetUPXPlatforms()
	assert.Contains(t, platforms, "linux")
	assert.Contains(t, platforms, "windows")
}

func TestHandlersConfig_IsUPXSupported_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	// Should use defaults
	assert.True(t, cfg.IsUPXSupported("linux"))
	assert.False(t, cfg.IsUPXSupported("darwin"))
}

// =============================================================================
// Dockerfile Paths Tests
// =============================================================================

func TestHandlersConfig_DockerfilePaths(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("get paths", func(t *testing.T) {
		paths := cfg.Handlers.GetDockerfilePaths()
		assert.NotEmpty(t, paths)
	})

	t.Run("has containers path", func(t *testing.T) {
		paths := cfg.Handlers.GetDockerfilePaths()
		assert.Contains(t, paths, "containers/{moniker}/Dockerfile")
	})

	t.Run("has root path", func(t *testing.T) {
		paths := cfg.Handlers.GetDockerfilePaths()
		assert.Contains(t, paths, "{root}/Dockerfile")
	})
}

func TestHandlersConfig_GetDockerfilePaths_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	paths := cfg.GetDockerfilePaths()
	assert.NotEmpty(t, paths, "nil config should return defaults")
	assert.Contains(t, paths, "containers/{moniker}/Dockerfile")
}

func TestResolveDockerfilePath(t *testing.T) {
	tests := []struct {
		name     string
		template string
		moniker  string
		root     string
		expected string
	}{
		{
			name:     "moniker substitution",
			template: "containers/{moniker}/Dockerfile",
			moniker:  "ext-eac",
			root:     "src/ext-eac",
			expected: "containers/ext-eac/Dockerfile",
		},
		{
			name:     "root substitution",
			template: "{root}/Dockerfile",
			moniker:  "ext-eac",
			root:     "src/ext-eac",
			expected: "src/ext-eac/Dockerfile",
		},
		{
			name:     "no substitution",
			template: "Dockerfile",
			moniker:  "test",
			root:     "src/test",
			expected: "Dockerfile",
		},
		{
			name:     "both substitutions",
			template: "{root}/{moniker}.Dockerfile",
			moniker:  "myapp",
			root:     "apps",
			expected: "apps/myapp.Dockerfile",
		},
		{
			name:     "empty moniker",
			template: "containers/{moniker}/Dockerfile",
			moniker:  "",
			root:     "src",
			expected: "containers//Dockerfile",
		},
		{
			name:     "complex path",
			template: "docker/{root}/images/{moniker}/Dockerfile",
			moniker:  "api",
			root:     "services/backend",
			expected: "docker/services/backend/images/api/Dockerfile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveDockerfilePath(tt.template, tt.moniker, tt.root)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// CI Platform Tests
// =============================================================================

func TestHandlersConfig_CIPlatforms(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("get platforms list", func(t *testing.T) {
		platforms := cfg.Handlers.GetCIPlatforms()
		assert.Contains(t, platforms, "linux/amd64")
		assert.Contains(t, platforms, "linux/arm64")
	})

	t.Run("get platforms string", func(t *testing.T) {
		platformStr := cfg.Handlers.GetCIPlatformsString()
		assert.Contains(t, platformStr, "linux/amd64")
		assert.Contains(t, platformStr, ",")
	})
}

func TestHandlersConfig_GetCIPlatforms_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	platforms := cfg.GetCIPlatforms()
	assert.Contains(t, platforms, "linux/amd64")
	assert.Contains(t, platforms, "linux/arm64")
}

func TestHandlersConfig_GetCIPlatformsString_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	platformStr := cfg.GetCIPlatformsString()
	assert.Equal(t, "linux/amd64,linux/arm64", platformStr)
}

// =============================================================================
// MkDocs Handler Tests
// =============================================================================

func TestHandlersConfig_GetMkDocsHandler(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("returns mkdocs handler", func(t *testing.T) {
		handler := cfg.Handlers.GetMkDocsHandler()
		require.NotNil(t, handler)
		assert.Equal(t, "mkdocs", handler.Name)
		assert.Equal(t, "docker", handler.Type)
	})

	t.Run("has build config", func(t *testing.T) {
		handler := cfg.Handlers.GetMkDocsHandler()
		require.NotNil(t, handler)
		assert.NotNil(t, handler.Build)
		assert.NotEmpty(t, handler.Build.Image)
	})
}

func TestHandlersConfig_GetMkDocsHandler_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	handler := cfg.GetMkDocsHandler()
	assert.Nil(t, handler)
}

// =============================================================================
// Handler Validation Tests
// =============================================================================

func TestHandler_Validate(t *testing.T) {
	t.Run("valid builtin handler", func(t *testing.T) {
		h := &Handler{Name: "go", Type: "builtin"}
		err := h.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid command handler", func(t *testing.T) {
		h := &Handler{Name: "npm", Type: "command"}
		err := h.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid script handler", func(t *testing.T) {
		h := &Handler{Name: "custom", Type: "script"}
		err := h.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid docker handler", func(t *testing.T) {
		h := &Handler{Name: "mkdocs", Type: "docker"}
		err := h.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing name", func(t *testing.T) {
		h := &Handler{Name: "", Type: "builtin"}
		err := h.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("invalid type", func(t *testing.T) {
		h := &Handler{Name: "test", Type: "invalid"}
		err := h.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type")
	})

	t.Run("empty type", func(t *testing.T) {
		h := &Handler{Name: "test", Type: ""}
		err := h.Validate()
		assert.Error(t, err)
	})
}

// =============================================================================
// Handler Spec Tests
// =============================================================================

func TestHandlerSpec_BuiltinHandler(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	goHandler := cfg.Handlers.Get("go")
	require.NotNil(t, goHandler)
	require.NotNil(t, goHandler.Build)

	assert.Equal(t, "go", goHandler.Build.Handler)
	assert.NotNil(t, goHandler.Build.Config)
}

func TestHandlerSpec_CommandHandler(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	npmHandler := cfg.Handlers.Get("npm")
	require.NotNil(t, npmHandler)
	require.NotNil(t, npmHandler.Build)

	// Command handlers have steps
	assert.NotEmpty(t, npmHandler.Build.Steps)

	// Check first step
	if len(npmHandler.Build.Steps) > 0 {
		step := npmHandler.Build.Steps[0]
		assert.NotEmpty(t, step.Name)
		assert.NotEmpty(t, step.Command)
	}
}

func TestHandlerSpec_DockerHandler(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	mkdocsHandler := cfg.Handlers.Get("mkdocs")
	require.NotNil(t, mkdocsHandler)
	require.NotNil(t, mkdocsHandler.Build)

	// Docker handlers have image
	assert.NotEmpty(t, mkdocsHandler.Build.Image)
	assert.NotEmpty(t, mkdocsHandler.Build.Command)
}

// =============================================================================
// Command Step Tests
// =============================================================================

func TestCommandStep_ConditionalExecution(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	npmHandler := cfg.Handlers.Get("npm")
	require.NotNil(t, npmHandler)
	require.NotNil(t, npmHandler.Build)

	// Find compile step with conditional
	var compileStep *CommandStep
	for i := range npmHandler.Build.Steps {
		if npmHandler.Build.Steps[i].Name == "compile" {
			compileStep = &npmHandler.Build.Steps[i]
			break
		}
	}

	if compileStep != nil {
		assert.NotEmpty(t, compileStep.When, "compile step should have when condition")
		assert.Contains(t, compileStep.When, "capability:")
	}
}

// =============================================================================
// Build Handler Map Tests
// =============================================================================

func TestHandlersConfig_BuildHandlerMap(t *testing.T) {
	cfg := &HandlersConfig{
		Handlers: []Handler{
			{Name: "handler1", Type: "builtin"},
			{Name: "handler2", Type: "command"},
			{Name: "handler3", Type: "docker"},
		},
	}

	cfg.buildHandlerMap()

	assert.NotNil(t, cfg.handlerMap)
	assert.Len(t, cfg.handlerMap, 3)
	assert.NotNil(t, cfg.handlerMap["handler1"])
	assert.NotNil(t, cfg.handlerMap["handler2"])
	assert.NotNil(t, cfg.handlerMap["handler3"])
}

func TestHandlersConfig_BuildHandlerMap_Empty(t *testing.T) {
	cfg := &HandlersConfig{
		Handlers: []Handler{},
	}

	cfg.buildHandlerMap()

	assert.NotNil(t, cfg.handlerMap)
	assert.Empty(t, cfg.handlerMap)
}

func TestHandlersConfig_BuildHandlerMap_DuplicateNames(t *testing.T) {
	cfg := &HandlersConfig{
		Handlers: []Handler{
			{Name: "duplicate", Type: "builtin"},
			{Name: "duplicate", Type: "command"}, // Same name, different type
		},
	}

	cfg.buildHandlerMap()

	// Last one wins
	assert.Len(t, cfg.handlerMap, 1)
	assert.Equal(t, "command", cfg.handlerMap["duplicate"].Type)
}

// =============================================================================
// Integration Tests with Real Config
// =============================================================================

func TestHandlersConfig_Integration_FullDispatch(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// Test dispatch for each module type
	testCases := []struct {
		moduleType   string
		capabilities []string
		buildDep     string
		expectedBH   string // expected build handler
		expectedTH   string // expected test handler
	}{
		{"go-cli", []string{"go_module", "executable", "cross_compile"}, "go", "go", "go"},
		{"go-library", []string{"go_module"}, "go", "go", "go"},
		{"mkdocs-site", []string{"documentation", "serveable", "container"}, "docker", "mkdocs", "docker"},
		{"vscode-ext", []string{"npm_package", "typescript"}, "npm", "npm", "npm"},
		{"scripts-package", []string{}, "", "", ""},
		{"configuration", []string{}, "", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.moduleType, func(t *testing.T) {
			buildHandler := cfg.Handlers.GetBuildHandler(tc.moduleType, tc.capabilities, tc.buildDep)
			assert.Equal(t, tc.expectedBH, buildHandler, "build handler mismatch for %s", tc.moduleType)

			testHandler := cfg.Handlers.GetTestHandler(tc.moduleType, tc.capabilities, tc.buildDep)
			assert.Equal(t, tc.expectedTH, testHandler, "test handler mismatch for %s", tc.moduleType)
		})
	}
}

func TestHandlersConfig_Integration_AllHandlersValid(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	for _, handler := range cfg.Handlers.Handlers {
		t.Run(handler.Name, func(t *testing.T) {
			err := handler.Validate()
			assert.NoError(t, err, "handler %s should be valid", handler.Name)
		})
	}
}

// =============================================================================
// Handler Flag Tests
// =============================================================================

func TestHandlersConfig_GetBuildFlags(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("go handler has build flags", func(t *testing.T) {
		flags := cfg.Handlers.GetBuildFlags("go")
		assert.NotEmpty(t, flags, "go handler should have build flags")
	})

	t.Run("tidy flag exists", func(t *testing.T) {
		flags := cfg.Handlers.GetBuildFlags("go")
		var found bool
		for _, f := range flags {
			if f.Name == "tidy" {
				found = true
				break
			}
		}
		assert.True(t, found, "should find tidy flag")
	})

	t.Run("compressed flag exists", func(t *testing.T) {
		flags := cfg.Handlers.GetBuildFlags("go")
		var found bool
		for _, f := range flags {
			if f.Name == "compressed" {
				found = true
				break
			}
		}
		assert.True(t, found, "should find compressed flag")
	})

	t.Run("version flag exists", func(t *testing.T) {
		flags := cfg.Handlers.GetBuildFlags("go")
		var found bool
		for _, f := range flags {
			if f.Name == "version" {
				found = true
				break
			}
		}
		assert.True(t, found, "should find version flag")
	})

	t.Run("unknown handler returns nil", func(t *testing.T) {
		flags := cfg.Handlers.GetBuildFlags("nonexistent")
		assert.Nil(t, flags)
	})

	t.Run("handler without flags returns empty", func(t *testing.T) {
		flags := cfg.Handlers.GetBuildFlags("docker")
		assert.Empty(t, flags)
	})
}

func TestHandlersConfig_GetBuildFlags_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	flags := cfg.GetBuildFlags("go")
	assert.Nil(t, flags)
}

func TestHandlersConfig_GetTestFlags(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("go handler test has no flags", func(t *testing.T) {
		flags := cfg.Handlers.GetTestFlags("go")
		assert.Empty(t, flags, "go handler test should have no flags")
	})
}

func TestHandlersConfig_GetTestFlags_NilConfig(t *testing.T) {
	var cfg *HandlersConfig
	flags := cfg.GetTestFlags("go")
	assert.Nil(t, flags)
}

func TestHandlersConfig_GetBuildFlagByName(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("find tidy flag", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByName("go", "tidy")
		require.NotNil(t, flag)
		assert.Equal(t, "tidy", flag.Name)
		assert.Equal(t, "bool", flag.Type)
		assert.Equal(t, "--tidy-first", flag.CLIPositive)
		assert.Equal(t, "--no-tidy", flag.CLINegative)
	})

	t.Run("find version flag", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByName("go", "version")
		require.NotNil(t, flag)
		assert.Equal(t, "version", flag.Name)
		assert.Equal(t, "string", flag.Type)
		assert.Equal(t, "--version", flag.ValueFlag)
	})

	t.Run("unknown flag returns nil", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByName("go", "nonexistent")
		assert.Nil(t, flag)
	})

	t.Run("unknown handler returns nil", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByName("nonexistent", "tidy")
		assert.Nil(t, flag)
	})
}

func TestHandlersConfig_GetBuildFlagByCLI(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("find by positive flag", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByCLI("go", "--tidy-first")
		require.NotNil(t, flag)
		assert.Equal(t, "tidy", flag.Name)
	})

	t.Run("find by negative flag", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByCLI("go", "--no-tidy")
		require.NotNil(t, flag)
		assert.Equal(t, "tidy", flag.Name)
	})

	t.Run("find by value flag", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByCLI("go", "--version")
		require.NotNil(t, flag)
		assert.Equal(t, "version", flag.Name)
	})

	t.Run("unknown cli flag returns nil", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByCLI("go", "--unknown")
		assert.Nil(t, flag)
	})
}

func TestHandlerFlag_GetDefault(t *testing.T) {
	t.Run("simple default", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:    "test",
			Type:    "bool",
			Default: true,
		}
		assert.Equal(t, true, flag.GetDefault(false))
		assert.Equal(t, true, flag.GetDefault(true))
	})

	t.Run("CI-specific default", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:    "tidy",
			Type:    "bool",
			Default: true,
			DefaultEnv: map[string]interface{}{
				"CI": false,
			},
		}
		assert.Equal(t, true, flag.GetDefault(false))  // local
		assert.Equal(t, false, flag.GetDefault(true))  // CI
	})

	t.Run("local-specific default", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:    "verbose",
			Type:    "bool",
			Default: false,
			DefaultEnv: map[string]interface{}{
				"local": true,
			},
		}
		assert.Equal(t, true, flag.GetDefault(false))  // local
		assert.Equal(t, false, flag.GetDefault(true))  // CI (falls back to default)
	})

	t.Run("nil flag returns nil", func(t *testing.T) {
		var flag *HandlerFlag
		assert.Nil(t, flag.GetDefault(false))
	})

	t.Run("nil default returns nil", func(t *testing.T) {
		flag := &HandlerFlag{
			Name: "test",
			Type: "bool",
		}
		assert.Nil(t, flag.GetDefault(false))
	})
}

func TestHandlerFlag_GetBoolDefault(t *testing.T) {
	t.Run("true default", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:    "test",
			Type:    "bool",
			Default: true,
		}
		assert.True(t, flag.GetBoolDefault(false))
	})

	t.Run("false default", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:    "test",
			Type:    "bool",
			Default: false,
		}
		assert.False(t, flag.GetBoolDefault(false))
	})

	t.Run("nil default returns false", func(t *testing.T) {
		flag := &HandlerFlag{
			Name: "test",
			Type: "bool",
		}
		assert.False(t, flag.GetBoolDefault(false))
	})

	t.Run("non-bool default returns false", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:    "test",
			Type:    "bool",
			Default: "not-a-bool",
		}
		assert.False(t, flag.GetBoolDefault(false))
	})

	t.Run("with CI default", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:    "tidy",
			Type:    "bool",
			Default: true,
			DefaultEnv: map[string]interface{}{
				"CI": false,
			},
		}
		assert.True(t, flag.GetBoolDefault(false))   // local
		assert.False(t, flag.GetBoolDefault(true))   // CI
	})
}

func TestHandlerFlag_GetStringDefault(t *testing.T) {
	t.Run("string default", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:    "version",
			Type:    "string",
			Default: "1.0.0",
		}
		assert.Equal(t, "1.0.0", flag.GetStringDefault(false))
	})

	t.Run("empty string default", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:    "version",
			Type:    "string",
			Default: "",
		}
		assert.Equal(t, "", flag.GetStringDefault(false))
	})

	t.Run("nil default returns empty", func(t *testing.T) {
		flag := &HandlerFlag{
			Name: "version",
			Type: "string",
		}
		assert.Equal(t, "", flag.GetStringDefault(false))
	})

	t.Run("non-string default returns empty", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:    "version",
			Type:    "string",
			Default: 123,
		}
		assert.Equal(t, "", flag.GetStringDefault(false))
	})
}

func TestHandlersConfig_GetAllBuildCLIFlags(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("go handler has multiple cli flags", func(t *testing.T) {
		flags := cfg.Handlers.GetAllBuildCLIFlags("go")
		assert.NotEmpty(t, flags)

		// Should include positive, negative, and value flags
		assert.Contains(t, flags, "--tidy-first")
		assert.Contains(t, flags, "--no-tidy")
		assert.Contains(t, flags, "--compressed")
		assert.Contains(t, flags, "--compressed-upx")
		assert.Contains(t, flags, "--version")
	})

	t.Run("unknown handler returns empty map", func(t *testing.T) {
		flags := cfg.Handlers.GetAllBuildCLIFlags("nonexistent")
		assert.Empty(t, flags)
	})

	t.Run("handler without flags returns empty map", func(t *testing.T) {
		flags := cfg.Handlers.GetAllBuildCLIFlags("docker")
		assert.Empty(t, flags)
	})
}

func TestHandlerFlag_Validate(t *testing.T) {
	t.Run("valid bool flag with positive and negative", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:        "tidy",
			Type:        "bool",
			CLIPositive: "--tidy-first",
			CLINegative: "--no-tidy",
		}
		err := flag.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid bool flag with only positive", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:        "compressed",
			Type:        "bool",
			CLIPositive: "--compressed",
		}
		err := flag.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid string flag with value_flag", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:      "version",
			Type:      "string",
			ValueFlag: "--version",
		}
		err := flag.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid string flag with cli_positive", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:        "output",
			Type:        "string",
			CLIPositive: "--output",
		}
		err := flag.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing name", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:        "",
			Type:        "bool",
			CLIPositive: "--test",
		}
		err := flag.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("invalid type", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:        "test",
			Type:        "invalid",
			CLIPositive: "--test",
		}
		err := flag.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type")
	})

	t.Run("bool flag without cli_positive", func(t *testing.T) {
		flag := &HandlerFlag{
			Name: "test",
			Type: "bool",
		}
		err := flag.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cli_positive")
	})

	t.Run("string flag without value_flag or cli_positive", func(t *testing.T) {
		flag := &HandlerFlag{
			Name: "test",
			Type: "string",
		}
		err := flag.Validate()
		assert.Error(t, err)
	})

	t.Run("int flag type is valid", func(t *testing.T) {
		flag := &HandlerFlag{
			Name:      "count",
			Type:      "int",
			ValueFlag: "--count",
		}
		err := flag.Validate()
		assert.NoError(t, err)
	})
}

func TestHandlersConfig_Integration_GoFlagsFromYAML(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	t.Run("tidy flag has correct defaults", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByName("go", "tidy")
		require.NotNil(t, flag)

		// Default is true for local builds
		assert.True(t, flag.GetBoolDefault(false), "tidy should default to true for local")
		// Default is false for CI
		assert.False(t, flag.GetBoolDefault(true), "tidy should default to false for CI")
	})

	t.Run("compressed flag defaults to false", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByName("go", "compressed")
		require.NotNil(t, flag)

		assert.False(t, flag.GetBoolDefault(false))
		assert.False(t, flag.GetBoolDefault(true))
	})

	t.Run("version flag defaults to empty", func(t *testing.T) {
		flag := cfg.Handlers.GetBuildFlagByName("go", "version")
		require.NotNil(t, flag)

		assert.Equal(t, "", flag.GetStringDefault(false))
		assert.Equal(t, "", flag.GetStringDefault(true))
	})
}
