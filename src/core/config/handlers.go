package config

import (
	"fmt"
	"strings"
)

// HandlersFileName is the config file name for handlers
const HandlersFileName = "handlers.yml"

// HandlersConfig represents the handlers.yml configuration
type HandlersConfig struct {
	Handlers []Handler      `yaml:"handlers"`
	Dispatch *DispatchRules `yaml:"dispatch,omitempty"`

	// Runtime lookups (built after loading)
	handlerMap map[string]*Handler
}

// Handler represents a build/test handler definition
type Handler struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description,omitempty"`
	Type        string       `yaml:"type"` // builtin, command, script, docker
	Build       *HandlerSpec `yaml:"build,omitempty"`
	Test        *HandlerSpec `yaml:"test,omitempty"`
}

// HandlerSpec represents a handler specification for build or test
type HandlerSpec struct {
	// For builtin handlers
	Handler string                 `yaml:"handler,omitempty"`
	Config  map[string]interface{} `yaml:"config,omitempty"`

	// For command handlers
	Steps []CommandStep `yaml:"steps,omitempty"`

	// For script handlers
	Script      string            `yaml:"script,omitempty"`
	Interpreter string            `yaml:"interpreter,omitempty"`
	Args        []string          `yaml:"args,omitempty"`

	// For docker handlers
	Image      string            `yaml:"image,omitempty"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
	Context    string            `yaml:"context,omitempty"`
	Command    []string          `yaml:"command,omitempty"`
	Workdir    string            `yaml:"workdir,omitempty"`
	Volumes    []string          `yaml:"volumes,omitempty"`
	Env        map[string]string `yaml:"env,omitempty"`
}

// CommandStep represents a single command step in a command handler
type CommandStep struct {
	Name            string            `yaml:"name"`
	Command         string            `yaml:"command"`
	Workdir         string            `yaml:"workdir,omitempty"`
	Env             map[string]string `yaml:"env,omitempty"`
	When            string            `yaml:"when,omitempty"`
	ContinueOnError bool              `yaml:"continue_on_error,omitempty"`
}

// DispatchRules defines rules for selecting handlers
type DispatchRules struct {
	Build []DispatchRule `yaml:"build,omitempty"`
	Test  []DispatchRule `yaml:"test,omitempty"`
}

// DispatchRule defines a single dispatch rule
type DispatchRule struct {
	Match   MatchCondition `yaml:"match"`
	Handler string         `yaml:"handler"`
}

// MatchCondition defines conditions for matching a module
type MatchCondition struct {
	Type         string   `yaml:"type,omitempty"`
	Capabilities []string `yaml:"capabilities,omitempty"`
	BuildDep     string   `yaml:"build_dep,omitempty"`
	Default      bool     `yaml:"default,omitempty"`
}

// CrossCompileTarget represents a cross-compilation target platform
type CrossCompileTarget struct {
	OS     string `yaml:"os"`
	Arch   string `yaml:"arch"`
	Suffix string `yaml:"suffix,omitempty"`
}

// BuildHandlerMap creates a lookup map for handlers
func (c *HandlersConfig) BuildHandlerMap() {
	c.handlerMap = make(map[string]*Handler)
	for i := range c.Handlers {
		c.handlerMap[c.Handlers[i].Name] = &c.Handlers[i]
	}
}

// buildHandlerMap is an alias for BuildHandlerMap (for internal use)
func (c *HandlersConfig) buildHandlerMap() {
	c.BuildHandlerMap()
}

// Get returns a handler by name
func (c *HandlersConfig) Get(name string) *Handler {
	if c == nil || c.handlerMap == nil {
		return nil
	}
	return c.handlerMap[name]
}

// GetBuildHandler returns the build handler name for a module type
// It evaluates dispatch rules and falls back to primary build dep
func (c *HandlersConfig) GetBuildHandler(moduleType string, capabilities []string, primaryBuildDep string) string {
	if c == nil || c.Dispatch == nil {
		return primaryBuildDep
	}

	for _, rule := range c.Dispatch.Build {
		if c.MatchesRule(rule, moduleType, capabilities, primaryBuildDep) {
			return c.resolveHandlerName(rule.Handler, primaryBuildDep)
		}
	}

	return primaryBuildDep
}

// GetTestHandler returns the test handler name for a module type
func (c *HandlersConfig) GetTestHandler(moduleType string, capabilities []string, primaryBuildDep string) string {
	if c == nil || c.Dispatch == nil {
		return primaryBuildDep
	}

	for _, rule := range c.Dispatch.Test {
		if c.MatchesRule(rule, moduleType, capabilities, primaryBuildDep) {
			return c.resolveHandlerName(rule.Handler, primaryBuildDep)
		}
	}

	return primaryBuildDep
}

// MatchesRule checks if a dispatch rule matches the given module properties
func (c *HandlersConfig) MatchesRule(rule DispatchRule, moduleType string, capabilities []string, primaryBuildDep string) bool {
	match := rule.Match

	// Default rule matches if nothing else does
	if match.Default {
		return true
	}

	// Check type match
	if match.Type != "" && match.Type != moduleType {
		return false
	}

	// Check build_dep match
	if match.BuildDep != "" && match.BuildDep != primaryBuildDep {
		return false
	}

	// Check capabilities match (all must be present)
	if len(match.Capabilities) > 0 {
		capSet := make(map[string]bool)
		for _, cap := range capabilities {
			capSet[cap] = true
		}
		for _, reqCap := range match.Capabilities {
			if !capSet[reqCap] {
				return false
			}
		}
	}

	return true
}

// resolveHandlerName resolves a handler name, replacing {primary_build_dep} placeholder
func (c *HandlersConfig) resolveHandlerName(handler string, primaryBuildDep string) string {
	if handler == "{primary_build_dep}" {
		return primaryBuildDep
	}
	return handler
}

// GetCrossCompileTargets returns cross-compile targets from the go handler config
func (c *HandlersConfig) GetCrossCompileTargets() []CrossCompileTarget {
	if c == nil {
		return defaultCrossCompileTargets()
	}

	goHandler := c.Get("go")
	if goHandler == nil || goHandler.Build == nil || goHandler.Build.Config == nil {
		return defaultCrossCompileTargets()
	}

	targetsRaw, ok := goHandler.Build.Config["cross_compile_targets"]
	if !ok {
		return defaultCrossCompileTargets()
	}

	targets, ok := targetsRaw.([]interface{})
	if !ok {
		return defaultCrossCompileTargets()
	}

	var result []CrossCompileTarget
	for _, t := range targets {
		if tMap, ok := t.(map[string]interface{}); ok {
			target := CrossCompileTarget{}
			if os, ok := tMap["os"].(string); ok {
				target.OS = os
			}
			if arch, ok := tMap["arch"].(string); ok {
				target.Arch = arch
			}
			if suffix, ok := tMap["suffix"].(string); ok {
				target.Suffix = suffix
			}
			if target.OS != "" && target.Arch != "" {
				result = append(result, target)
			}
		}
	}

	if len(result) == 0 {
		return defaultCrossCompileTargets()
	}

	return result
}

// defaultCrossCompileTargets returns the default cross-compile targets
func defaultCrossCompileTargets() []CrossCompileTarget {
	return []CrossCompileTarget{
		{OS: "linux", Arch: "amd64", Suffix: ""},
		{OS: "linux", Arch: "arm64", Suffix: ""},
		{OS: "darwin", Arch: "amd64", Suffix: ""},
		{OS: "darwin", Arch: "arm64", Suffix: ""},
		{OS: "windows", Arch: "amd64", Suffix: ".exe"},
	}
}

// GetUPXPlatforms returns platforms that support UPX compression
func (c *HandlersConfig) GetUPXPlatforms() []string {
	if c == nil {
		return []string{"linux", "windows"}
	}

	goHandler := c.Get("go")
	if goHandler == nil || goHandler.Build == nil || goHandler.Build.Config == nil {
		return []string{"linux", "windows"}
	}

	platformsRaw, ok := goHandler.Build.Config["upx_platforms"]
	if !ok {
		return []string{"linux", "windows"}
	}

	platforms, ok := platformsRaw.([]interface{})
	if !ok {
		return []string{"linux", "windows"}
	}

	var result []string
	for _, p := range platforms {
		if s, ok := p.(string); ok {
			result = append(result, s)
		}
	}

	if len(result) == 0 {
		return []string{"linux", "windows"}
	}

	return result
}

// IsUPXSupported checks if a platform supports UPX compression
func (c *HandlersConfig) IsUPXSupported(platform string) bool {
	for _, p := range c.GetUPXPlatforms() {
		if p == platform {
			return true
		}
	}
	return false
}

// GetDockerfilePaths returns the dockerfile search paths for the docker handler
func (c *HandlersConfig) GetDockerfilePaths() []string {
	if c == nil {
		return []string{"containers/{moniker}/Dockerfile", "{root}/Dockerfile"}
	}

	dockerHandler := c.Get("docker")
	if dockerHandler == nil || dockerHandler.Build == nil || dockerHandler.Build.Config == nil {
		return []string{"containers/{moniker}/Dockerfile", "{root}/Dockerfile"}
	}

	pathsRaw, ok := dockerHandler.Build.Config["dockerfile_paths"]
	if !ok {
		return []string{"containers/{moniker}/Dockerfile", "{root}/Dockerfile"}
	}

	paths, ok := pathsRaw.([]interface{})
	if !ok {
		return []string{"containers/{moniker}/Dockerfile", "{root}/Dockerfile"}
	}

	var result []string
	for _, p := range paths {
		if s, ok := p.(string); ok {
			result = append(result, s)
		}
	}

	if len(result) == 0 {
		return []string{"containers/{moniker}/Dockerfile", "{root}/Dockerfile"}
	}

	return result
}

// ResolveDockerfilePath resolves a dockerfile path template
func ResolveDockerfilePath(pathTemplate string, moniker string, root string) string {
	result := pathTemplate
	result = strings.ReplaceAll(result, "{moniker}", moniker)
	result = strings.ReplaceAll(result, "{root}", root)
	return result
}

// GetMkDocsHandler returns the mkdocs handler specification
func (c *HandlersConfig) GetMkDocsHandler() *Handler {
	if c == nil {
		return nil
	}
	return c.Get("mkdocs")
}

// GetCIPlatforms returns the CI platforms for docker builds
func (c *HandlersConfig) GetCIPlatforms() []string {
	if c == nil {
		return []string{"linux/amd64", "linux/arm64"}
	}

	dockerHandler := c.Get("docker")
	if dockerHandler == nil || dockerHandler.Build == nil || dockerHandler.Build.Config == nil {
		return []string{"linux/amd64", "linux/arm64"}
	}

	platformsRaw, ok := dockerHandler.Build.Config["ci_platforms"]
	if !ok {
		return []string{"linux/amd64", "linux/arm64"}
	}

	platforms, ok := platformsRaw.([]interface{})
	if !ok {
		return []string{"linux/amd64", "linux/arm64"}
	}

	var result []string
	for _, p := range platforms {
		if s, ok := p.(string); ok {
			result = append(result, s)
		}
	}

	if len(result) == 0 {
		return []string{"linux/amd64", "linux/arm64"}
	}

	return result
}

// GetCIPlatformsString returns CI platforms as a comma-separated string
func (c *HandlersConfig) GetCIPlatformsString() string {
	return strings.Join(c.GetCIPlatforms(), ",")
}

// ValidateHandler validates that a handler is properly defined
func (h *Handler) Validate() error {
	if h.Name == "" {
		return fmt.Errorf("handler name is required")
	}

	validTypes := map[string]bool{"builtin": true, "command": true, "script": true, "docker": true}
	if !validTypes[h.Type] {
		return fmt.Errorf("handler %s has invalid type: %s", h.Name, h.Type)
	}

	return nil
}
