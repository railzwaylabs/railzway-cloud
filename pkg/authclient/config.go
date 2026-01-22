package authclient

import (
	"os"
	"strconv"
	"time"
)

// Config controls the railzway-auth admin client.
type Config struct {
	BaseURL    string
	AdminToken string
	TenantSlug string
	Timeout    time.Duration
}

func LoadFromEnv() Config {
	return Config{
		BaseURL:    os.Getenv("AUTH_SERVICE_URL"),
		AdminToken: os.Getenv("AUTH_SERVICE_ADMIN_TOKEN"),
		TenantSlug: os.Getenv("AUTH_SERVICE_TENANT"),
		Timeout:    time.Second * time.Duration(getInt("AUTH_SERVICE_TIMEOUT", 10)),
	}
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
