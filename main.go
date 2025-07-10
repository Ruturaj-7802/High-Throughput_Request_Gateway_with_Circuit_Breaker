package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	cb "github.com/ruturaj-7802/gateway/circuitbreaker"
	"github.com/ruturaj-7802/gateway/config"
	"github.com/ruturaj-7802/gateway/metrics"
)

var routeConfig config.Config

// tracks round-robin index for each service
var rrCounter = make(map[string]int)

// circuit breaker map per backend service
var breakers = make(map[string]*cb.Circuitbreaker)

func main() {
	var err error

	routeConfig, err = config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// initialize breaker for each backend
	for _, backends := range routeConfig {
		for _, backend := range backends {
			breakers[backend] = cb.NewBreaker(backend)
			metrics.InitBackend(backend)
			log.Printf("Initialized circuit breaker for backend: %s", backend)
		}
	}

	http.HandleFunc("/v1/proxy/", handleProxy)
	http.HandleFunc("/metrics", handleMetrics)

	fmt.Println("Gateway running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	// path will be => /v1/proxy/{service_name}
	if len(parts) < 4 {
		http.Error(w, "Service name required ", http.StatusBadRequest)
		return
	}
	service := parts[3]

	fmt.Println("Incoming service:", service)

	backends, ok := routeConfig[service]
	if !ok || len(backends) == 0 {
		log.Printf("Service '%s' not found in route config", service)
		http.Error(w, "Service not found!", http.StatusNotFound)
		return
	}

	// log.Printf("Available backends for %s: %v", service, backends)

	total := len(backends)
	for i := 0; i < total; i++ {

		// round robin
		idx := rrCounter[service] % len(backends)
		target := backends[idx]
		rrCounter[service]++

		breaker, exists := breakers[target]
		if !exists {
			log.Printf("No circuit breaker found for backend %s", target)
			continue
		}

		if !breaker.AllowRequest() {
			// log.Printf("Skipping %s (circuit open)", target)
			continue
		}
		// log.Printf("Circuit breaker CLOSED for %s - proceeding", target)
		client := http.Client{Timeout: 5 * time.Second}

		// forward GET to selected backend
		resp, err := client.Get(target)

		if err != nil {
			log.Printf("Request to %s failed after %v: %v", target, err)
			breaker.RecordResult(false)
			metrics.RecordRequest(target, false)
			continue
		}

		if resp.StatusCode >= 500 {
			log.Printf("Request to %s failed with status code: %d (duration: %v)", target, resp.StatusCode)
			breaker.RecordResult(false)
			metrics.RecordRequest(target, false)
			resp.Body.Close()
			continue
		}

		breaker.RecordResult(true)
		metrics.RecordRequest(target, true)

		defer resp.Body.Close()
		// set response status to match backend response
		w.WriteHeader(resp.StatusCode)
		// copy backend response body to client
		_, _ = io.Copy(w, resp.Body)
		log.Printf("########################################################################")
		return
	}

	log.Printf("All backends failed for service: %s", service)
	http.Error(w, "All backends unavailable", http.StatusServiceUnavailable)
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	for backend, stat := range metrics.GetAllStats() {
		fmt.Println(w, "%s -> total=%d, success=%d, failure=%d\n", backend, stat.TotalRequests, stat.Successes, stat.Failures)
	}
}
