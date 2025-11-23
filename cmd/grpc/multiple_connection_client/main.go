package main

import (
	"context"
	"flag"
	"log"
	"time"

	"example.com/benchmark/internal/benchmark"

	pb "example.com/benchmark/benchmarkpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var (
		addr        string
		total       int
		concurrency int
		warmup      int
		timeoutMs   int
	)

	flag.StringVar(&addr, "addr", "localhost:50051", "gRPC server address")
	flag.IntVar(&total, "n", 1000, "total number of benchmark requests")
	flag.IntVar(&concurrency, "c", 10, "number of concurrent goroutines")
	flag.IntVar(&warmup, "warmup", 100, "number of warmup requests (not measured)")
	flag.IntVar(&timeoutMs, "timeout", 5000, "per-request timeout in ms")
	flag.Parse()

	if total <= 0 || concurrency <= 0 {
		log.Fatal("n and c must be > 0")
	}
	if concurrency > total {
		concurrency = total
	}

	doRequest := func() error {
		return doEcho(addr, time.Duration(timeoutMs)*time.Millisecond)
	}

	benchmark.ExecBenchmark(
		doRequest,
		total,
		concurrency,
		warmup,
	)
}

func doEcho(addr string, timeout time.Duration) error {
	// リクエストごとに新しいコネクションを作成
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewEchoServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = client.Echo(ctx, &pb.EchoRequest{Message: "hello"})
	return err
}
