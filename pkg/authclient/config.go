package authclient

import (
	"os"
	"strconv"
	"time"
)

// Config controls the railzway-auth admin client.
type Config struct {
	BaseURL    string
	TenantSlug string
	ClientID   string
	ClientSecret string
	ClientScope string
	Timeout    time.Duration
}

func LoadFromEnv() Config {
	return Config{
		BaseURL:    os.Getenv("AUTH_SERVICE_URL"),
		TenantSlug: os.Getenv("AUTH_SERVICE_TENANT"),
		ClientID:   os.Getenv("AUTH_SERVICE_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH_SERVICE_CLIENT_SECRET"),
		ClientScope: os.Getenv("AUTH_SERVICE_CLIENT_SCOPE"),
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
