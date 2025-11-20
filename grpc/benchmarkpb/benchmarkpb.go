package benchmarkpb

import (
	"context"
)

// Minimal types to satisfy imports in example code.
type EchoRequest struct {
	Message string `json:"message"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

// Client API
type EchoServiceClient interface {
	Echo(ctx context.Context, in *EchoRequest, opts ...interface{}) (*EchoResponse, error)
}

func NewEchoServiceClient(conn interface{}) EchoServiceClient {
	return &localEchoClient{}
}

type localEchoClient struct{}

func (c *localEchoClient) Echo(ctx context.Context, in *EchoRequest, opts ...interface{}) (*EchoResponse, error) {
	return &EchoResponse{Message: "ok"}, nil
}

// Server API
type EchoServiceServer interface {
	Echo(context.Context, *EchoRequest) (*EchoResponse, error)
}

type UnimplementedEchoServiceServer struct{}

func (*UnimplementedEchoServiceServer) Echo(context.Context, *EchoRequest) (*EchoResponse, error) {
	return nil, nil
}

// Register stub
func RegisterEchoServiceServer(s interface{}, srv EchoServiceServer) {}
