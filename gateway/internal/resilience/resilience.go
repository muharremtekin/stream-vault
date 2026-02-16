package resilience

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	gobreaker "github.com/sony/gobreaker/v2"

	"github.com/streamvault/gateway/internal/config"
	"github.com/streamvault/gateway/internal/metrics"
)

// ErrBulkheadFull is returned when a service's concurrency limit is reached.
var ErrBulkheadFull = errors.New("bulkhead: max concurrent requests reached")

// Resilience orchestrates circuit breaker, bulkhead, and timeout patterns
// for each upstream service.
type Resilience struct {
	breakers  map[string]*gobreaker.CircuitBreaker[struct{}]
	bulkheads map[string]chan struct{}
	timeouts  map[string]time.Duration
	cfg       config.ResilienceConfig
}

// New creates a Resilience instance with per-service circuit breakers,
// bulkhead semaphores, and timeout settings.
func New(cfg config.ResilienceConfig, serviceNames []string) *Resilience {
	r := &Resilience{
		breakers:  make(map[string]*gobreaker.CircuitBreaker[struct{}], len(serviceNames)),
		bulkheads: make(map[string]chan struct{}, len(serviceNames)),
		timeouts:  make(map[string]time.Duration, len(serviceNames)),
		cfg:       cfg,
	}

	for _, name := range serviceNames {
		r.initService(name)
	}

	return r
}

func (r *Resilience) initService(name string) {
	// Circuit breaker
	cbSettings := gobreaker.Settings{
		Name:     fmt.Sprintf("cb-%s", name),
		MaxRequests: uint32(r.cfg.CircuitBreaker.SuccessThreshold),
		Interval:    0, // don't clear counts periodically
		Timeout:     r.cfg.CircuitBreaker.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(r.cfg.CircuitBreaker.FailureThreshold)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			serviceName := name[3:] // strip "cb-" prefix
			log.Warn().
				Str("service", serviceName).
				Str("from", from.String()).
				Str("to", to.String()).
				Msg("circuit breaker state change")

			// Update Prometheus gauge: 0=closed, 1=half-open, 2=open
			var stateVal float64
			switch to {
			case gobreaker.StateClosed:
				stateVal = 0
			case gobreaker.StateHalfOpen:
				stateVal = 1
			case gobreaker.StateOpen:
				stateVal = 2
				metrics.GatewayCircuitBreakerTripsTotal.WithLabelValues(serviceName).Inc()
			}
			metrics.GatewayCircuitBreakerState.WithLabelValues(serviceName).Set(stateVal)
		},
	}
	r.breakers[name] = gobreaker.NewCircuitBreaker[struct{}](cbSettings)

	// Bulkhead
	maxConcurrent := r.cfg.DefaultMaxConcurrent
	if svc, ok := r.cfg.Services[name]; ok && svc.MaxConcurrent > 0 {
		maxConcurrent = svc.MaxConcurrent
	}
	if maxConcurrent <= 0 {
		maxConcurrent = 30
	}
	r.bulkheads[name] = make(chan struct{}, maxConcurrent)

	// Timeout
	timeout := r.cfg.DefaultTimeout
	if svc, ok := r.cfg.Services[name]; ok && svc.Timeout > 0 {
		timeout = svc.Timeout
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	r.timeouts[name] = timeout

	log.Info().
		Str("service", name).
		Int("bulkhead_max", maxConcurrent).
		Dur("timeout", timeout).
		Uint32("cb_failure_threshold", r.cfg.CircuitBreaker.FailureThreshold).
		Msg("resilience initialized")
}

// GetTimeout returns the per-service proxy timeout.
func (r *Resilience) GetTimeout(service string) time.Duration {
	if t, ok := r.timeouts[service]; ok {
		return t
	}
	if r.cfg.DefaultTimeout > 0 {
		return r.cfg.DefaultTimeout
	}
	return 5 * time.Second
}

// AcquireBulkhead tries to acquire a slot in the service's concurrency
// semaphore. Returns false if the semaphore is full or the context is done.
func (r *Resilience) AcquireBulkhead(service string, ctx context.Context) bool {
	sem, ok := r.bulkheads[service]
	if !ok {
		return true // no bulkhead configured, allow
	}
	select {
	case sem <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	default:
		metrics.GatewayBulkheadRejectsTotal.WithLabelValues(service).Inc()
		return false
	}
}

// ReleaseBulkhead releases a slot in the service's concurrency semaphore.
func (r *Resilience) ReleaseBulkhead(service string) {
	if sem, ok := r.bulkheads[service]; ok {
		<-sem
	}
}

// Execute runs fn inside the service's circuit breaker. If the circuit is
// open, it returns gobreaker.ErrOpenState without calling fn.
func (r *Resilience) Execute(service string, fn func() error) error {
	cb, ok := r.breakers[service]
	if !ok {
		return fn()
	}
	_, err := cb.Execute(func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}

// IsCircuitError returns true if the error indicates the circuit breaker
// rejected the request (OPEN or HALF-OPEN at capacity).
func IsCircuitError(err error) bool {
	return errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests)
}

// StatusCapture wraps an http.ResponseWriter to capture the status code
// without buffering the response body.
type StatusCapture struct {
	http.ResponseWriter
	Code    int
	Written bool
}

// WriteHeader captures the status code and delegates to the underlying writer.
func (sc *StatusCapture) WriteHeader(code int) {
	sc.Code = code
	sc.Written = true
	sc.ResponseWriter.WriteHeader(code)
}

// Write delegates to the underlying writer, setting code 200 if WriteHeader
// was not called explicitly.
func (sc *StatusCapture) Write(b []byte) (int, error) {
	if !sc.Written {
		sc.Code = http.StatusOK
		sc.Written = true
	}
	return sc.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter for middleware that needs it.
func (sc *StatusCapture) Unwrap() http.ResponseWriter {
	return sc.ResponseWriter
}

// Flush implements http.Flusher for streaming proxies.
func (sc *StatusCapture) Flush() {
	if f, ok := sc.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
