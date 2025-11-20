package main

import (
	"context"
	"log"
	"net"

	pb "example.com/benchmark/benchmarkpb"
	"google.golang.org/grpc"
)

type echoServer struct {
	pb.UnimplementedEchoServiceServer
}

func (s *echoServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	// シンプルなビジネスロジック（RESTと揃える）
	_ = req // ここで何か処理したい場合は追加

	return &pb.EchoResponse{
		Message: "ok",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterEchoServiceServer(grpcServer, &echoServer{})

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
