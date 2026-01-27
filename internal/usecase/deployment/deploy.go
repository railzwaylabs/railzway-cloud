package deployment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/provisioning"
	"github.com/railzwaylabs/railzway-cloud/internal/organization"
	"github.com/railzwaylabs/railzway-cloud/pkg/authclient"
)

type DeployUseCase struct {
	repo          instance.Repository
	provisioner   provisioning.Provisioner
	dbProvisioner provisioning.DatabaseProvisioner
	dbConfig      provisioning.DBConfig // Default config connection params (host/port)
	runtimeCfg    RuntimeConfig
	orgService    *organization.Service
	cfg           *config.Config // OAuth and other config
	authClient    *authclient.Client
}

type RuntimeConfig struct {
	RateLimitRedisAddr     string
	RateLimitRedisPassword string
	RateLimitRedisDB       int
}

func NewDeployUseCase(
	repo instance.Repository,
	provisioner provisioning.Provisioner,
	dbProvisioner provisioning.DatabaseProvisioner,
	dbConfig provisioning.DBConfig,
	runtimeCfg RuntimeConfig,
	orgService *organization.Service,
	cfg *config.Config,
	authClient *authclient.Client,
) *DeployUseCase {
	return &DeployUseCase{
		repo:          repo,
		provisioner:   provisioner,
		dbProvisioner: dbProvisioner,
		dbConfig:      dbConfig,
		runtimeCfg:    runtimeCfg,
		orgService:    orgService,
		cfg:           cfg,
		authClient:    authClient,
	}
}

// Execute deploys or updates an instance for an organization.
func (uc *DeployUseCase) Execute(ctx context.Context, orgID int64, version string) error {
	// 1. Load Instance
	inst, err := uc.repo.FindByOrgID(ctx, orgID)
	if err != nil {
		return err
	}
	if inst == nil {
		return fmt.Errorf("instance not found for org %d", orgID)
	}

	// 2. Provision Database (Idempotent)
	if inst.DBUser == "" {
		// First time provisioning
		inst.DBUser = fmt.Sprintf("railzway_user_%d", inst.OrgID)
		inst.DBName = fmt.Sprintf("railzway_org_%d", inst.OrgID)
		inst.DBHost = uc.dbConfig.Host
		inst.DBPort = uc.dbConfig.Port

		// Generate random password
		password, err := generatePassword()
		if err != nil {
			return fmt.Errorf("failed to generate db password: %w", err)
		}
		inst.DBPassword = password
	}

	// Always ensure DB exists/user password is synced
	if err := uc.dbProvisioner.Provision(ctx, inst.OrgID, inst.DBPassword); err != nil {
		return fmt.Errorf("db provisioning failed: %w", err)
	}

	// 3. Prepare Config
	orgSlug, err := uc.orgService.GetSlug(ctx, inst.OrgID)
	if err != nil {
		return fmt.Errorf("failed to resolve org slug: %w", err)
	}

	launchURL := resolveLaunchURL(uc.cfg, orgSlug, inst.LaunchURL)
	if launchURL == "" {
		return fmt.Errorf("launch url is required")
	}

	client, err := uc.ensureOAuthClient(ctx, inst.OrgID, launchURL)
	if err != nil {
		return err
	}
	inst.LaunchURL = launchURL
	inst.OAuthClientID = client.ClientID
	if strings.TrimSpace(client.ClientSecret) != "" {
		inst.OAuthClientSecret = client.ClientSecret
	}
	if strings.TrimSpace(inst.OAuthClientSecret) == "" {
		return fmt.Errorf("auth client secret missing for org %d", inst.OrgID)
	}

	deployCfg := provisioning.DeploymentConfig{
		OrgID:         inst.OrgID,
		OrgSlug:       orgSlug,
		Version:       version,
		Tier:          inst.Tier,
		ComputeEngine: inst.ComputeEngine,
		DBConfig: provisioning.DBConfig{
			Host:     inst.DBHost,
			Port:     inst.DBPort,
			Name:     inst.DBName,
			User:     inst.DBUser,
			Password: inst.DBPassword,
		},
		RateLimitRedisAddr:     uc.runtimeCfg.RateLimitRedisAddr,
		RateLimitRedisPassword: uc.runtimeCfg.RateLimitRedisPassword,
		RateLimitRedisDB:       uc.runtimeCfg.RateLimitRedisDB,
		// OAuth Configuration
		OAuth2URI:          uc.cfg.OAuth2URI,
		OAuth2ClientID:     inst.OAuthClientID,
		OAuth2ClientSecret: inst.OAuthClientSecret,
		AuthJWTSecret:      generateJWTSecret(uc.cfg.TenantAuthJWTSecretKey, inst.OrgID),
	}

	// 3. Provision
	if err := uc.provisioner.Deploy(ctx, &deployCfg); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	// 4. Update State
	inst.DesiredVersion = version
	inst.Status = instance.StatusProvisioning
	inst.UpdatedAt = time.Now().UTC()
	// Note: Status will transition to Active once health checks pass.
	// This is handled by a separate monitoring process or webhook from Nomad.

	return uc.repo.Save(ctx, inst)
}

// generateJWTSecret generates a deterministic but unique JWT secret for each organization.
func generateJWTSecret(masterKey string, orgID int64) string {
	if masterKey == "" {
		masterKey = "default-insecure-key-change-in-production"
	}
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", masterKey, orgID)))
	return hex.EncodeToString(hash[:])
}

func resolveLaunchURL(cfg *config.Config, orgSlug string, current string) string {
	trimmed := strings.TrimSpace(current)
	if trimmed != "" {
		return trimmed
	}
	root := strings.TrimSpace(cfg.AppRootDomain)
	if root == "" || strings.TrimSpace(orgSlug) == "" {
		return ""
	}
	scheme := strings.TrimSpace(cfg.AppRootScheme)
	if scheme == "" {
		if strings.EqualFold(cfg.Environment, "production") {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := fmt.Sprintf("%s.%s", orgSlug, root)
	return fmt.Sprintf("%s://%s/login/railzway_com", scheme, host)
}

func (uc *DeployUseCase) ensureOAuthClient(ctx context.Context, orgID int64, launchURL string) (*authclient.OAuthClient, error) {
	if uc.authClient == nil {
		return nil, fmt.Errorf("auth client not configured")
	}
	creds, err := uc.authClient.EnsureOAuthClient(ctx, authclient.EnsureOAuthClientRequest{
		ExternalOrgID: orgID,
		RedirectURIs:  []string{launchURL},
	})
	if err != nil {
		return nil, fmt.Errorf("ensure oauth client: %w", err)
	}
	return creds, nil
}

func coalesce(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
