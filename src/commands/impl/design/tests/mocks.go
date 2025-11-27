// Package tests provides mock utilities for design command tests.
//
// This file contains Docker mocking and other test doubles for integration testing.
package tests

import (
	"fmt"
	"os/exec"
)

// MockDockerClient provides test doubles for Docker operations.
type MockDockerClient struct {
	isRunning      bool
	containers     map[string]*MockContainer
	shouldFailRun  bool
	shouldFailStop bool
}

// MockContainer represents a mock Docker container.
type MockContainer struct {
	ID     string
	Name   string
	Port   int
	Status string
}

// globalMockDocker is the shared mock Docker client for tests.
var globalMockDocker *MockDockerClient

// InitMockDocker initializes the global mock Docker client.
func InitMockDocker() {
	globalMockDocker = &MockDockerClient{
		isRunning:  true,
		containers: make(map[string]*MockContainer),
	}
}

// ResetMockDocker resets the mock Docker client state.
func ResetMockDocker() {
	if globalMockDocker != nil {
		globalMockDocker.containers = make(map[string]*MockContainer)
		globalMockDocker.isRunning = true
		globalMockDocker.shouldFailRun = false
		globalMockDocker.shouldFailStop = false
	}
}

// SetDockerRunning sets whether Docker should appear to be running.
func SetDockerRunning(running bool) {
	if globalMockDocker == nil {
		InitMockDocker()
	}
	globalMockDocker.isRunning = running
}

// SetDockerShouldFailRun makes Docker run commands fail.
func SetDockerShouldFailRun(shouldFail bool) {
	if globalMockDocker == nil {
		InitMockDocker()
	}
	globalMockDocker.shouldFailRun = shouldFail
}

// AddMockContainer adds a mock container to the Docker client.
func AddMockContainer(id, name string, port int) {
	if globalMockDocker == nil {
		InitMockDocker()
	}
	globalMockDocker.containers[id] = &MockContainer{
		ID:     id,
		Name:   name,
		Port:   port,
		Status: "running",
	}
}

// GetMockContainer retrieves a mock container by ID.
func GetMockContainer(id string) *MockContainer {
	if globalMockDocker == nil {
		return nil
	}
	return globalMockDocker.containers[id]
}

// IsDockerAvailable checks if docker command is available.
// This is used in tests to determine if we should skip Docker-dependent tests.
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	err := cmd.Run()
	return err == nil
}

// SkipIfDockerUnavailable returns an error if Docker is not available.
// Use this in test steps that require Docker.
func SkipIfDockerUnavailable() error {
	if !IsDockerAvailable() {
		return fmt.Errorf("Docker is not available - skipping test")
	}
	return nil
}
