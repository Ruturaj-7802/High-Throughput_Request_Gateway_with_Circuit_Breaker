package circuitbreaker

import (
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
	State        State
	LastFailure  time.Time
	service      string
}

// Closed => working fine
// Open => failed for N consecutive times
// Half-Open => after Timeout, one request is send to Open circuit, if succeeds, -> Closed, else -> Open and rest Timeout timer

func NewBreaker(service string) *Circuitbreaker {
	return &Circuitbreaker{
		State:   Closed,
		service: service,
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
			log.Printf("[CIRCUIT] %s → HALF-OPEN (timeout expired)", cb.service) // 🔥
			return true
		}
		log.Printf("[CIRCUIT] %s → BLOCKED (OPEN)", cb.service) // 🔥
		return false
	default:
		return true
	}
}

// record result of request
func (cb *Circuitbreaker) RecordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if success {
		cb.FailureCount = 0
		cb.State = Closed
		log.Printf("[CIRCUIT] %s → CLOSED (recovered)", cb.service)
	} else {
		cb.FailureCount++
		cb.LastFailure = time.Now()
		if cb.FailureCount >= failureThreshold {
			cb.State = Open
		}
	}
}
