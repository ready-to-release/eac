package godog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CommandLookupFunc finds a command function by its space-separated name (e.g., "show help").
// Returns the command function and true if found, nil and false otherwise.
type CommandLookupFunc func(cmdName string) (func() int, bool)

// osMu serializes access to os.Args, os.Stdout, os.Stderr, and env vars during
// in-process command dispatch. godog runs scenarios sequentially so this is safe,
// but the mutex protects against any future concurrency.
var osMu sync.Mutex

// ExitCodeDispatchDeclined is returned by the in-process dispatcher when it
// cannot handle a command in-process. This happens when:
//   - The command is not registered (e.g., circular dependency prevents import)
//   - The command requests --help (which requires cobra in a subprocess)
//
// RunCommand checks for this sentinel and falls back to subprocess execution.
const ExitCodeDispatchDeclined = -1

// MakeInProcessDispatcher returns a CommandDispatcher that executes commands
// in-process via a command lookup function instead of spawning a subprocess.
// This avoids subprocess overhead and makes BDD tests run much faster.
//
// The lookupFn is called with a space-separated command name (e.g., "show help")
// and should return the command function and whether it was found.
// Typically this wraps registry.Global().Get() from the clibase/registry package.
func MakeInProcessDispatcher(ctx *TestContext, lookupFn CommandLookupFunc) func(args []string) (string, int) {
	return func(args []string) (string, int) {
		if len(args) < 1 {
			return "in-process dispatch: need at least 1 arg", 1
		}

		// Decline --help requests: cobra handles these in subprocess mode
		// but in-process commands don't have cobra's help infrastructure.
		for _, arg := range args {
			if arg == "--help" || arg == "-h" {
				return "", ExitCodeDispatchDeclined
			}
		}

		// Try longest match first (e.g., "templates install docs" before "templates install")
		// then fall back to shorter matches (e.g., "init" for single-word commands).
		// This mirrors the CLI's dispatch behavior in main.go.
		var cmdFn func() int
		var ok bool
		for i := len(args); i >= 1; i-- {
			candidate := args[0]
			for j := 1; j < i; j++ {
				if strings.HasPrefix(args[j], "-") {
					break // Flags are not part of the command name
				}
				candidate += " " + args[j]
			}
			cmdFn, ok = lookupFn(candidate)
			if ok {
				break
			}
		}
		if !ok {
			return "", ExitCodeDispatchDeclined
		}

		osMu.Lock()
		defer osMu.Unlock()

		// Save original state
		origArgs := os.Args
		origStdout := os.Stdout
		origStderr := os.Stderr
		origRepoRoot := os.Getenv("CLIE_REPO_ROOT")
		origContainerRoot := os.Getenv("CLIE_CONTAINER_ROOT")
		origTestScope := os.Getenv("CLIE_TEST_SCOPE")
		origPWD := os.Getenv("CLIE_PWD")
		origWd, _ := os.Getwd()

		// Restore everything on exit
		defer func() {
			os.Args = origArgs
			os.Stdout = origStdout
			os.Stderr = origStderr
			os.Setenv("CLIE_REPO_ROOT", origRepoRoot)
			os.Setenv("CLIE_CONTAINER_ROOT", origContainerRoot)
			os.Setenv("CLIE_TEST_SCOPE", origTestScope)
			os.Setenv("CLIE_PWD", origPWD)
			os.Chdir(origWd)
		}()

		// Set os.Args as the command expects (["eac", "get", "test-results", ...])
		os.Args = append([]string{"eac"}, args...)

		// Set environment for isolated test
		if ctx.IsolatedDir != "" {
			os.Setenv("CLIE_REPO_ROOT", ctx.IsolatedDir)
			os.Setenv("CLIE_TEST_SCOPE", "1")
			// Change working directory so commands that resolve relative paths
			// (e.g., validate specs) operate within the isolated directory.
			workDir := ctx.IsolatedDir
			if ctx.CurrentWorkDir != "" {
				workDir = ctx.CurrentWorkDir
			}
			// Set CLIE_PWD for relative path resolution within the isolated directory.
			// Mirrors buildIsolationEnvironment() for subprocess execution.
			// Without it, commands fall back to os.Getwd() which may differ on Windows.
			os.Setenv("CLIE_PWD", workDir)
			os.Chdir(workDir)
		}
		if ctx.OriginalRepoRoot != "" {
			os.Setenv("CLIE_CONTAINER_ROOT", ctx.OriginalRepoRoot)
		}

		// Apply mock environment variables so in-process commands use mocks
		// (e.g., CLIE_MOCK_DOCKER, CLIE_MOCK_SECURITY) just like subprocess mode.
		restoreMockEnv := applyMockEnvironment(ctx)
		defer restoreMockEnv()

		// Capture stdout and stderr into a combined buffer
		stdoutR, stdoutW, err := os.Pipe()
		if err != nil {
			return fmt.Sprintf("failed to create stdout pipe: %v", err), 1
		}
		stderrR, stderrW, err := os.Pipe()
		if err != nil {
			stdoutR.Close()
			stdoutW.Close()
			return fmt.Sprintf("failed to create stderr pipe: %v", err), 1
		}

		os.Stdout = stdoutW
		os.Stderr = stderrW

		// Redirect the zap logger to write to the stderr pipe so that commands
		// using logging.C().Info/Infof have their output captured.
		// Without this, zap writes to the original stderr fd, bypassing our pipe.
		restoreLogger := redirectLoggerToWriter(stderrW)
		defer restoreLogger()

		// Read pipes in background goroutines (separate buffers to avoid data race)
		var stdoutBuf, stderrBuf bytes.Buffer
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			io.Copy(&stdoutBuf, stdoutR)
		}()
		go func() {
			defer wg.Done()
			io.Copy(&stderrBuf, stderrR)
		}()

		// Execute the command
		exitCode := cmdFn()

		// Close write ends to unblock readers, then wait for goroutines
		stdoutW.Close()
		stderrW.Close()
		wg.Wait()
		stdoutR.Close()
		stderrR.Close()

		// Combine stdout and stderr (stdout first for deterministic output)
		stdoutBuf.Write(stderrBuf.Bytes())
		return stdoutBuf.String(), exitCode
	}
}

// applyMockEnvironment loads testing-mocks.yml and sets mock environment variables
// (e.g., CLIE_MOCK_DOCKER=true) on the current process. Returns a cleanup function
// that restores original values. This mirrors buildMockingEnvironment() for subprocess
// execution, ensuring in-process commands see the same mock configuration.
func applyMockEnvironment(ctx *TestContext) func() {
	if ctx.OriginalRepoRoot == "" {
		return func() {} // No repo root, nothing to configure
	}

	mockConfigPath := filepath.Join(ctx.OriginalRepoRoot, ".eac", "testing-mocks.yml")
	mockConfig, err := config.LoadTestingMocks(mockConfigPath)
	if err != nil {
		return func() {} // Config not found or invalid, skip mocking
	}

	mockEnvVars := mockConfig.ToEnvironmentVariables()
	if len(mockEnvVars) == 0 {
		return func() {}
	}

	// Also include per-scenario overrides
	for key, value := range ctx.MockOverrides {
		mockEnvVars = append(mockEnvVars, fmt.Sprintf("%s=%s", key, value))
	}

	// Save originals and set new values
	origValues := make(map[string]string, len(mockEnvVars))
	origExists := make(map[string]bool, len(mockEnvVars))
	for _, envVar := range mockEnvVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		if orig, exists := os.LookupEnv(key); exists {
			origValues[key] = orig
			origExists[key] = true
		}
		os.Setenv(key, value)
	}

	return func() {
		for _, envVar := range mockEnvVars {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := parts[0]
			if origExists[key] {
				os.Setenv(key, origValues[key])
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

// redirectLoggerToWriter replaces the global zap logger with one that writes to w.
// Returns a cleanup function that restores the original logger.
func redirectLoggerToWriter(w *os.File) func() {
	// Create a minimal console encoder (message only, matching the command output style)
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:       zapcore.OmitKey,
		LevelKey:      zapcore.OmitKey,
		NameKey:       zapcore.OmitKey,
		CallerKey:     zapcore.OmitKey,
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: zapcore.OmitKey,
		LineEnding:    zapcore.DefaultLineEnding,
	}
	encoder := zapcore.NewConsoleEncoder(encoderCfg)
	core := zapcore.NewCore(encoder, zapcore.AddSync(w), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	return logging.SetTestLogger(zapLogger)
}
