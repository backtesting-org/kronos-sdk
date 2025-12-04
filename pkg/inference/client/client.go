package client

import (
	"context"

	pb "github.com/backtesting-org/kronos-sdk/grpc/gen/inference"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InferenceClient provides ML inference capabilities by communicating with
// a user's external inference server via gRPC.
type InferenceClient interface {
	// Predict sends features to the inference server and returns the confidence score.
	// Confidence is typically in range 0-100, where higher values indicate stronger signals.
	Predict(ctx context.Context, asset portfolio.Asset, features map[string]float64) (float64, error)
}

// Client is the default implementation of InferenceClient.
// It wraps the generated gRPC client and provides a convenient interface.
type Client struct {
	grpcClient pb.ModelInferenceClient
}

// NewClient creates a new inference client from an endpoint address.
// Example: client.NewClient("localhost:50051")
func NewClient(endpoint string) (*Client, error) {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		grpcClient: pb.NewModelInferenceClient(conn),
	}, nil
}

// NewClientFromGRPC creates a new inference client from an existing gRPC client.
// This is useful for advanced users who want custom gRPC options.
func NewClientFromGRPC(grpcClient pb.ModelInferenceClient) *Client {
	return &Client{
		grpcClient: grpcClient,
	}
}

// Predict sends the feature map to the user's inference server and returns the confidence score.
func (c *Client) Predict(ctx context.Context, asset portfolio.Asset, features map[string]float64) (float64, error) {
	// Convert feature map to protobuf message
	req := &pb.Features{
		NamedFeatures: features,
	}

	// Call the user's inference server via gRPC
	resp, err := c.grpcClient.Predict(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.Confidence, nil
}
