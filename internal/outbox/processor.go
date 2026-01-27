package outbox

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"github.com/railzwaylabs/railzway-cloud/internal/usecase/deployment"
	"github.com/railzwaylabs/railzway-cloud/pkg/railzwayclient"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Processor struct {
	db           *gorm.DB
	deployUC     *deployment.DeployUseCase
	ossClient    *railzwayclient.Client
	logger       *zap.Logger
	pollInterval time.Duration
	batchSize    int
	maxAttempts  int
}

func NewProcessor(db *gorm.DB, deployUC *deployment.DeployUseCase, ossClient *railzwayclient.Client, logger *zap.Logger) *Processor {
	return &Processor{
		db:           db,
		deployUC:     deployUC,
		ossClient:    ossClient,
		logger:       logger,
		pollInterval: 5 * time.Second,
		batchSize:    5,
		maxAttempts:  10,
	}
}

// Run polls the outbox so side effects happen after durable writes, keeping DB state authoritative.
func (p *Processor) Run(ctx context.Context) {
	if err := p.processBatch(ctx); err != nil {
		p.logger.Error("outbox_initial_poll_failed", zap.Error(err))
	}

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.processBatch(ctx); err != nil {
				p.logger.Error("outbox_poll_failed", zap.Error(err))
			}
		}
	}
}

func (p *Processor) processBatch(ctx context.Context) error {
	events, err := p.fetchAndLockPending(ctx)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := p.processEvent(ctx, event); err != nil {
			p.logger.Error("outbox_event_processing_failed",
				zap.Error(err),
				zap.Int64("event_id", event.ID),
				zap.String("event_type", string(event.EventType)),
			)
		}
	}

	return nil
}

func (p *Processor) fetchAndLockPending(ctx context.Context) ([]Event, error) {
	var events []Event
	now := time.Now().UTC()

	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(
			`SELECT * FROM outbox_events
			 WHERE status IN (?, ?)
			   AND (next_attempt_at IS NULL OR next_attempt_at <= ?)
			   AND attempts < ?
			 ORDER BY created_at ASC
			 LIMIT ?
			 FOR UPDATE SKIP LOCKED`,
			StatusPending,
			StatusFailed,
			now,
			p.maxAttempts,
			p.batchSize,
		).Scan(&events).Error; err != nil {
			return err
		}

		if len(events) == 0 {
			return nil
		}

		ids := make([]int64, 0, len(events))
		for i := range events {
			ids = append(ids, events[i].ID)
			events[i].Attempts++
		}

		return tx.Model(&Event{}).
			Where("id IN ?", ids).
			Updates(map[string]any{
				"status":     StatusProcessing,
				"attempts":   gorm.Expr("attempts + 1"),
				"locked_at":  now,
				"updated_at": now,
				"last_error": nil,
			}).Error
	})

	return events, err
}

func (p *Processor) processEvent(ctx context.Context, event Event) error {
	switch event.EventType {
	case EventTypeDeployInstance:
		return p.handleDeployInstance(ctx, event)
	default:
		return p.markEventFailed(ctx, event, fmt.Errorf("unsupported event type: %s", event.EventType))
	}
}

type organizationRecord struct {
	ID            int64  `gorm:"primaryKey"`
	Name          string `gorm:"type:varchar(255)"`
	OSSCustomerID string `gorm:"column:oss_customer_id"`
}

func (organizationRecord) TableName() string {
	return "organizations"
}

func (p *Processor) handleDeployInstance(ctx context.Context, event Event) error {
	inst, err := p.loadInstance(ctx, event.InstanceID)
	if err != nil {
		return p.markEventFailed(ctx, event, fmt.Errorf("load instance: %w", err))
	}
	if inst == nil {
		return p.markEventFailed(ctx, event, fmt.Errorf("instance not found"))
	}
	if inst.OrgID != event.OrgID {
		return p.markEventFailed(ctx, event, fmt.Errorf("instance org mismatch"))
	}

	if inst.Status == instance.StatusActive {
		return p.markEventCompleted(ctx, event.ID)
	}

	if err := p.markInstanceProvisioning(ctx, inst.ID); err != nil {
		return p.markEventFailed(ctx, event, err)
	}

	org, err := p.loadOrganization(ctx, event.OrgID)
	if err != nil {
		return p.markEventFailed(ctx, event, fmt.Errorf("load organization: %w", err))
	}
	if org == nil {
		return p.markEventFailed(ctx, event, fmt.Errorf("organization not found"))
	}

	if err := p.ensureCustomer(ctx, org); err != nil {
		return p.markEventFailed(ctx, event, err)
	}

	if err := p.ensureSubscription(ctx, inst, org); err != nil {
		return p.markEventFailed(ctx, event, err)
	}

	// Activate subscription BEFORE deployment to ensure billing is active
	// before the instance starts running
	if inst.SubscriptionID != "" {
		if err := p.ossClient.ActivateSubscription(ctx, inst.SubscriptionID); err != nil {
			return p.markEventFailed(ctx, event, fmt.Errorf("activate subscription: %w", err))
		}
		p.logger.Info("subscription_activated",
			zap.String("subscription_id", inst.SubscriptionID),
			zap.Int64("org_id", inst.OrgID),
		)
	}

	// Deploy instance - if this fails, we need to rollback the subscription
	if err := p.deployUC.Execute(ctx, inst.OrgID, inst.DesiredVersion); err != nil {
		// Rollback: cancel subscription immediately since deployment failed
		p.rollbackSubscription(ctx, inst.SubscriptionID)
		return p.markEventFailed(ctx, event, fmt.Errorf("deployment failed: %w", err))
	}

	return p.markEventCompleted(ctx, event.ID)
}

func (p *Processor) loadInstance(ctx context.Context, instanceID int64) (*instance.Instance, error) {
	var inst instance.Instance
	if err := p.db.WithContext(ctx).First(&inst, "id = ?", instanceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &inst, nil
}

func (p *Processor) loadOrganization(ctx context.Context, orgID int64) (*organizationRecord, error) {
	var org organizationRecord
	if err := p.db.WithContext(ctx).First(&org, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

func (p *Processor) ensureCustomer(ctx context.Context, org *organizationRecord) error {
	if org.OSSCustomerID != "" {
		return nil
	}

	externalID := fmt.Sprintf("org_%d", org.ID)
	orgEmail := fmt.Sprintf("org_%d@railzway.com", org.ID)

	customer, err := p.ossClient.EnsureCustomer(ctx, org.Name, orgEmail, externalID)
	if err != nil {
		return fmt.Errorf("ensure oss customer: %w", err)
	}

	now := time.Now().UTC()
	if err := p.db.WithContext(ctx).Model(&organizationRecord{}).
		Where("id = ?", org.ID).
		Updates(map[string]any{
			"oss_customer_id": customer.ID,
			"updated_at":      now,
		}).Error; err != nil {
		return fmt.Errorf("update organization oss customer id: %w", err)
	}

	org.OSSCustomerID = customer.ID
	return nil
}

func (p *Processor) ensureSubscription(ctx context.Context, inst *instance.Instance, org *organizationRecord) error {
	if inst.SubscriptionID != "" {
		subscription, err := p.ossClient.GetSubscription(ctx, inst.SubscriptionID)
		if err != nil {
			return fmt.Errorf("load subscription: %w", err)
		}
		if subscription != nil && shouldReplaceSubscription(subscription.Status) {
			p.logger.Warn("subscription_inactive_replacing",
				zap.String("subscription_id", inst.SubscriptionID),
				zap.String("status", subscription.Status),
				zap.Int64("org_id", inst.OrgID),
			)
			inst.SubscriptionID = ""
		} else {
			return nil
		}
	}

	priceID := strings.TrimSpace(inst.PriceID)
	if priceID == "" {
		return fmt.Errorf("missing price_id in instance - price must be selected during provisioning")
	}

	items := []railzwayclient.CreateSubscriptionItemRequest{
		{
			PriceID:  priceID,
			Quantity: 1,
		},
	}

	subscription, err := p.ossClient.CreateSubscription(ctx, org.OSSCustomerID, "monthly", items)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}

	now := time.Now().UTC()
	updates := map[string]any{
		"subscription_id": subscription.ID,
		"updated_at":      now,
	}
	if inst.PriceID == "" && priceID != "" {
		updates["price_id"] = priceID
	}

	if err := p.db.WithContext(ctx).Model(&instance.Instance{}).
		Where("id = ?", inst.ID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("update subscription id: %w", err)
	}

	inst.SubscriptionID = subscription.ID
	if inst.PriceID == "" && priceID != "" {
		inst.PriceID = priceID
	}
	return nil
}

func shouldReplaceSubscription(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "CANCELED", "CANCELLED", "ENDED":
		return true
	default:
		return false
	}
}

func (p *Processor) markInstanceProvisioning(ctx context.Context, instanceID int64) error {
	allowed := []instance.InstanceStatus{instance.StatusInit, instance.StatusProvisionFailed}
	return p.markInstanceStatus(ctx, instanceID, allowed, instance.StatusProvisioning, "")
}

func (p *Processor) markInstanceProvisionFailed(ctx context.Context, instanceID int64, errMsg string) error {
	allowed := []instance.InstanceStatus{instance.StatusInit, instance.StatusProvisioning}
	return p.markInstanceStatus(ctx, instanceID, allowed, instance.StatusProvisionFailed, errMsg)
}

func (p *Processor) markInstanceStatus(ctx context.Context, instanceID int64, allowed []instance.InstanceStatus, next instance.InstanceStatus, errMsg string) error {
	now := time.Now().UTC()
	updates := map[string]any{
		"status":     next,
		"updated_at": now,
	}
	if errMsg == "" {
		updates["last_error"] = nil
	} else {
		updates["last_error"] = errMsg
	}

	result := p.db.WithContext(ctx).Model(&instance.Instance{}).
		Where("id = ? AND status IN ?", instanceID, allowed).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	currentStatus, err := p.getInstanceStatus(ctx, instanceID)
	if err != nil {
		return err
	}
	if currentStatus == next {
		return nil
	}

	return fmt.Errorf("invalid state transition from %s to %s", currentStatus, next)
}

func (p *Processor) getInstanceStatus(ctx context.Context, instanceID int64) (instance.InstanceStatus, error) {
	var status string
	if err := p.db.WithContext(ctx).Model(&instance.Instance{}).
		Select("status").
		Where("id = ?", instanceID).
		Scan(&status).Error; err != nil {
		return "", err
	}
	return instance.InstanceStatus(status), nil
}

func (p *Processor) markEventCompleted(ctx context.Context, eventID int64) error {
	now := time.Now().UTC()
	return p.db.WithContext(ctx).Model(&Event{}).
		Where("id = ? AND status = ?", eventID, StatusProcessing).
		Updates(map[string]any{
			"status":       StatusCompleted,
			"processed_at": now,
			"updated_at":   now,
			"last_error":   nil,
		}).Error
}

func (p *Processor) markEventFailed(ctx context.Context, event Event, err error) error {
	if err == nil {
		return nil
	}

	if event.InstanceID != 0 {
		_ = p.markInstanceProvisionFailed(ctx, event.InstanceID, err.Error())
	}

	now := time.Now().UTC()
	nextAttempt := now.Add(backoffDuration(event.Attempts))

	updateErr := p.db.WithContext(ctx).Model(&Event{}).
		Where("id = ?", event.ID).
		Updates(map[string]any{
			"status":          StatusFailed,
			"last_error":      err.Error(),
			"next_attempt_at": nextAttempt,
			"updated_at":      now,
		}).Error
	if updateErr != nil {
		return fmt.Errorf("mark event failed: %w (original error: %v)", updateErr, err)
	}
	return err
}

func backoffDuration(attempt int) time.Duration {
	if attempt <= 1 {
		return 10 * time.Second
	}

	maxBackoff := 5 * time.Minute
	base := 10 * time.Second
	shift := attempt - 1
	if shift > 6 {
		shift = 6
	}

	d := base * time.Duration(1<<shift)
	if d > maxBackoff {
		return maxBackoff
	}
	return d
}

// rollbackSubscription cancels a subscription immediately when deployment fails.
// This prevents users from being charged for instances that failed to deploy.
func (p *Processor) rollbackSubscription(ctx context.Context, subscriptionID string) {
	if subscriptionID == "" {
		return
	}

	p.logger.Info("rolling_back_subscription",
		zap.String("subscription_id", subscriptionID),
	)

	// Cancel immediately (not at period end) since deployment failed
	if err := p.ossClient.CancelSubscription(ctx, subscriptionID, false); err != nil {
		p.logger.Error("failed_to_rollback_subscription",
			zap.String("subscription_id", subscriptionID),
			zap.Error(err),
		)
	} else {
		p.logger.Info("subscription_rolled_back",
			zap.String("subscription_id", subscriptionID),
		)
	}
}
