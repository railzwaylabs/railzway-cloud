package railzwayclient

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	BaseURL string
	APIKey  string

	Timeout time.Duration

	RetryCount int
	RetryDelay time.Duration

	RateLimit int
	RateBurst int

	CacheTTL  time.Duration
	CacheSize int

	CircuitBreakerEnabled bool
	CBFailureThreshold    int
	CBRecoveryTime        time.Duration
	CBMinRequests         int
	CBSamplingDuration    time.Duration
	CBHalfOpenMaxSuccess  int
}

func LoadFromEnv() Config {
	return Config{
		BaseURL: os.Getenv("RAILZWAY_CLIENT_URL"),
		APIKey:  os.Getenv("RAILZWAY_API_KEY"),

		Timeout: time.Second * time.Duration(getInt("RAILZWAY_CLIENT_TIMEOUT", 30)),

		RetryCount: getInt("RAILZWAY_CLIENT_RETRY_COUNT", 3),
		RetryDelay: time.Second * time.Duration(getInt("RAILZWAY_CLIENT_RETRY_DELAY", 2)),

		RateLimit: getInt("RAILZWAY_CLIENT_RATE_LIMIT", 100),
		RateBurst: getInt("RAILZWAY_CLIENT_RATE_BURST", 2),

		CacheTTL:  time.Second * time.Duration(getInt("RAILZWAY_CLIENT_CACHE_TTL", 300)),
		CacheSize: getInt("RAILZWAY_CLIENT_CACHE_SIZE", 1000),

		CircuitBreakerEnabled: getBool("RAILZWAY_CLIENT_ENABLE_CIRCUIT_BREAKER", true),
		CBFailureThreshold:    getInt("RAILZWAY_CLIENT_CIRCUIT_BREAKER_FAILURE_THRESHOLD", 5),
		CBRecoveryTime:        time.Second * time.Duration(getInt("RAILZWAY_CLIENT_CIRCUIT_BREAKER_RECOVERY_TIME", 60)),
		CBMinRequests:         getInt("RAILZWAY_CLIENT_CIRCUIT_BREAKER_MIN_REQUESTS", 10),
		CBSamplingDuration:    time.Second * time.Duration(getInt("RAILZWAY_CLIENT_CIRCUIT_BREAKER_SAMPLING_DURATION", 60)),
		CBHalfOpenMaxSuccess:  getInt("RAILZWAY_CLIENT_CIRCUIT_BREAKER_HALF_OPEN_MAX_SUCCESS", 3),
	}
}

func getInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getBool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		return v == "true"
	}
	return def
}
