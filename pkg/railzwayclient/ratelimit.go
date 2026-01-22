package railzwayclient

import (
	"context"

	"golang.org/x/time/rate"
)

type RateLimiter struct{ l *rate.Limiter }

func NewRateLimiter(rpm, burst int) *RateLimiter {
	return &RateLimiter{
		l: rate.NewLimiter(rate.Limit(rpm)/60, burst),
	}
}
func (r *RateLimiter) Wait(ctx context.Context) { r.l.Wait(ctx) }
