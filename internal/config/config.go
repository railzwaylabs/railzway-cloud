package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds application configuration.
type Config struct {
	AppName    string
	AppVersion string
	Port       string

	Environment      string
	AuthCookieSecure bool
	AuthJWTSecret    string
	AdminAPIToken    string
	AppRootDomain    string
	AppRootScheme    string

	OTLPEndpoint string

	// Railzway OSS instance configuration
	DefaultRailzwayOSSVersion string // Default version for new Railzway OSS instances

	DBType            string
	DBHost            string
	DBPort            string
	DBName            string
	DBUser            string
	DBPassword        string
	DBSSLMode         string
	DBMaxIdleConn     int
	DBMaxOpenConn     int
	DBConnMaxLifetime int
	DBConnMaxIdleTime int

	ProvisionDBHost                 string
	ProvisionDBPort                 string
	ProvisionDBName                 string
	ProvisionDBUser                 string
	ProvisionDBPassword             string
	ProvisionDBSSLMode              string
	ProvisionRateLimitRedisAddr     string
	ProvisionRateLimitRedisPassword string
	ProvisionRateLimitRedisDB       int

	OAuth2ClientID     string // Cloud backend OAuth (for Cloud UI)
	OAuth2ClientSecret string // Cloud backend OAuth (for Cloud UI)
	OAuth2URI          string // OAuth provider base URL (e.g., https://railzway.us.auth0.com)
	OAuth2CallbackURL  string // OAuth callback URL (e.g., http://localhost:8080/auth/callback)

	// Tenant OAuth Configuration (for deployed OSS instances)
	TenantOAuth2ClientID     string // Shared OAuth app for all tenant instances
	TenantOAuth2ClientSecret string // Shared OAuth secret
	TenantAuthJWTSecretKey   string // Master key for generating per-org JWT secrets

	StaticDir string
}

// Load loads configuration from environment variables and .env file.
func Load() *Config {
	_ = godotenv.Load()
	environment := getenv("ENVIRONMENT", "development")
	authCookieSecure := environment == "production"
	if !authCookieSecure {
		authCookieSecure = getenvBool("AUTH_COOKIE_SECURE", false)
	}

	dbHost := getenv("DB_HOST", "localhost")
	dbPort := getenv("DB_PORT", "5433")
	dbName := getenv("DB_NAME", "cloud")
	dbUser := getenv("DB_USER", "postgres")
	dbPassword := getenv("DB_PASSWORD", "35411231")
	dbSSLMode := getenv("DB_SSL_MODE", "disable")
	provisionDBHost := getenv("PROVISION_DB_HOST", dbHost)
	provisionDBPort := getenv("PROVISION_DB_PORT", dbPort)
	provisionDBName := getenv("PROVISION_DB_NAME", dbName)
	provisionDBUser := getenv("PROVISION_DB_USER", dbUser)
	provisionDBPassword := getenv("PROVISION_DB_PASSWORD", dbPassword)
	provisionDBSSLMode := getenv("PROVISION_DB_SSL_MODE", dbSSLMode)
	provisionRateLimitRedisAddr := strings.TrimSpace(getenv("PROVISION_RATE_LIMIT_REDIS_ADDR", ""))
	provisionRateLimitRedisPassword := strings.TrimSpace(getenv("PROVISION_RATE_LIMIT_REDIS_PASSWORD", ""))
	provisionRateLimitRedisDB := getenvInt("PROVISION_RATE_LIMIT_REDIS_DB", 0)

	callbackURL := strings.TrimSpace(getenv("OAUTH2_CALLBACK_URL", ""))
	if callbackURL == "" {
		callbackURL = strings.TrimSpace(getenv("OAUTH2_CALLBACK_URI", ""))
	}

	cfg := Config{
		AppName:                         getenv("APP_SERVICE", "railzway-cloud"),
		AppVersion:                      getenv("APP_VERSION", "0.1.0"),
		Port:                            getenv("PORT", "8081"),
		Environment:                     environment,
		AuthCookieSecure:                authCookieSecure,
		AuthJWTSecret:                   strings.TrimSpace(getenv("AUTH_JWT_SECRET", "")),
		AdminAPIToken:                   strings.TrimSpace(getenv("ADMIN_API_TOKEN", "")),
		AppRootDomain:                   strings.TrimLeft(strings.TrimSpace(getenv("APP_ROOT_DOMAIN", "")), "."),
		AppRootScheme:                   strings.TrimSpace(getenv("APP_ROOT_SCHEME", "")),
		OTLPEndpoint:                    getenv("OTLP_ENDPOINT", "localhost:4317"),
		DefaultRailzwayOSSVersion:       getenv("RAILZWAY_OSS_VERSION", "v1.6.0"),
		DBType:                          getenv("DB_TYPE", "postgres"),
		DBHost:                          dbHost,
		DBPort:                          dbPort,
		DBName:                          dbName,
		DBUser:                          dbUser,
		DBPassword:                      dbPassword,
		DBSSLMode:                       dbSSLMode,
		DBMaxIdleConn:                   10,
		DBMaxOpenConn:                   100,
		DBConnMaxLifetime:               3600,
		DBConnMaxIdleTime:               60,
		ProvisionDBHost:                 provisionDBHost,
		ProvisionDBPort:                 provisionDBPort,
		ProvisionDBName:                 provisionDBName,
		ProvisionDBUser:                 provisionDBUser,
		ProvisionDBPassword:             provisionDBPassword,
		ProvisionDBSSLMode:              provisionDBSSLMode,
		ProvisionRateLimitRedisAddr:     provisionRateLimitRedisAddr,
		ProvisionRateLimitRedisPassword: provisionRateLimitRedisPassword,
		ProvisionRateLimitRedisDB:       provisionRateLimitRedisDB,
		OAuth2ClientID:                  strings.TrimSpace(getenv("OAUTH2_CLIENT_ID", "")),
		OAuth2ClientSecret:              strings.TrimSpace(getenv("OAUTH2_CLIENT_SECRET", "")),
		OAuth2URI:                       strings.TrimSpace(getenv("OAUTH2_URI", "")),
		OAuth2CallbackURL:               callbackURL,
		// Tenant OAuth for deployed instances
		TenantOAuth2ClientID:     strings.TrimSpace(getenv("TENANT_OAUTH2_CLIENT_ID", "")),
		TenantOAuth2ClientSecret: strings.TrimSpace(getenv("TENANT_OAUTH2_CLIENT_SECRET", "")),
		TenantAuthJWTSecretKey:   strings.TrimSpace(getenv("TENANT_AUTH_JWT_SECRET_KEY", "")),
		StaticDir:                getenv("STATIC_DIR", "apps/railzway/dist"), // Assumes running from repo root
	}

	return &cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvBool(key string, def bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return def
	}
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

func getenvInt64(key string, def int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return def
	}
	return parsed
}

func getenvInt(key string, def int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return parsed
}

func parseServices(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		log.Println("no services enabled for migration")
	}
	return out
}
