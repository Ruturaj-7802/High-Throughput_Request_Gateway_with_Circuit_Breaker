package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/ruturaj-7802/gateway/config"
)

var routeConfig config.Config

func main() {
	var err error

	routeConfig, err = config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	http.HandleFunc("/v1/proxy/", handleProxy)

	fmt.Println("Gateway running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

var rrCounter = make(map[string]int)

// tracks round-robin index for each service

func handleProxy(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	// path will be => /v1/proxy/{service_name}
	if len(parts) < 4 {
		http.Error(w, "Service name required ", http.StatusBadRequest)
		return
	}
	service := parts[3]

	backends, ok := routeConfig[service]
	if !ok || len(backends) == 0 {
		http.Error(w, "Service not found!", http.StatusNotFound)
		return
	}

	// round robin
	idx := rrCounter[service] % len(backends)
	target := backends[idx]
	rrCounter[service]++

	// forward GET to selected backend
	resp, err := http.Get(target)
	if err != nil {
		http.Error(w, "Upstream error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	// set response status to match backend response
	w.WriteHeader(resp.StatusCode)
	// copy backend response body to client
	_, _ = io.Copy(w, resp.Body)
}
