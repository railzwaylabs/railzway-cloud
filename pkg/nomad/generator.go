package nomad

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/nomad/api"
)

// GenerateJob creates a Nomad job specification based on the provided configuration.
// It enforces resource limits, priorities, and placement constraints defined by the pricing tier.
func GenerateJob(cfg JobConfig) (*api.Job, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid job config: %w", err)
	}

	jobName := fmt.Sprintf("railzway-org-%d", cfg.OrgID)
	jobType := "service"
	region := "global"

	// 1. Allocate Resources
	cpu, memoryMB, priority, tierLabel := allocateResources(cfg.Tier)

	// 2. Get Quotas
	quotaCfg := getTierQuotaConfig(cfg.Tier)

	// 3. Determine Update Strategy
	updateStanza := getUpdateStrategy(cfg.Tier)

	// 4. Build Environment Variables
	envVars := buildEnvVars(cfg, quotaCfg)

	// Task Group
	taskGroup := &api.TaskGroup{
		Name:   &jobName,
		Count:  intToPtr(1),
		Update: updateStanza,
		RestartPolicy: &api.RestartPolicy{
			Attempts: intToPtr(3),
			Interval: timeToPtr(5 * 60 * 1000 * 1000000), // 5m
			Delay:    timeToPtr(15 * 1000 * 1000000),     // 15s
			Mode:     stringToPtr("fail"),
		},
		Networks: []*api.NetworkResource{
			{
				DynamicPorts: []api.Port{
					{Label: "http", To: 8080},
				},
			},
		},
		Constraints: []*api.Constraint{
			{
				LTarget: "${node.meta.tier}",
				RTarget: tierLabel,
				Operand: "=",
			},
			{
				LTarget: "${node.meta.compute}",
				RTarget: string(cfg.ComputeEngine),
				Operand: "=",
			},
		},
	}

	// Skip constraints in development environment to allow running on local dev nomad agent
	if cfg.Version == "development" || strings.ToLower(os.Getenv("APP_ENV")) == "development" {
		taskGroup.Constraints = nil
	}

	// Task
	task := &api.Task{
		Name:   "railzway",
		Driver: "docker",
		Config: map[string]interface{}{
			"image": fmt.Sprintf("ghcr.io/smallbiznis/railzway:%s", cfg.Version),
			"ports": []string{"http"},
		},
		Env: envVars,
		Resources: &api.Resources{
			CPU:      &cpu,
			MemoryMB: &memoryMB,
		},
	}

	// Service (for Consul & Traefik)
	host := "railzway.com"
	if os.Getenv("APP_ROOT_DOMAIN") != "" {
		host = strings.ToLower(os.Getenv("APP_ROOT_DOMAIN"))
	}

	host = strings.TrimSpace(fmt.Sprintf("%s.%s", cfg.OrgSlug, host))
	serviceName := fmt.Sprintf("railzway-org-%d", cfg.OrgID)
	service := &api.Service{
		Name:      serviceName,
		PortLabel: "http",
		Tags: []string{
			"traefik.enable=true",
			fmt.Sprintf("traefik.http.routers.org-%d.rule=Host(`%s`)", cfg.OrgID, host),
			fmt.Sprintf("traefik.http.routers.org-%d.entrypoints=web", cfg.OrgID), // Assume 'web' is port 80
		},
		Checks: []api.ServiceCheck{
			{
				Type:     "http",
				Path:     "/health",
				Interval: 10 * 1000 * 1000000, // 10s
				Timeout:  2 * 1000 * 1000000,  // 2s
			},
		},
	}

	taskGroup.Tasks = []*api.Task{task}
	taskGroup.Services = []*api.Service{service}

	// Job
	job := &api.Job{
		ID:          &jobName,
		Name:        &jobName,
		Type:        &jobType,
		Region:      &region,
		Datacenters: []string{"dc1"}, // Could be dynamic if needed
		Priority:    &priority,
		TaskGroups:  []*api.TaskGroup{taskGroup},
	}

	return job, nil
}

// Helpers
func intToPtr(i int) *int                      { return &i }
func boolToPtr(b bool) *bool                   { return &b }
func stringToPtr(s string) *string             { return &s }
func timeToPtr(d time.Duration) *time.Duration { return &d }

type tierQuotaConfig struct {
	IngestOrgRate       string
	IngestOrgBurst      string
	IngestEndpointRate  string
	IngestEndpointBurst string

	QuotaOrgUser       string
	QuotaOrgDashboard  string
	QuotaOrgDataSource string
	QuotaOrgAPIKey     string
	QuotaUserOrg       string
	QuotaUsageMonthly  string
}

func getTierQuotaConfig(tier Tier) tierQuotaConfig {
	switch tier {
	case TierFreeTrial:
		return tierQuotaConfig{
			IngestOrgRate:       "5",
			IngestOrgBurst:      "10",
			IngestEndpointRate:  "5",
			IngestEndpointBurst: "10",
			QuotaOrgUser:        "2",
			QuotaOrgDashboard:   "5",
			QuotaOrgDataSource:  "2",
			QuotaOrgAPIKey:      "2",
			QuotaUserOrg:        "1",
			QuotaUsageMonthly:   "10000",
		}
	case TierStarter:
		return tierQuotaConfig{
			IngestOrgRate:       "10",
			IngestOrgBurst:      "20",
			IngestEndpointRate:  "10",
			IngestEndpointBurst: "20",
			QuotaOrgUser:        "5",
			QuotaOrgDashboard:   "20",
			QuotaOrgDataSource:  "5",
			QuotaOrgAPIKey:      "5",
			QuotaUserOrg:        "2",
			QuotaUsageMonthly:   "100000",
		}
	case TierPro:
		return tierQuotaConfig{
			IngestOrgRate:       "25",
			IngestOrgBurst:      "50",
			IngestEndpointRate:  "15",
			IngestEndpointBurst: "30",
			QuotaOrgUser:        "10",
			QuotaOrgDashboard:   "100",
			QuotaOrgDataSource:  "10",
			QuotaOrgAPIKey:      "10",
			QuotaUserOrg:        "10",
			QuotaUsageMonthly:   "1000000",
		}
	case TierTeam:
		return tierQuotaConfig{
			IngestOrgRate:       "100",
			IngestOrgBurst:      "200",
			IngestEndpointRate:  "50",
			IngestEndpointBurst: "100",
			QuotaOrgUser:        "50",
			QuotaOrgDashboard:   "500",
			QuotaOrgDataSource:  "50",
			QuotaOrgAPIKey:      "50",
			QuotaUserOrg:        "20",
			QuotaUsageMonthly:   "10000000",
		}
	case TierEnterprise:
		return tierQuotaConfig{
			IngestOrgRate:       "500",
			IngestOrgBurst:      "1000",
			IngestEndpointRate:  "200",
			IngestEndpointBurst: "500",
			QuotaOrgUser:        "100",
			QuotaOrgDashboard:   "1000",
			QuotaOrgDataSource:  "100",
			QuotaOrgAPIKey:      "100",
			QuotaUserOrg:        "50",
			QuotaUsageMonthly:   "100000000",
		}
	default:
		// Fallback to Pro settings if unknown (safe default)
		return tierQuotaConfig{
			IngestOrgRate:       "25",
			IngestOrgBurst:      "50",
			IngestEndpointRate:  "15",
			IngestEndpointBurst: "30",
			QuotaOrgUser:        "10",
			QuotaOrgDashboard:   "100",
			QuotaOrgDataSource:  "10",
			QuotaOrgAPIKey:      "10",
			QuotaUserOrg:        "10",
			QuotaUsageMonthly:   "1000000",
		}
	}
}

func allocateResources(tier Tier) (cpu int, memoryMB int, priority int, tierLabel string) {
	// Deterministic allocation of resources based on Tier
	switch tier {
	case TierFreeTrial:
		return 250, 512, 40, "free-trial"
	case TierStarter:
		return 500, 1024, 60, "starter"
	case TierPro:
		return 1000, 2048, 75, "pro"
	case TierTeam:
		return 2000, 4096, 90, "team"
	case TierEnterprise:
		return 4000, 8192, 100, "enterprise"
	default:
		// Safe default (Pro)
		return 1000, 2048, 75, "pro"
	}
}

func getUpdateStrategy(tier Tier) *api.UpdateStrategy {
	updateStanza := &api.UpdateStrategy{
		MaxParallel:      intToPtr(1),
		MinHealthyTime:   timeToPtr(10 * 1000 * 1000000),      // 10s
		HealthyDeadline:  timeToPtr(5 * 60 * 1000 * 1000000),  // 5m
		ProgressDeadline: timeToPtr(10 * 60 * 1000 * 1000000), // 10m
		AutoRevert:       boolToPtr(true),
		Canary:           intToPtr(0),
	}

	if tier == TierTeam || tier == TierEnterprise {
		// Stricter update strategy for Team/Enterprise
		updateStanza.Canary = intToPtr(1)
		updateStanza.AutoRevert = boolToPtr(true)
	} else if tier == TierPro {
		// Moderate update strategy
		updateStanza.AutoRevert = boolToPtr(true)
	} else {
		// Simple rolling update for Free Trial/Starter
		updateStanza.AutoRevert = boolToPtr(false)
	}

	return updateStanza
}

func buildEnvVars(cfg JobConfig, quotaCfg tierQuotaConfig) map[string]string {
	env := map[string]string{
		"APP_MODE":              "cloud",
		"DEFAULT_ORG":           fmt.Sprintf("%d", cfg.OrgID),
		"ENVIRONMENT":           "production",
		"CLOUD_METRICS_ENABLED": "false",

		// Bootstrap Configuration
		"BOOTSTRAP_DEFAULT_ORG_ID":   fmt.Sprintf("%d", cfg.BootstrapOrgID),
		"BOOTSTRAP_DEFAULT_ORG_NAME": cfg.BootstrapOrgName,

		// Database Injection
		"DB_HOST":     cfg.DBConfig.Host,
		"DB_PORT":     fmt.Sprintf("%d", cfg.DBConfig.Port),
		"DB_NAME":     cfg.DBConfig.Name,
		"DB_USER":     cfg.DBConfig.User,
		"DB_PASSWORD": cfg.DBConfig.Password,
		"DATABASE_URL": fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.DBConfig.User, cfg.DBConfig.Password, cfg.DBConfig.Host, cfg.DBConfig.Port, cfg.DBConfig.Name),

		// OAuth Federation
		"OAUTH2_CLIENT_ID":     cfg.OAuth2ClientID,
		"OAUTH2_CLIENT_SECRET": cfg.OAuth2ClientSecret,
		"AUTH_JWT_SECRET":      cfg.AuthJWTSecret,

		"AUTH_RAILZWAY_COM_NAME":          "Railzway.com",
		"AUTH_RAILZWAY_COM_ENABLED":       "true",
		"AUTH_RAILZWAY_COM_ALLOW_SIGN_UP": "false",

		// Map Cloud OAuth2 config to provider env vars
		"AUTH_RAILZWAY_COM_CLIENT_ID":     cfg.OAuth2ClientID,
		"AUTH_RAILZWAY_COM_CLIENT_SECRET": cfg.OAuth2ClientSecret,

		// Construct URLs using OAuth2URI (e.g., https://accounts.railzway.com)
		"AUTH_RAILZWAY_COM_AUTH_URL":  fmt.Sprintf("%s/authorize", cfg.OAuth2URI),
		"AUTH_RAILZWAY_COM_TOKEN_URL": fmt.Sprintf("%s/oauth/token", cfg.OAuth2URI),
		"AUTH_RAILZWAY_COM_API_URL":   fmt.Sprintf("%s/userinfo", cfg.OAuth2URI),
		"AUTH_RAILZWAY_COM_SCOPES":    "openid email profile",

		// Redis
		"RATE_LIMIT_REDIS_ADDR":     cfg.RateLimitRedisAddr,
		"RATE_LIMIT_REDIS_PASSWORD": cfg.RateLimitRedisPassword,
		"RATE_LIMIT_REDIS_DB":       fmt.Sprintf("%d", cfg.RateLimitRedisDB),

		// Usage Ingest
		"USAGE_INGEST_ORG_RATE":                quotaCfg.IngestOrgRate,
		"USAGE_INGEST_ORG_BURST":               quotaCfg.IngestOrgBurst,
		"USAGE_INGEST_ENDPOINT_RATE":           quotaCfg.IngestEndpointRate,
		"USAGE_INGEST_ENDPOINT_BURST":          quotaCfg.IngestEndpointBurst,
		"USAGE_INGEST_CONCURRENCY_TTL_SECONDS": "3",

		// Usage Quotas
		"QUOTA_ENABLED":           "true",
		"QUOTA_ORG_USER":          quotaCfg.QuotaOrgUser,
		"QUOTA_ORG_DASHBOARD":     quotaCfg.QuotaOrgDashboard,
		"QUOTA_ORG_DATA_SOURCE":   quotaCfg.QuotaOrgDataSource,
		"QUOTA_ORG_API_KEY":       quotaCfg.QuotaOrgAPIKey,
		"QUOTA_USER_ORG":          quotaCfg.QuotaUserOrg,
		"QUOTA_GLOBAL_USER":       "-1",
		"QUOTA_GLOBAL_ORG":        "1",
		"QUOTA_ORG_USAGE_MONTHLY": quotaCfg.QuotaUsageMonthly,
	}
	return env
}
