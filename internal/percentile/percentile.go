package percentile

import (
	"time"
)

type BenchmarkConfig struct {
	Total       int
	Concurrency int
	Warmup      int
	Timeout     time.Duration
}

type Result struct {
	TotalRequests int
	Errors        int64
	TotalTime     time.Duration
	QPS           float64
	P50           time.Duration
	P95           time.Duration
	P99           time.Duration
}

// Percentile helper assumes durs is already sorted ascending.
func Percentile(durs []time.Duration, p float64) time.Duration {
	if len(durs) == 0 {
		return 0
	}
	if p <= 0 {
		return durs[0]
	}
	if p >= 1 {
		return durs[len(durs)-1]
	}
	idx := int(float64(len(durs)) * p)
	if idx <= 0 {
		idx = 1
	}
	if idx > len(durs) {
		idx = len(durs)
	}
	return durs[idx-1]
}
