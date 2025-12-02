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

	// Handler-specific flags (available for all handler types)
	Flags []HandlerFlag `yaml:"flags,omitempty"`

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

// HandlerFlag defines a flag that can be passed to a handler
type HandlerFlag struct {
	Name        string            `yaml:"name"`                  // Flag name (used in code)
	Type        string            `yaml:"type"`                  // bool, string, int
	CLIPositive string            `yaml:"cli_positive,omitempty"` // e.g., "--tidy-first"
	CLINegative string            `yaml:"cli_negative,omitempty"` // e.g., "--no-tidy" (for bool flags)
	Default     interface{}       `yaml:"default,omitempty"`     // Default value
	DefaultEnv  map[string]interface{} `yaml:"default_env,omitempty"` // Default by environment (CI: false)
	Description string            `yaml:"description,omitempty"` // Help text
	ValueFlag   string            `yaml:"value_flag,omitempty"`  // For string flags: "--version VALUE"
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

// GetBuildFlags returns the flags defined for a handler's build operation
func (c *HandlersConfig) GetBuildFlags(handlerName string) []HandlerFlag {
	if c == nil {
		return nil
	}
	handler := c.Get(handlerName)
	if handler == nil || handler.Build == nil {
		return nil
	}
	return handler.Build.Flags
}

// GetTestFlags returns the flags defined for a handler's test operation
func (c *HandlersConfig) GetTestFlags(handlerName string) []HandlerFlag {
	if c == nil {
		return nil
	}
	handler := c.Get(handlerName)
	if handler == nil || handler.Test == nil {
		return nil
	}
	return handler.Test.Flags
}

// GetFlagByName returns a specific flag definition by name from build flags
func (c *HandlersConfig) GetBuildFlagByName(handlerName, flagName string) *HandlerFlag {
	flags := c.GetBuildFlags(handlerName)
	for i := range flags {
		if flags[i].Name == flagName {
			return &flags[i]
		}
	}
	return nil
}

// GetFlagByCLI returns a flag definition matching a CLI flag (positive or negative)
func (c *HandlersConfig) GetBuildFlagByCLI(handlerName, cliFlag string) *HandlerFlag {
	flags := c.GetBuildFlags(handlerName)
	for i := range flags {
		if flags[i].CLIPositive == cliFlag || flags[i].CLINegative == cliFlag {
			return &flags[i]
		}
		if flags[i].ValueFlag == cliFlag {
			return &flags[i]
		}
	}
	return nil
}

// GetFlagDefault returns the effective default value for a flag, considering environment
func (f *HandlerFlag) GetDefault(isCI bool) interface{} {
	if f == nil {
		return nil
	}

	// Check environment-specific defaults
	if f.DefaultEnv != nil {
		if isCI {
			if ciVal, ok := f.DefaultEnv["CI"]; ok {
				return ciVal
			}
		} else {
			if localVal, ok := f.DefaultEnv["local"]; ok {
				return localVal
			}
		}
	}

	return f.Default
}

// GetBoolDefault returns the default value as a bool
func (f *HandlerFlag) GetBoolDefault(isCI bool) bool {
	val := f.GetDefault(isCI)
	if val == nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// GetStringDefault returns the default value as a string
func (f *HandlerFlag) GetStringDefault(isCI bool) string {
	val := f.GetDefault(isCI)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetAllBuildCLIFlags returns all CLI flags for a handler's build operation
// Returns a map of cli-flag -> HandlerFlag for easy lookup
func (c *HandlersConfig) GetAllBuildCLIFlags(handlerName string) map[string]*HandlerFlag {
	result := make(map[string]*HandlerFlag)
	flags := c.GetBuildFlags(handlerName)
	for i := range flags {
		flag := &flags[i]
		if flag.CLIPositive != "" {
			result[flag.CLIPositive] = flag
		}
		if flag.CLINegative != "" {
			result[flag.CLINegative] = flag
		}
		if flag.ValueFlag != "" {
			result[flag.ValueFlag] = flag
		}
	}
	return result
}

// GetAllTestCLIFlags returns all CLI flags for a handler's test operation
func (c *HandlersConfig) GetAllTestCLIFlags(handlerName string) map[string]*HandlerFlag {
	result := make(map[string]*HandlerFlag)
	flags := c.GetTestFlags(handlerName)
	for i := range flags {
		flag := &flags[i]
		if flag.CLIPositive != "" {
			result[flag.CLIPositive] = flag
		}
		if flag.CLINegative != "" {
			result[flag.CLINegative] = flag
		}
		if flag.ValueFlag != "" {
			result[flag.ValueFlag] = flag
		}
	}
	return result
}

// ValidateFlag validates that a flag is properly defined
func (f *HandlerFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("flag name is required")
	}

	validTypes := map[string]bool{"bool": true, "string": true, "int": true}
	if !validTypes[f.Type] {
		return fmt.Errorf("flag %s has invalid type: %s (must be bool, string, or int)", f.Name, f.Type)
	}

	// Bool flags should have CLIPositive and optionally CLINegative
	if f.Type == "bool" {
		if f.CLIPositive == "" {
			return fmt.Errorf("flag %s (bool) must have cli_positive defined", f.Name)
		}
	}

	// String/int flags should have ValueFlag
	if f.Type == "string" || f.Type == "int" {
		if f.ValueFlag == "" && f.CLIPositive == "" {
			return fmt.Errorf("flag %s (%s) must have value_flag or cli_positive defined", f.Name, f.Type)
		}
	}

	return nil
}
