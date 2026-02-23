// helpers.go - Shared utility functions for builders
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	dockerutil "github.com/ready-to-release/eac/go/adapters/docker/util"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/platform"
	"github.com/ready-to-release/eac/go/core/tool"
)

// Logln writes a formatted string with platform-specific line ending to the writer.
func Logln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+platform.LineEnding, args...)
}

// RunCommandWithLog executes a command in the specified directory via the tool executor.
// Output is streamed to the provided writer.
// Returns exit code (0 = success, non-zero = failure).
func RunCommandWithLog(ctx context.Context, dir string, logWriter io.Writer, name string, args ...string) int {
	toolDef := tool.GlobalRegistry().GetOrAdhoc(name)
	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: dir,
		ModuleRoot:    dir,
		LogWriter:     logWriter,
		StdoutWriter:  logWriter,
		StderrWriter:  logWriter,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(ctx, toolDef, execCtx)
	if err != nil {
		Logln(logWriter, "\nError: failed to execute command: %v", err)
		return 1
	}

	if result.ExitCode != 0 {
		Logln(logWriter, "[debug] command exited with code %d", result.ExitCode)
	} else {
		Logln(logWriter, "[debug] command completed successfully (exit code 0)")
	}
	return result.ExitCode
}

// RunCommandWithEnv executes a command with custom environment variables via the tool executor.
func RunCommandWithEnv(ctx context.Context, dir string, logWriter io.Writer, env []string, name string, args ...string) int {
	toolDef := tool.GlobalRegistry().GetOrAdhoc(name)

	// Build full env by appending custom vars to current environment
	fullEnv := append(os.Environ(), env...)

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: dir,
		ModuleRoot:    dir,
		LogWriter:     logWriter,
		StdoutWriter:  logWriter,
		StderrWriter:  logWriter,
		FullEnv:       fullEnv,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(ctx, toolDef, execCtx)
	if err != nil {
		Logln(logWriter, "\nError: failed to execute command: %v", err)
		return 1
	}

	return result.ExitCode
}

// RunCommandWithStdin executes a command with stdin input via the tool executor.
// Output is streamed to the provided writer.
// Returns exit code (0 = success, non-zero = failure).
func RunCommandWithStdin(ctx context.Context, dir string, logWriter io.Writer, stdin io.Reader, name string, args ...string) int {
	toolDef := tool.GlobalRegistry().GetOrAdhoc(name)
	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: dir,
		ModuleRoot:    dir,
		LogWriter:     logWriter,
		StdoutWriter:  logWriter,
		StderrWriter:  logWriter,
		StdinReader:   stdin,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(ctx, toolDef, execCtx)
	if err != nil {
		Logln(logWriter, "\nError: failed to execute command: %v", err)
		return 1
	}
	return result.ExitCode
}

// CopyFile copies a file from src to dst, preserving permissions.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

// FormatDockerVolumePath formats a path for use as a Docker volume mount source
// On Windows, converts C:\path to /c/path for Docker compatibility.
func FormatDockerVolumePath(path string) string {
	return dockerutil.FormatDockerVolume(path)
}

// detectDockerBuilder returns the docker-driver builder for the active Docker context.
// Docker Desktop can switch between contexts (e.g., "default" vs "desktop-linux"),
// and each context has a matching docker-driver builder with the same name.
// We always resolve this explicitly to avoid context/builder mismatch errors
// and to ignore any non-docker-driver builder the user may have set as their buildx default.
func detectDockerBuilder() string {
	toolDef := tool.GlobalRegistry().GetOrAdhoc("docker")
	execCtx := &tool.ExecutionContext{
		ArgsOverrides: []string{"context", "show"},
	}
	result, err := tool.GlobalExecutor().Execute(context.Background(), toolDef, execCtx)
	if err == nil && result.ExitCode == 0 {
		name := strings.TrimSpace(string(result.Stdout))
		if name != "" {
			return name
		}
	}
	return "default"
}

// IsDockerInDocker detects if we're running inside a Docker container.
func IsDockerInDocker() bool {
	return dockerutil.IsDinD()
}

// IsDockerAvailable checks if Docker daemon is accessible.
func IsDockerAvailable() bool {
	return dockerutil.IsDockerAvailable()
}

// ExecutePostBuildSteps runs any post-build steps defined for the component.
// Parameters:
//   - moniker: Module moniker
//   - component: Component name (e.g., "typescript", "go")
//   - workspaceRoot: Absolute path to repository root
//   - outputDir: Absolute path to component build output (out/build/<module>/<component>/)
//   - logWriter: Writer for log output
//
// Returns 0 on success, non-zero on failure.
func ExecutePostBuildSteps(moniker, component, workspaceRoot, outputDir string, logWriter io.Writer) int {
	cfg := config.Global()
	if cfg == nil || cfg.Repository == nil {
		return 0
	}

	// Get module configuration
	module := cfg.Repository.GetByMoniker(moniker)
	if module == nil {
		return 0
	}

	// Get component entry
	compEntry, ok := module.Components[component]
	if !ok || compEntry == nil || compEntry.Build == nil || compEntry.Build.PostBuild == nil {
		return 0
	}

	postBuild := compEntry.Build.PostBuild

	// Execute copy_files if configured
	for _, cf := range postBuild.CopyFiles {
		if compEntry.Root == "" {
			continue
		}
		componentRoot := filepath.Join(workspaceRoot, compEntry.Root)

		if containsGlobChars(cf.From) {
			if code := copyFilesGlob(componentRoot, workspaceRoot, cf, logWriter); code != 0 {
				return code
			}
		} else {
			if code := copyFileLiteral(componentRoot, workspaceRoot, cf, logWriter); code != 0 {
				return code
			}
		}
	}

	return 0
}

// containsGlobChars returns true if s contains glob metacharacters.
func containsGlobChars(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// globStaticPrefix returns the path prefix before the first glob segment.
// For example, "out/**/*" returns "out", "**/*" returns "".
func globStaticPrefix(pattern string) string {
	parts := strings.Split(filepath.ToSlash(pattern), "/")
	var prefix []string
	for _, p := range parts {
		if containsGlobChars(p) {
			break
		}
		prefix = append(prefix, p)
	}
	return strings.Join(prefix, "/")
}

// copyFileLiteral copies a single file (no glob) from component root to workspace.
func copyFileLiteral(componentRoot, workspaceRoot string, cf config.CopyFileEntry, logWriter io.Writer) int {
	srcPath := filepath.Join(componentRoot, cf.From)
	dstPath := filepath.Join(workspaceRoot, cf.To)

	// Validate paths are within workspace (security check)
	srcRel, err := filepath.Rel(workspaceRoot, srcPath)
	if err != nil || strings.HasPrefix(srcRel, "..") {
		Logln(logWriter, "Error: copy_files source must be within workspace: %s", cf.From)
		return 1
	}
	dstRel, err := filepath.Rel(workspaceRoot, dstPath)
	if err != nil || strings.HasPrefix(dstRel, "..") {
		Logln(logWriter, "Error: copy_files target must be within workspace: %s", cf.To)
		return 1
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		Logln(logWriter, "Error: failed to create directory for %s: %v", cf.To, err)
		return 1
	}

	if err := CopyFile(srcPath, dstPath); err != nil {
		Logln(logWriter, "Error: failed to copy %s to %s: %v", cf.From, cf.To, err)
		return 1
	}
	Logln(logWriter, "Post-build: copied %s to %s", cf.From, cf.To)
	return 0
}

// copyFilesGlob expands a glob pattern from component root and copies matched files
// into the target directory, preserving directory structure relative to the glob's
// static prefix.
func copyFilesGlob(componentRoot, workspaceRoot string, cf config.CopyFileEntry, logWriter io.Writer) int {
	// Validate target directory is within workspace
	dstBase := filepath.Join(workspaceRoot, cf.To)
	dstRel, err := filepath.Rel(workspaceRoot, dstBase)
	if err != nil || strings.HasPrefix(dstRel, "..") {
		Logln(logWriter, "Error: copy_files target must be within workspace: %s", cf.To)
		return 1
	}

	// Use forward slashes for doublestar pattern
	pattern := filepath.ToSlash(cf.From)

	matches, err := doublestar.Glob(os.DirFS(componentRoot), pattern)
	if err != nil {
		Logln(logWriter, "Error: invalid glob pattern %s: %v", cf.From, err)
		return 1
	}

	prefix := globStaticPrefix(pattern)
	copied := 0

	for _, match := range matches {
		srcPath := filepath.Join(componentRoot, match)

		info, err := os.Stat(srcPath)
		if err != nil || info.IsDir() {
			continue
		}

		// Validate source is within workspace
		srcRel, err := filepath.Rel(workspaceRoot, srcPath)
		if err != nil || strings.HasPrefix(srcRel, "..") {
			Logln(logWriter, "Error: copy_files source must be within workspace: %s", match)
			return 1
		}

		// Strip the static prefix to get the relative path under the target directory
		relPath := match
		if prefix != "" {
			relPath = strings.TrimPrefix(match, prefix+"/")
		}

		dstPath := filepath.Join(dstBase, relPath)

		// Validate destination is within workspace
		dstPathRel, err := filepath.Rel(workspaceRoot, dstPath)
		if err != nil || strings.HasPrefix(dstPathRel, "..") {
			Logln(logWriter, "Error: copy_files target must be within workspace: %s", dstPath)
			return 1
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			Logln(logWriter, "Error: failed to create directory for %s: %v", relPath, err)
			return 1
		}

		if err := CopyFile(srcPath, dstPath); err != nil {
			Logln(logWriter, "Error: failed to copy %s to %s: %v", match, cf.To, err)
			return 1
		}
		copied++
	}

	Logln(logWriter, "Post-build: copied %d files matching %s to %s", copied, cf.From, cf.To)
	return 0
}

// substituteVars replaces variable placeholders in a string.
func substituteVars(s string, vars map[string]string) string {
	result := s
	for k, v := range vars {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}
