package client

import (
	pb "github.com/backtesting-org/kronos-sdk/grpc/gen/inference"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Module provides the inference client for ML predictions.
// Users must provide their own gRPC connection configuration.
var Module = fx.Module("inference-client",
	fx.Provide(
		fx.Annotate(
			NewClient,
			fx.As(new(InferenceClient)),
		),
	),
)

// ConnectionParams allows users to provide custom gRPC connection settings.
type ConnectionParams struct {
	// Address is the gRPC server address (e.g., "localhost:50051")
	Address string
	// Options are custom gRPC dial options (optional)
	Options []grpc.DialOption
}

// NewGRPCClient creates a gRPC client connection to the inference server.
// This is a helper function that users can use in their fx.Provide.
func NewGRPCClient(params ConnectionParams) (pb.ModelInferenceClient, error) {
	// Default options if none provided
	opts := params.Options
	if opts == nil {
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
	}

	// Connect to the gRPC server
	conn, err := grpc.NewClient(params.Address, opts...)
	if err != nil {
		return nil, err
	}

	// Create the gRPC client
	return pb.NewModelInferenceClient(conn), nil
}
