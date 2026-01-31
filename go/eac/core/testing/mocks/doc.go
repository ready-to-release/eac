// Package mocks provides mock implementations of port interfaces for testing.
//
// These mocks allow tests to be decoupled from concrete implementations,
// making tests faster and more focused. All mocks implement their
// corresponding port interfaces from github.com/ready-to-release/eac/contracts/eac-core-interfaces.
//
// Usage:
//
//	cfg := mocks.NewMockConfig().WithRepoRoot("/test/root")
//	registry := mocks.NewMockModuleRegistry().WithModule(
//	    mocks.NewMockModule("test-module").WithGoComponent("go/test"),
//	)
package mocks
