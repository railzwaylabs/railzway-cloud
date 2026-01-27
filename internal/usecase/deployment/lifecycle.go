package deployment

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/billing"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/provisioning"
	"github.com/railzwaylabs/railzway-cloud/internal/organization"
)

type LifecycleUseCase struct {
	repo          instance.Repository
	provisioner   provisioning.Provisioner
	billingEngine billing.Engine
	orgService    *organization.Service
	cfg           *config.Config
}

func NewLifecycleUseCase(r instance.Repository, p provisioning.Provisioner, b billing.Engine, orgService *organization.Service, cfg *config.Config) *LifecycleUseCase {
	return &LifecycleUseCase{
		repo:          r,
		provisioner:   p,
		billingEngine: b,
		orgService:    orgService,
		cfg:           cfg,
	}
}

func (uc *LifecycleUseCase) Stop(ctx context.Context, orgID int64) error {
	inst, err := uc.repo.FindByOrgID(ctx, orgID)
	if err != nil {
		return err
	}
	if inst == nil {
		return fmt.Errorf("instance not found")
	}

	// 1. Stop Infrastructure
	if err := uc.provisioner.Stop(ctx, orgID); err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	// 2. Pause Billing
	if inst.SubscriptionID != "" {
		if err := uc.billingEngine.PauseSubscription(ctx, inst.SubscriptionID); err != nil {
			fmt.Printf("warning: failed to pause subscription %s: %v\n", inst.SubscriptionID, err)
		}
	}

	// 3. Update State
	inst.MarkStopped()
	return uc.repo.Save(ctx, inst)
}

func (uc *LifecycleUseCase) Pause(ctx context.Context, orgID int64) error {
	return uc.Stop(ctx, orgID)
}

func (uc *LifecycleUseCase) Start(ctx context.Context, orgID int64) error {
	inst, err := uc.repo.FindByOrgID(ctx, orgID)
	if err != nil {
		return err
	}
	if inst == nil {
		return fmt.Errorf("instance not found")
	}

	if inst.Status != instance.StatusStopped {
		return fmt.Errorf("instance is not stopped")
	}

	// 1. Start Infrastructure
	orgSlug, err := uc.orgService.GetSlug(ctx, inst.OrgID)
	if err != nil {
		return fmt.Errorf("failed to resolve org slug: %w", err)
	}

	deployCfg := provisioning.DeploymentConfig{
		OrgID:              inst.OrgID,
		OrgSlug:            orgSlug,
		Version:            inst.DesiredVersion,
		Tier:               inst.Tier,
		ComputeEngine:      inst.ComputeEngine,
		OAuth2URI:          uc.cfg.OAuth2URI,
		OAuth2ClientID:     coalesce(inst.OAuthClientID, uc.cfg.TenantOAuth2ClientID),
		OAuth2ClientSecret: coalesce(inst.OAuthClientSecret, uc.cfg.TenantOAuth2ClientSecret),
		AuthJWTSecret:      generateJWTSecret(uc.cfg.TenantAuthJWTSecretKey, inst.OrgID),
		// Bootstrap Configuration
		BootstrapOrgID:   inst.OrgID,
		BootstrapOrgName: orgSlug,
	}
	if err := uc.provisioner.Deploy(ctx, &deployCfg); err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	// 2. Resume Billing
	if inst.SubscriptionID != "" {
		if err := uc.billingEngine.ResumeSubscription(ctx, inst.SubscriptionID); err != nil {
			fmt.Printf("warning: failed to resume subscription %s: %v\n", inst.SubscriptionID, err)
		}
	}

	// 3. Update State
	inst.MarkRunning(inst.CurrentVersion)
	// Manual time update if repo doesn't
	inst.UpdatedAt = time.Now().UTC()

	return uc.repo.Save(ctx, inst)
}

func (uc *LifecycleUseCase) GetStatus(ctx context.Context, orgID int64) (*instance.Instance, error) {
	inst, err := uc.repo.FindByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, nil
	}

	if uc.provisioner != nil && inst.Status == instance.StatusProvisioning {
		if next, ok := uc.resolveProvisioningStatus(ctx, inst); ok {
			if inst.Status != next {
				inst.Status = next
				if next == instance.StatusActive && inst.CurrentVersion == "" {
					inst.CurrentVersion = inst.DesiredVersion
				}
				inst.UpdatedAt = time.Now().UTC()
				if err := uc.repo.Save(ctx, inst); err != nil {
					fmt.Printf("warning: failed to save instance state: %v\n", err)
				}
			}
		}
	}

	return inst, nil
}

func (uc *LifecycleUseCase) resolveProvisioningStatus(ctx context.Context, inst *instance.Instance) (instance.InstanceStatus, bool) {
	raw, err := uc.provisioner.GetStatus(ctx, inst.OrgID)
	if err != nil {
		return "", false
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "running":
		if uc.instanceReachable(ctx, inst.LaunchURL) {
			return instance.StatusActive, true
		}
		return "", false
	case "failed", "lost", "complete":
		return instance.StatusProvisionFailed, true
	default:
		return "", false
	}
}

func (uc *LifecycleUseCase) instanceReachable(ctx context.Context, launchURL string) bool {
	target := strings.TrimSpace(launchURL)
	if target == "" {
		return false
	}

	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return false
	}
	_ = resp.Body.Close()
	return true
}
