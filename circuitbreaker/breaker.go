package circuitbreaker

import (
	"sync"
	"time"
)

type State string

const (
	Closed   State = "Closed"
	Open     State = "Open"
	HalfOpen State = "Half-Open"
)

type Circuitbreaker struct {
	mu           sync.Mutex
	FailureCount int
	State        State
	LastFailure  time.Time
}

// Closed => working fine
// Open => failed for N consecutive times
// Half-Open => after Timeout, one request is send to Open circuit, if succeeds, -> Closed, else -> Open and rest Timeout timer

const failureThreshold = 3
const openTimeout = 10 * time.Second

func NewBreaker() *Circuitbreaker {
	return &Circuitbreaker{
		State: Closed,
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
			return true
		}
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
	} else {
		cb.FailureCount++
		cb.LastFailure = time.Now()
		if cb.FailureCount >= failureThreshold {
			cb.State = Open
		}
	}
}
