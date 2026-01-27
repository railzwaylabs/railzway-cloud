package nomad

import "errors"

// Tier represents the pricing tier of the organization.
type Tier string

const (
	TierFreeTrial  Tier = "FREE_TRIAL"
	TierStarter    Tier = "STARTER"
	TierPro        Tier = "PRO"
	TierTeam       Tier = "TEAM"
	TierEnterprise Tier = "ENTERPRISE"
)

// ComputeEngine represents the underlying compute provider.
type ComputeEngine string

const (
	EngineHetzner      ComputeEngine = "hetzner"
	EngineDigitalOcean ComputeEngine = "digitalocean"
	EngineGCP          ComputeEngine = "gcp"
	EngineAWS          ComputeEngine = "aws"
)

// JobConfig holds the configuration required to generate a Nomad job.
type JobConfig struct {
	OrgID                  int64
	OrgSlug                string
	Tier                   Tier
	ComputeEngine          ComputeEngine
	Version                string
	DBConfig               DBConfig
	RateLimitRedisAddr     string
	RateLimitRedisPassword string
	RateLimitRedisDB       int

	// OAuth Configuration for federation sign-in
	OAuth2URI          string
	OAuth2ClientID     string
	OAuth2ClientSecret string
	OAuth2CallbackURL  string
	AuthJWTSecret      string

	// Bootstrap Configuration for deployed Railzway OSS instance
	BootstrapOrgID   int64  // OrgID from Cloud (same as OrgID)
	BootstrapOrgName string // Organization name
}

type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

// Validate checks if the JobConfig is valid.
func (c JobConfig) Validate() error {
	if c.OrgID <= 0 {
		return errors.New("invalid org_id")
	}
	if c.Version == "" {
		return errors.New("version is required")
	}
	switch c.Tier {
	case TierFreeTrial, TierStarter, TierPro, TierTeam, TierEnterprise:
		// Valid tiers
	default:
		return errors.New("invalid or unsupported tier")
	}
	switch c.ComputeEngine {
	case EngineHetzner, EngineDigitalOcean, EngineGCP, EngineAWS:
		// Valid engines
	default:
		return errors.New("invalid or unsupported compute engine")
	}
	return nil
}
