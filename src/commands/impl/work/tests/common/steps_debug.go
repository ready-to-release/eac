package common

import (
	"fmt"
	"os"
	"path/filepath"
)

// Debug mode steps

func DebugLogsAreWrittenTo(ctx *TestContext, logPath string) error {
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		os.MkdirAll(logPath, 0755)
	}
	debugFile := filepath.Join(logPath, "test-debug.log")
	ctx.DebugFiles = append(ctx.DebugFiles, debugFile)
	return nil
}

func DebugLogsContainWorkspaceDetails(ctx *TestContext) error {
	return debugLogsContainContent(ctx, "workspace")
}

func DebugLogsContainWorkspaceRemovalDetails(ctx *TestContext) error {
	return debugLogsContainContent(ctx, "removal")
}

func DebugLogsContainBranchInformation(ctx *TestContext) error {
	return debugLogsContainContent(ctx, "branch")
}

func DebugLogsContainRebaseDetails(ctx *TestContext) error {
	return debugLogsContainContent(ctx, "rebase")
}

func DebugLogsContainMergeDetails(ctx *TestContext) error {
	return debugLogsContainContent(ctx, "merge")
}

func debugLogsContainContent(ctx *TestContext, keyword string) error {
	if len(ctx.DebugFiles) == 0 {
		for _, arg := range ctx.Args {
			if arg == "--debug" || arg == "-d" {
				return nil
			}
		}
		return fmt.Errorf("no debug logs found")
	}
	return nil
}
