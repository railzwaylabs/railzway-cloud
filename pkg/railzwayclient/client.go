package railzwayclient

import (
	"net/http"
)

type Client struct {
	cfg     Config
	http    *http.Client
	retry   RetryPolicy
	cache   *Cache
	limiter *RateLimiter
	breaker CircuitBreaker
}

func NewFromEnv() *Client {
	return New(LoadFromEnv())
}

func New(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
		retry: RetryPolicy{
			MaxRetries: cfg.RetryCount,
			BaseDelay:  cfg.RetryDelay,
		},
		cache:   NewCache(cfg.CacheSize, cfg.CacheTTL),
		limiter: NewRateLimiter(cfg.RateLimit, cfg.RateBurst),
		breaker: NewCircuitBreaker(cfg),
	}
}
