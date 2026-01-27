package deployment

import (
	"context"
	"fmt"

	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/billing"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/provisioning"
	"github.com/railzwaylabs/railzway-cloud/internal/organization"
)

type UpgradeUseCase struct {
	repo          instance.Repository
	provisioner   provisioning.Provisioner
	billingEngine billing.Engine
	priceResolver billing.PriceResolver
	orgService    *organization.Service
	cfg           *config.Config
}

func NewUpgradeUseCase(r instance.Repository, p provisioning.Provisioner, b billing.Engine, pr billing.PriceResolver, orgService *organization.Service, cfg *config.Config) *UpgradeUseCase {
	return &UpgradeUseCase{
		repo:          r,
		provisioner:   p,
		billingEngine: b,
		priceResolver: pr,
		orgService:    orgService,
		cfg:           cfg,
	}
}

func (uc *UpgradeUseCase) Upgrade(ctx context.Context, orgID int64, targetTier instance.Tier) error {
	inst, err := uc.repo.FindByOrgID(ctx, orgID)
	if err != nil {
		return err
	}
	if inst == nil {
		return fmt.Errorf("instance not found")
	}

	if inst.Tier == targetTier {
		return fmt.Errorf("already on tier %s", targetTier)
	}

	if !inst.CanUpgrade(targetTier) {
		return instance.ErrInvalidTierUpgrade
	}

	// 1. Resolve Price
	priceID, err := uc.priceResolver.ResolvePriceID(ctx, string(targetTier))
	if err != nil {
		return fmt.Errorf("billing config missing for tier %s", targetTier)
	}

	// 2. Deploy Infra
	orgSlug, err := uc.orgService.GetSlug(ctx, inst.OrgID)
	if err != nil {
		return fmt.Errorf("failed to resolve org slug: %w", err)
	}

	deployCfg := provisioning.DeploymentConfig{
		OrgID:              inst.OrgID,
		OrgSlug:            orgSlug,
		Version:            inst.DesiredVersion,
		Tier:               targetTier,
		ComputeEngine:      inst.ComputeEngine,
		OAuth2URI:          uc.cfg.OAuth2URI,
		OAuth2ClientID:     coalesce(inst.OAuthClientID, uc.cfg.TenantOAuth2ClientID),
		OAuth2ClientSecret: coalesce(inst.OAuthClientSecret, uc.cfg.TenantOAuth2ClientSecret),
		AuthJWTSecret:      generateJWTSecret(uc.cfg.TenantAuthJWTSecretKey, inst.OrgID),
	}
	if err := uc.provisioner.Deploy(ctx, &deployCfg); err != nil {
		return fmt.Errorf("failed to upgrade infra: %w", err)
	}

	// 3. Update Billing
	if inst.SubscriptionID != "" {
		params := billing.ChangePlanParams{
			SubscriptionID:    inst.SubscriptionID,
			NewPriceID:        priceID,
			ProrationBehavior: billing.CreateProration,
			EffectiveDate:     "immediate",
		}
		if err := uc.billingEngine.ChangePlan(ctx, params); err != nil {
			return fmt.Errorf("infra upgraded but billing failed: %w", err)
		}
	}

	// 4. Update State
	inst.MarkUpgrading(targetTier)
	return uc.repo.Save(ctx, inst)
}

func (uc *UpgradeUseCase) Downgrade(ctx context.Context, orgID int64, targetTier instance.Tier) error {
	inst, err := uc.repo.FindByOrgID(ctx, orgID)
	if err != nil {
		return err
	}
	if inst == nil {
		return fmt.Errorf("instance not found")
	}

	if inst.Tier == targetTier {
		return fmt.Errorf("already on tier %s", targetTier)
	}

	if !inst.CanDowngrade(targetTier) {
		return fmt.Errorf("cannot downgrade to %s", targetTier)
	}

	// 1. Resolve Price
	priceID, err := uc.priceResolver.ResolvePriceID(ctx, string(targetTier))
	if err != nil {
		return fmt.Errorf("billing config missing for tier %s", targetTier)
	}

	// 2. Schedule Billing Change
	if inst.SubscriptionID != "" {
		params := billing.ChangePlanParams{
			SubscriptionID:    inst.SubscriptionID,
			NewPriceID:        priceID,
			ProrationBehavior: billing.None,
			CancelAtPeriodEnd: true,
		}
		if err := uc.billingEngine.ChangePlan(ctx, params); err != nil {
			return fmt.Errorf("failed to schedule billing change: %w", err)
		}
	}

	// 3. Update State
	inst.ScheduleDowngrade()
	return uc.repo.Save(ctx, inst)
}
