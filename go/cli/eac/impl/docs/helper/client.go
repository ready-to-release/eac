package docs

import (
	"context"

	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

// Client manages MkDocs container operations.
type Client struct {
	docker docker.DockerClient
	ctx    context.Context
}

// NewClient creates a new docs client.
func NewClient() (*Client, error) {
	log.Debugf("Initializing Docker client")

	cli, err := docker.NewDockerClient()
	if err != nil {
		log.Errorf("Failed to create Docker client: error=%v", err)
		return nil, err
	}

	log.Debugf("Docker client initialized successfully")
	return &Client{
		docker: cli,
		ctx:    context.Background(),
	}, nil
}

// Close closes the Docker client connection.
func (c *Client) Close() {
	if c.docker != nil {
		log.Debugf("Closing Docker client")
		c.docker.Close()
	}
}

// IsRunning checks if the MkDocs container is already running.
func (c *Client) IsRunning() (bool, *ContainerInfo, error) {
	log.Debugf("Checking if MkDocs container is running")
	return isContainerRunning(c.docker, c.ctx)
}

// StartContainer starts the MkDocs container.
func (c *Client) StartContainer(port int) (*ContainerInfo, error) {
	log.Debugf("Starting MkDocs container: port=%d", port)
	return startMkDocsContainer(c.docker, c.ctx, port)
}

// StopContainer stops the MkDocs container.
func (c *Client) StopContainer() error {
	log.Debugf("Stopping MkDocs container")
	return stopMkDocsContainer(c.docker, c.ctx)
}

// OpenBrowser opens the default web browser to the given URL.
// In DinD mode, this is a no-op and returns nil.
func (c *Client) OpenBrowser(url string) error {
	log.Debugf("Opening browser: url=%s", url)
	return openBrowser(url)
}

// OpenBrowserWithFallback opens the browser and returns whether it was opened.
// In DinD mode, returns (false, nil) to indicate browser was skipped.
func (c *Client) OpenBrowserWithFallback(url string) (opened bool, err error) {
	log.Debugf("Attempting to open browser with fallback: url=%s", url)
	opened, err = docker.OpenBrowserWithFallback(url)
	if err != nil {
		log.Debugf("Browser opening failed: error=%v", err)
	} else if !opened {
		log.Debugf("Browser opening skipped")
	} else {
		log.Debugf("Browser opened successfully")
	}
	return opened, err
}

// StreamLogs streams container logs to stdout.
func (c *Client) StreamLogs() error {
	log.Debugf("Streaming container logs")
	return streamContainerLogs(c.docker, c.ctx)
}
