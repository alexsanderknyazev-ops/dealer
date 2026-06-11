package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrOpen = errors.New("circuit breaker is open")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half_open"
	default:
		return "closed"
	}
}

// Config задаёт пороги circuit breaker.
type Config struct {
	FailureThreshold int
	SuccessThreshold int
	OpenTimeout      time.Duration
}

func (c Config) normalized() Config {
	if c.FailureThreshold <= 0 {
		c.FailureThreshold = 5
	}
	if c.SuccessThreshold <= 0 {
		c.SuccessThreshold = 2
	}
	if c.OpenTimeout <= 0 {
		c.OpenTimeout = 30 * time.Second
	}
	return c
}

// Breaker — классический circuit breaker (closed → open → half-open).
type Breaker struct {
	mu       sync.Mutex
	cfg      Config
	state    State
	failures int
	success  int
	openedAt time.Time
}

func New(cfg Config) *Breaker {
	return &Breaker{cfg: cfg.normalized(), state: StateClosed}
}

func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeHalfOpen(time.Now())
	return b.state
}

func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	b.maybeHalfOpenLocked(now)
	switch b.state {
	case StateOpen:
		return ErrOpen
	default:
		return nil
	}
}

func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case StateHalfOpen:
		b.success++
		if b.success >= b.cfg.SuccessThreshold {
			b.resetLocked(StateClosed)
		}
	case StateClosed:
		b.failures = 0
	}
}

func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	switch b.state {
	case StateHalfOpen:
		b.tripLocked(now)
	case StateClosed:
		b.failures++
		if b.failures >= b.cfg.FailureThreshold {
			b.tripLocked(now)
		}
	}
}

func (b *Breaker) maybeHalfOpen(now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeHalfOpenLocked(now)
}

func (b *Breaker) maybeHalfOpenLocked(now time.Time) {
	if b.state != StateOpen {
		return
	}
	if now.Sub(b.openedAt) >= b.cfg.OpenTimeout {
		b.state = StateHalfOpen
		b.success = 0
		b.failures = 0
	}
}

func (b *Breaker) tripLocked(now time.Time) {
	b.state = StateOpen
	b.openedAt = now
	b.failures = 0
	b.success = 0
}

func (b *Breaker) resetLocked(state State) {
	b.state = state
	b.failures = 0
	b.success = 0
}
