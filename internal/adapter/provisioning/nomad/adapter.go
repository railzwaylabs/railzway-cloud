package nomad

import (
	"context"

	"github.com/railzwaylabs/railzway-cloud/internal/domain/provisioning"
	"github.com/railzwaylabs/railzway-cloud/pkg/nomad"
)

type Adapter struct {
	client *nomad.Client
}

func NewAdapter(client *nomad.Client) *Adapter {
	return &Adapter{client: client}
}

func (a *Adapter) Deploy(ctx context.Context, cfg *provisioning.DeploymentConfig) error {
	jobCfg := nomad.JobConfig{
		OrgID:   cfg.OrgID,
		OrgSlug: cfg.OrgSlug,
		Version: cfg.Version,
		Tier:    nomad.Tier(cfg.Tier),

		ComputeEngine: nomad.ComputeEngine(cfg.ComputeEngine),
		DBConfig: nomad.DBConfig{
			Host:     cfg.DBConfig.Host,
			Port:     cfg.DBConfig.Port,
			Name:     cfg.DBConfig.Name,
			User:     cfg.DBConfig.User,
			Password: cfg.DBConfig.Password,
		},
		RateLimitRedisAddr:     cfg.RateLimitRedisAddr,
		RateLimitRedisPassword: cfg.RateLimitRedisPassword,
		RateLimitRedisDB:       cfg.RateLimitRedisDB,

		OAuth2URI:          cfg.OAuth2URI,
		OAuth2ClientID:     cfg.OAuth2ClientID,
		OAuth2ClientSecret: cfg.OAuth2ClientSecret,
		AuthJWTSecret:      cfg.AuthJWTSecret,
	}
	return a.client.DeployInstance(jobCfg)
}

func (a *Adapter) Stop(ctx context.Context, orgID int64) error {
	return a.client.StopInstance(orgID)
}

func (a *Adapter) GetStatus(ctx context.Context, orgID int64) (string, error) {
	return a.client.GetInstanceStatus(orgID)
}
