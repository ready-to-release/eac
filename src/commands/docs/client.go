package docs

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// Client manages MkDocs container operations
type Client struct {
	docker *client.Client
	ctx    context.Context
}

// ContainerInfo contains information about the running MkDocs container
type ContainerInfo struct {
	Name string
	URL  string
	Port int
}

// NewClient creates a new docs client
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Test connection
	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("docker is not running: %w", err)
	}

	return &Client{
		docker: cli,
		ctx:    ctx,
	}, nil
}

// Close closes the Docker client connection
func (c *Client) Close() {
	if c.docker != nil {
		c.docker.Close()
	}
}

// IsRunning checks if the MkDocs container is already running
func (c *Client) IsRunning() (bool, *ContainerInfo, error) {
	return isContainerRunning(c.docker, c.ctx)
}

// StartContainer starts the MkDocs container
func (c *Client) StartContainer(port int) (*ContainerInfo, error) {
	return startMkDocsContainer(c.docker, c.ctx, port)
}

// StopContainer stops the MkDocs container
func (c *Client) StopContainer() error {
	return stopMkDocsContainer(c.docker, c.ctx)
}

// OpenBrowser opens the default web browser to the given URL
func (c *Client) OpenBrowser(url string) error {
	return openBrowser(url)
}

// StreamLogs streams container logs to stdout
func (c *Client) StreamLogs() error {
	return streamContainerLogs(c.docker, c.ctx)
}
