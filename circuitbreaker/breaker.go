package circuitbreaker

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type State string

const (
	Closed   State = "Closed"
	Open     State = "Open"
	HalfOpen State = "Half-Open"
)

const (
	failureThreshold = 3
	openTimeout      = 10 * time.Second
)

type Circuitbreaker struct {
	mu           sync.Mutex
	FailureCount int
	SuccessCount int
	TotalCount   int
	State        State
	LastFailure  time.Time
	ServiceURL   string // For logging
}

// Closed => working fine
// Open => failed for N consecutive times
// Half-Open => after Timeout, one request is send to Open circuit, if succeeds, -> Closed, else -> Open and rest Timeout timer

func NewBreaker(service string) *Circuitbreaker {
	return &Circuitbreaker{
		State:      Closed,
		ServiceURL: service,
	}
}

// check if the request is allowed
func (cb *Circuitbreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.State {
	case Open:
		if time.Since(cb.LastFailure) > openTimeout {
			cb.State = HalfOpen
			log.Printf("[CIRCUIT] %s → HALF-OPEN (timeout expired)", cb.ServiceURL) // 🔥
			return true
		}
		log.Printf("[CIRCUIT] %s → BLOCKED (OPEN)", cb.ServiceURL)
		return false
	default:
		return true
	}
}

func (cb *Circuitbreaker) PrintSummary() {
	fmt.Printf("\n[CIRCUIT STATUS] %s\n", cb.ServiceURL)
	fmt.Printf("  Total:    %d\n", cb.TotalCount)
	fmt.Printf("  Success:  %d\n", cb.SuccessCount)
	fmt.Printf("  Failures: %d\n", cb.FailureCount)
	fmt.Printf("  State:    %s\n\n", cb.State)
}

// record result of request
func (cb *Circuitbreaker) RecordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if success {
		cb.FailureCount = 0
		cb.SuccessCount++
		if cb.State != Closed {
			cb.State = Closed
			log.Printf("[CIRCUIT] %s → CLOSED (recovered)", cb.ServiceURL)
		}
	} else {
		cb.FailureCount++
		cb.LastFailure = time.Now()

		log.Printf("[CIRCUIT] %s → Failure #%d", cb.ServiceURL, cb.FailureCount)

		if cb.FailureCount >= failureThreshold {
			cb.State = Open
			log.Printf("[CIRCUIT] %s → OPEN (threshold reached)", cb.ServiceURL)
		}
	}

	// Print summary
	cb.PrintSummary()
}
