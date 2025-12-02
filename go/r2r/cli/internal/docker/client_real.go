package docker

import (
	"context"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// RealDockerClient wraps the official Docker client to implement DockerClient interface
type RealDockerClient struct {
	*client.Client
}

// NewRealDockerClient creates a new RealDockerClient instance
func NewRealDockerClient(opts ...client.Opt) (DockerClient, error) {
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}
	return &RealDockerClient{Client: cli}, nil
}

// ImageInspect wraps the client ImageInspect to match our interface (without variadic options)
func (r *RealDockerClient) ImageInspect(ctx context.Context, imageID string) (image.InspectResponse, error) {
	return r.Client.ImageInspect(ctx, imageID)
}

// Ensure RealDockerClient implements DockerClient interface
var _ DockerClient = (*RealDockerClient)(nil)
