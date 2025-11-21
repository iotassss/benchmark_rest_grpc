// client.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type EchoRequest struct {
	Message string `json:"message"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

func main() {
	var (
		targetURL   string
		total       int
		concurrency int
		warmup      int
	)

	flag.StringVar(&targetURL, "url", "http://localhost:8080/echo", "target REST endpoint")
	flag.IntVar(&total, "n", 1000, "total number of benchmark requests")
	flag.IntVar(&concurrency, "c", 10, "number of concurrent goroutines")
	flag.IntVar(&warmup, "warmup", 100, "number of warmup requests (not measured)")
	flag.Parse()

	if total <= 0 || concurrency <= 0 {
		log.Fatal("n and c must be > 0")
	}
	if concurrency > total {
		concurrency = total
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        concurrency * 2,
			MaxIdleConnsPerHost: concurrency * 2,
		},
	}

	// --- ウォームアップ（計測対象外） ---
	log.Printf("Warmup: %d requests...\n", warmup)
	for i := 0; i < warmup; i++ {
		if err := doRequest(client, targetURL); err != nil {
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
				if err := doRequest(client, targetURL); err != nil {
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

	// エラーがあった場合、0値を除いた範囲だけを見る
	valid := durations[:idx]
	sort.Slice(valid, func(i, j int) bool { return valid[i] < valid[j] })

	totalRequests := len(valid)
	if totalRequests == 0 {
		log.Fatal("no successful requests")
	}

	qps := float64(totalRequests) / elapsed.Seconds()
	p50 := valid[int(float64(totalRequests)*0.50)-1]
	p95 := valid[int(float64(totalRequests)*0.95)-1]
	p99 := valid[int(float64(totalRequests)*0.99)-1]

	fmt.Printf("Total requests (success): %d\n", totalRequests)
	fmt.Printf("Errors: %d\n", errors)
	fmt.Printf("Total time: %v\n", elapsed)
	fmt.Printf("QPS: %.2f\n", qps)
	fmt.Printf("Latency p50: %v\n", p50)
	fmt.Printf("Latency p95: %v\n", p95)
	fmt.Printf("Latency p99: %v\n", p99)
}

func doRequest(client *http.Client, url string) error {
	reqBody, _ := json.Marshal(EchoRequest{Message: "hello"})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	// decode もオーバーヘッドに含める
	var out EchoResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	return nil
}
