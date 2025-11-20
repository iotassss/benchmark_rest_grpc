package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

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

	// 単一コネクションを全 goroutine で共有
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

	// --- ウォームアップ（計測対象外） ---
	log.Printf("Warmup: %d requests...\n", warmup)
	for i := 0; i < warmup; i++ {
		if err := doEcho(client, time.Duration(timeoutMs)*time.Millisecond); err != nil {
			log.Printf("warmup error: %v\n", err)
		}
	}

	durations := make([]time.Duration, total)
	var idx int64
	var errors int64

	log.Printf("Benchmark: %d requests, %d concurrency\n", total, concurrency)

	start := time.Now()

	var wg sync.WaitGroup
	requestsPerWorker := total / concurrency
	extra := total % concurrency

	for w := 0; w < concurrency; w++ {
		wg.Add(1)

		nReq := requestsPerWorker
		if w < extra {
			nReq++
		}

		go func(num int) {
			defer wg.Done()
			for i := 0; i < num; i++ {
				s := time.Now()
				if err := doEcho(client, time.Duration(timeoutMs)*time.Millisecond); err != nil {
					atomic.AddInt64(&errors, 1)
					continue
				}
				d := time.Since(s)

				pos := atomic.AddInt64(&idx, 1) - 1
				if int(pos) < len(durations) {
					durations[pos] = d
				}
			}
		}(nReq)
	}

	wg.Wait()
	elapsed := time.Since(start)

	valid := durations[:idx]
	if len(valid) == 0 {
		log.Fatal("no successful requests")
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i] < valid[j] })

	totalRequests := len(valid)
	qps := float64(totalRequests) / elapsed.Seconds()

	p50 := percentile(valid, 0.50)
	p95 := percentile(valid, 0.95)
	p99 := percentile(valid, 0.99)

	fmt.Printf("Total requests (success): %d\n", totalRequests)
	fmt.Printf("Errors: %d\n", errors)
	fmt.Printf("Total time: %v\n", elapsed)
	fmt.Printf("QPS: %.2f\n", qps)
	fmt.Printf("Latency p50: %v\n", p50)
	fmt.Printf("Latency p95: %v\n", p95)
	fmt.Printf("Latency p99: %v\n", p99)
}

func doEcho(client pb.EchoServiceClient, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := client.Echo(ctx, &pb.EchoRequest{Message: "hello"})
	return err
}

func percentile(durs []time.Duration, p float64) time.Duration {
	if len(durs) == 0 {
		return 0
	}
	if p <= 0 {
		return durs[0]
	}
	if p >= 1 {
		return durs[len(durs)-1]
	}
	idx := int(math.Ceil(float64(len(durs)) * p))
	if idx <= 0 {
		idx = 1
	}
	if idx > len(durs) {
		idx = len(durs)
	}
	return durs[idx-1]
}
