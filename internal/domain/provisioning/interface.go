package provisioning

import (
	"context"

	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
)

// DeploymentConfig defines the parameters required for a deployment.
type DeploymentConfig struct {
	OrgID                  int64
	OrgSlug                string
	OrgName                string
	Version                string
	Tier                   instance.Tier
	ComputeEngine          instance.ComputeEngine
	DBConfig               DBConfig
	RateLimitRedisAddr     string
	RateLimitRedisPassword string
	RateLimitRedisDB       int

	// OAuth Configuration
	OAuth2URI                   string
	OAuth2ClientID              string
	OAuth2ClientSecret          string
	PaymentProviderConfigSecret string
}

// DBConfig holds the database connection details for the instance.
type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

// DatabaseProvisioner defines the interface for provisioning tenant databases.
type DatabaseProvisioner interface {
	// Provision creates the database and user for the given organization.
	// It must be idempotent.
	Provision(ctx context.Context, orgID int64, password string) error
}

// Provisioner defines the interface for the underlying infrastructure orchestrator (e.g., Nomad).
type Provisioner interface {
	// Deploy creates or updates the workload for the given configuration.
	Deploy(ctx context.Context, config *DeploymentConfig) error

	// Stop stops the workload.
	Stop(ctx context.Context, orgID int64) error

	// GetStatus retrieves the current status of the workload.
	GetStatus(ctx context.Context, orgID int64) (string, error)
}
