package railzwayclient

import (
	"github.com/sony/gobreaker"
)

type CircuitBreaker interface {
	Execute(fn func() error) error
}

/*
|--------------------------------------------------------------------------
| Noop Breaker (disabled)
|--------------------------------------------------------------------------
*/

type noopBreaker struct{}

func (n *noopBreaker) Execute(fn func() error) error {
	return fn()
}

func NoopBreaker() CircuitBreaker {
	return &noopBreaker{}
}

/*
|--------------------------------------------------------------------------
| Gobreaker implementation
|--------------------------------------------------------------------------
*/

type gobreakerWrapper struct {
	cb *gobreaker.CircuitBreaker
}

func (g *gobreakerWrapper) Execute(fn func() error) error {
	_, err := g.cb.Execute(func() (any, error) {
		return nil, fn()
	})
	return err
}

func NewGobreaker(cfg Config) CircuitBreaker {
	settings := gobreaker.Settings{
		Name: "valora-client",

		MaxRequests: uint32(cfg.CBHalfOpenMaxSuccess),

		Interval: cfg.CBSamplingDuration,
		Timeout:  cfg.CBRecoveryTime,

		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < uint32(cfg.CBMinRequests) {
				return false
			}
			return counts.TotalFailures >= uint32(cfg.CBFailureThreshold)
		},

		IsSuccessful: func(err error) bool {
			return err == nil
		},
	}

	return &gobreakerWrapper{
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

func NewCircuitBreaker(cfg Config) CircuitBreaker {
	if !cfg.CircuitBreakerEnabled {
		return NoopBreaker()
	}
	return NewGobreaker(cfg)
}
