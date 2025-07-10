package metrics

import (
	"sync"
)

type BackendMetrics struct {
	TotalRequests int
	Successes     int
	Failures      int
}

var mu sync.Mutex
var stats = make(map[string]*BackendMetrics)

func InitBackend(url string) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := stats[url]; !ok {
		stats[url] = &BackendMetrics{}
	}
}

func RecordRequest(url string, success bool) {
	mu.Lock()
	defer mu.Unlock()
	if stat, ok := stats[url]; ok {
		stat.TotalRequests++
		if success {
			stat.Successes++
		} else {
			stat.Failures++
		}
	}
}

func GetAllStats() map[string]BackendMetrics {
	mu.Lock()
	defer mu.Unlock()

	// Return a copy to avoid data races
	copy := make(map[string]BackendMetrics)
	for k, v := range stats {
		copy[k] = *v
	}
	return copy
}
