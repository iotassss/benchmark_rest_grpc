// client.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/benchmark/internal/benchmark"
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
	doRequest := func() error {
		return doHTTPRequest(client, targetURL)
	}

	benchmark.ExecBenchmark(
		doRequest,
		total,
		concurrency,
		warmup,
	)
}

func doHTTPRequest(client *http.Client, url string) error {
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
