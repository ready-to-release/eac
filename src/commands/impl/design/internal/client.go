package design

import (
	"context"
	"fmt"
)

// Client provides a high-level interface for Structurizr operations
type Client struct {
	containerMgr *ContainerManager
	ctx          context.Context
}

// NewClient creates a new Structurizr client
func NewClient() (*Client, error) {
	ctx := context.Background()

	containerMgr, err := NewContainerManager(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		containerMgr: containerMgr,
		ctx:          ctx,
	}, nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	if c.containerMgr != nil {
		return c.containerMgr.Close()
	}
	return nil
}

// ListModules returns all modules with architecture documentation
func (c *Client) ListModules() ([]ModuleInfo, error) {
	return ListAvailableModules()
}

// ValidateModule validates that a module exists and has required files
func (c *Client) ValidateModule(module string) error {
	return ValidateModule(module)
}

// StartContainer starts a Structurizr container for the module
func (c *Client) StartContainer(module string, port int) (*ContainerInfo, error) {
	// Validate module first
	err := ValidateModule(module)
	if err != nil {
		return nil, err
	}

	return c.containerMgr.StartContainer(module, port)
}

// StopContainer stops the Structurizr container for a module
func (c *Client) StopContainer(module string) error {
	return c.containerMgr.StopContainer(module)
}

// IsRunning checks if a container is running for the module
func (c *Client) IsRunning(module string) (bool, *ContainerInfo, error) {
	return c.containerMgr.IsRunning(module)
}

// OpenBrowser opens the Structurizr URL in the default browser
func (c *Client) OpenBrowser(url string) error {
	return OpenBrowser(url)
}

// GetModuleInfo returns detailed information about a module
func (c *Client) GetModuleInfo(module string) (*ModuleInfo, error) {
	modules, err := ListAvailableModules()
	if err != nil {
		return nil, err
	}

	for _, m := range modules {
		if m.Name == module {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("module '%s' not found", module)
}
