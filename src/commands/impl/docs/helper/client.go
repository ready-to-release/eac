package docs

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/ready-to-release/eac/src/commands/internal/serve"
	"github.com/ready-to-release/eac/src/core/logging"
	"go.uber.org/zap"
)

// Client manages MkDocs container operations
type Client struct {
	docker *client.Client
	ctx    context.Context
	logger *logging.Logger
}

// NewClient creates a new docs client with logger
func NewClient(logger *logging.Logger) (*Client, error) {
	logger.Debug("Initializing Docker client")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Error("Failed to create Docker client", zap.Error(err))
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Test connection
	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		cli.Close()
		logger.Error("Docker daemon not responding", zap.Error(err))
		return nil, fmt.Errorf("docker is not running: %w", err)
	}

	logger.Debug("Docker client initialized successfully")
	return &Client{
		docker: cli,
		ctx:    ctx,
		logger: logger,
	}, nil
}

// Close closes the Docker client connection
func (c *Client) Close() {
	if c.docker != nil {
		c.logger.Debug("Closing Docker client")
		c.docker.Close()
	}
}

// IsRunning checks if the MkDocs container is already running
func (c *Client) IsRunning() (bool, *ContainerInfo, error) {
	c.logger.Debug("Checking if MkDocs container is running")
	return isContainerRunning(c.docker, c.ctx, c.logger)
}

// StartContainer starts the MkDocs container
func (c *Client) StartContainer(port int) (*ContainerInfo, error) {
	c.logger.Debug("Starting MkDocs container", zap.Int("port", port))
	return startMkDocsContainer(c.docker, c.ctx, port, c.logger)
}

// StopContainer stops the MkDocs container
func (c *Client) StopContainer() error {
	c.logger.Debug("Stopping MkDocs container")
	return stopMkDocsContainer(c.docker, c.ctx, c.logger)
}

// OpenBrowser opens the default web browser to the given URL.
// In DinD mode, this is a no-op and returns nil.
func (c *Client) OpenBrowser(url string) error {
	c.logger.Debug("Opening browser", zap.String("url", url))
	return openBrowser(url, c.logger)
}

// OpenBrowserWithFallback opens the browser and returns whether it was opened.
// In DinD mode, returns (false, nil) to indicate browser was skipped.
func (c *Client) OpenBrowserWithFallback(url string) (opened bool, err error) {
	c.logger.Debug("Attempting to open browser with fallback", zap.String("url", url))
	opened, err = serve.OpenBrowserWithFallback(url)
	if err != nil {
		c.logger.Debug("Browser opening failed", zap.Error(err))
	} else if !opened {
		c.logger.Debug("Browser opening skipped")
	} else {
		c.logger.Debug("Browser opened successfully")
	}
	return opened, err
}

// StreamLogs streams container logs to stdout
func (c *Client) StreamLogs() error {
	c.logger.Debug("Streaming container logs")
	return streamContainerLogs(c.docker, c.ctx, c.logger)
}
