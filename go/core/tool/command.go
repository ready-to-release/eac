package tool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// BuildCommand creates an exec.Cmd from a ToolDefinition and ExecutionContext
// without executing it. This is for callers that need pipe-based or streaming
// execution patterns that the standard executor does not support.
//
// For system tools, returns a command using the binary directly.
// For container tools, returns a "docker run" command with proper mounts/env.
//
// The returned command has:
//   - Binary resolved from the tool definition (or "docker" for containers)
//   - Arguments from tool.Args/Command + execCtx.ArgsOverrides (placeholders resolved)
//   - Working directory set (for system tools) or container workdir (for containers)
//   - Environment variables merged
//   - Process group configured for clean termination
//
// The caller is responsible for starting, waiting, and process lifecycle.
// Returns nil if tool or execCtx is nil.
func BuildCommand(ctx context.Context, td *ToolDefinition, execCtx *ExecutionContext) *exec.Cmd {
	if td == nil || execCtx == nil {
		return nil
	}

	switch td.Type {
	case ToolTypeSystem:
		return buildSystemCommand(ctx, td, execCtx)
	case ToolTypeContainer:
		return buildContainerCommand(ctx, td, execCtx)
	default:
		return nil
	}
}

// buildSystemCommand creates an exec.Cmd for a system binary tool.
func buildSystemCommand(ctx context.Context, td *ToolDefinition, execCtx *ExecutionContext) *exec.Cmd {

	// Build argument list
	args := make([]string, 0, len(td.Args)+len(execCtx.ArgsOverrides))
	args = append(args, td.Args...)
	args = append(args, execCtx.ArgsOverrides...)

	// Resolve placeholders in arguments
	for i, arg := range args {
		args[i] = resolvePlaceholders(arg, execCtx.Placeholders)
	}

	cmd := exec.CommandContext(ctx, td.Binary, args...)
	SetProcessGroup(cmd)

	// Set working directory
	workDir := execCtx.ModuleRoot
	if td.WorkDir != "" {
		workDir = resolvePlaceholders(td.WorkDir, execCtx.Placeholders)
	}
	if workDir != "" {
		if filepath.IsAbs(workDir) {
			cmd.Dir = workDir
		} else {
			cmd.Dir = filepath.Join(execCtx.WorkspaceRoot, workDir)
		}
	} else {
		cmd.Dir = execCtx.WorkspaceRoot
	}

	// Set environment
	if execCtx.FullEnv != nil {
		cmd.Env = execCtx.FullEnv
	} else {
		cmd.Env = os.Environ()
	}
	for k, v := range td.Env {
		cmd.Env = append(cmd.Env, k+"="+resolvePlaceholders(v, execCtx.Placeholders))
	}
	for k, v := range execCtx.EnvOverrides {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	return cmd
}

// buildContainerCommand creates a docker run command for pipe-based execution.
// This allows callers to use the same pipe patterns (StdoutPipe, StderrPipe)
// for both system and container tools.
func buildContainerCommand(ctx context.Context, td *ToolDefinition, execCtx *ExecutionContext) *exec.Cmd {
	args := []string{"run", "--rm"}

	// Interactive mode if stdin is expected (allows pipe attachment)
	args = append(args, "-i")

	// Container working directory
	if td.WorkDir != "" {
		workDir := resolvePlaceholders(td.WorkDir, execCtx.Placeholders)
		args = append(args, "-w", workDir)
	}

	// Volume mounts
	for _, mount := range td.Mounts {
		resolved := mount.ResolvePlaceholders(execCtx.Placeholders)

		// Ensure mount source is absolute
		source := resolved.Source
		if !filepath.IsAbs(source) {
			source = filepath.Join(execCtx.WorkspaceRoot, source)
		}

		// DinD: Translate container path to host path for Docker daemon
		source = execCtx.TranslatePathForMount(source)

		// Format for Docker (Windows C:\path -> /c/path)
		mountStr := formatDockerVolume(source) + ":" + resolved.Target
		if resolved.ReadOnly {
			mountStr += ":ro"
		}
		args = append(args, "-v", mountStr)
	}

	// Environment variables - apply same precedence as executeContainer:
	// 1. Per-tool host-env (forwarded from host)
	// 2. Static tool.Env from YAML
	// 3. Per-call execCtx.EnvOverrides
	// Note: Global credentials are NOT included here as BuildCommand doesn't have access to executor state.
	// Callers needing credential forwarding should use Execute() instead.
	for _, name := range td.HostEnv {
		if val := os.Getenv(name); val != "" {
			args = append(args, "-e", name+"="+val)
		}
	}
	for k, v := range td.Env {
		args = append(args, "-e", k+"="+resolvePlaceholders(v, execCtx.Placeholders))
	}
	for k, v := range execCtx.EnvOverrides {
		args = append(args, "-e", k+"="+v)
	}

	// User
	if td.User != "" {
		args = append(args, "-u", td.User)
	}

	// Network
	if td.Network != "" {
		args = append(args, "--network", td.Network)
	}

	// Privileged
	if td.Privileged {
		args = append(args, "--privileged")
	}

	// Resource limits
	if td.Resources != nil {
		if td.Resources.Memory != "" {
			args = append(args, "--memory", td.Resources.Memory)
		}
		if td.Resources.CPUs > 0 {
			args = append(args, "--cpus", fmt.Sprintf("%.2f", float64(td.Resources.CPUs)))
		}
		if td.Resources.ShmSize != "" {
			args = append(args, "--shm-size", td.Resources.ShmSize)
		}
	}

	// Image reference
	imageRef := td.FullImage()
	if imageRef == "" && td.LocalPath != "" {
		imageRef = td.LocalImageTag()
	}
	if imageRef == "" {
		// No valid image reference - caller should validate before building command
		return nil
	}
	args = append(args, imageRef)

	// Container command (entrypoint is handled by image, command is passed as args)
	if len(td.Entrypoint) > 0 {
		// Override entrypoint
		args = append([]string{"run", "--rm", "-i", "--entrypoint", td.Entrypoint[0]}, args[3:]...)
		// Add remaining entrypoint args
		args = append(args, td.Entrypoint[1:]...)
	}
	args = append(args, td.Command...)
	args = append(args, execCtx.ArgsOverrides...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	SetProcessGroup(cmd)

	// Set environment for docker command itself (inherits from host)
	if execCtx.FullEnv != nil {
		cmd.Env = execCtx.FullEnv
	} else {
		cmd.Env = os.Environ()
	}

	return cmd
}
