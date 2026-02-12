// Package specs contains godog step implementations for core features.
//
// This file contains CI mocking functions used to simulate CI pipeline
// status in cache invalidation tests.
package specs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

func runCommandWithMockedCI(ctx *eacgodog.TestContext, cmdLine string) error {
	ctx.MustBeIsolated()

	mockJSON, err := json.Marshal(cacheCtx.mockedCIStatus)
	if err != nil {
		return fmt.Errorf("failed to marshal mock CI status: %w", err)
	}

	mockDir := filepath.Join(ctx.IsolatedDir, "out", "test-mocks")
	if err := os.MkdirAll(mockDir, 0o755); err != nil {
		return fmt.Errorf("failed to create mock dir: %w", err)
	}

	mockPath := filepath.Join(mockDir, "ci-status.json")
	if err := os.WriteFile(mockPath, mockJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write mock file: %w", err)
	}

	ctx.SetMockOverride("CLIE_MOCK_CI_STATUS", mockPath)
	ctx.SetMockOverride("CLIE_MOCK_HEAD_SHA", cacheCtx.currentHeadSHA)
	ctx.SetMockOverride("CLIE_MOCK_CHANGED_FILES", strings.Join(cacheCtx.changedFiles, ","))

	if err := ctx.RunCommand(cmdLine); err != nil {
		return err
	}

	parseYAMLOutput(ctx)

	return nil
}

func setupMockedCIStatus(ctx *eacgodog.TestContext, table *godog.Table) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	for i, row := range table.Rows {
		if i == 0 {
			continue
		}
		if len(row.Cells) < 3 {
			continue
		}

		module := row.Cells[0].Value
		lastSuccessSHA := row.Cells[1].Value
		hasFilesChanged, _ := strconv.ParseBool(row.Cells[2].Value)

		cacheCtx.mockedCIStatus[module] = mockedModuleCI{
			LastSuccessSHA:  lastSuccessSHA,
			HasFilesChanged: hasFilesChanged,
		}
	}

	return nil
}

func mockValidCIAtHead(ctx *eacgodog.TestContext, module string) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	cacheCtx.mockedCIStatus[module] = mockedModuleCI{
		LastSuccessSHA:   cacheCtx.currentHeadSHA,
		HasFilesChanged:  false,
		HasValidCIAtHead: true,
	}
	return nil
}

func mockNoCIHistory(ctx *eacgodog.TestContext, module string) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	cacheCtx.mockedCIStatus[module] = mockedModuleCI{
		NoHistory: true,
	}
	return nil
}

func mockCIAtDifferentSHA(ctx *eacgodog.TestContext, module string) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	cacheCtx.mockedCIStatus[module] = mockedModuleCI{
		LastSuccessSHA:  "different-sha-123456",
		HasFilesChanged: true,
	}
	return nil
}
