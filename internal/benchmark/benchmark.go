package benchmark

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func ExecBenchmark(
	// サーバーにリクエストを送る関数
	doRequest func() error,
	total,
	concurrency,
	warmup int,
	// ) (latencies []time.Duration, errorsCount int64, elapsed time.Duration) {
) {
	// --- ウォームアップ（計測対象外） ---
	log.Printf("Warmup: %d requests...\n", warmup)
	for i := 0; i < warmup; i++ {
		if err := doRequest(); err != nil {
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

		// 端数を前の worker に配る
		nReq := requestsPerWorker
		if w < extra {
			nReq++
		}

		go func(num int) {
			defer wg.Done()
			for i := 0; i < num; i++ {
				s := time.Now()
				if err := doRequest(); err != nil {
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

	// durationsは全リクエスト件数分があらかじめ予約されているため、失敗分は0値のまま後半に残る
	// エラーがあった場合、0値を除いた範囲だけを見る
	valid := durations[:idx]
	sort.Slice(valid, func(i, j int) bool { return valid[i] < valid[j] })

	PrintSummary(errors, elapsed, valid)
}

func PrintSummary(errors int64, elapsed time.Duration, latencies []time.Duration) {
	totalRequests := len(latencies)
	if totalRequests == 0 {
		log.Fatal("no successful requests")
	}

	qps := float64(totalRequests) / elapsed.Seconds()
	p50 := latencies[int(float64(totalRequests)*0.50)-1]
	p95 := latencies[int(float64(totalRequests)*0.95)-1]
	p99 := latencies[int(float64(totalRequests)*0.99)-1]

	fmt.Printf("Total requests (success): %d\n", totalRequests)
	fmt.Printf("Errors: %d\n", errors)
	fmt.Printf("Total time: %v\n", elapsed)
	fmt.Printf("QPS: %.2f\n", qps)
	fmt.Printf("Latency p50: %v\n", p50)
	fmt.Printf("Latency p95: %v\n", p95)
	fmt.Printf("Latency p99: %v\n", p99)
}
