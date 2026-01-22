package reconciler

import (
	"context"
	"strings"
	"time"

	"github.com/smallbiznis/railzway-cloud/internal/domain/instance"
	"github.com/smallbiznis/railzway-cloud/internal/domain/provisioning"
	"go.uber.org/zap"
)

type InstanceReconciler struct {
	repo        instance.Repository
	provisioner provisioning.Provisioner
	logger      *zap.Logger
	interval    time.Duration
	batchSize   int
}

func NewInstanceReconciler(repo instance.Repository, provisioner provisioning.Provisioner, logger *zap.Logger) *InstanceReconciler {
	return &InstanceReconciler{
		repo:        repo,
		provisioner: provisioner,
		logger:      logger.Named("instance.reconciler"),
		interval:    10 * time.Second,
		batchSize:   50,
	}
}

func (r *InstanceReconciler) Run(ctx context.Context) {
	if err := r.reconcile(ctx); err != nil {
		r.logger.Error("reconcile_initial_failed", zap.Error(err))
	}

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.reconcile(ctx); err != nil {
				r.logger.Error("reconcile_failed", zap.Error(err))
			}
		}
	}
}

func (r *InstanceReconciler) reconcile(ctx context.Context) error {
	if r.provisioner == nil {
		return nil
	}
	statuses := []instance.InstanceStatus{instance.StatusProvisioning, instance.StatusUpgrading}
	items, err := r.repo.ListByStatus(ctx, statuses, r.batchSize)
	if err != nil {
		return err
	}

	for _, inst := range items {
		r.reconcileInstance(ctx, inst)
	}
	return nil
}

func (r *InstanceReconciler) reconcileInstance(ctx context.Context, inst *instance.Instance) {
	if inst == nil {
		return
	}
	raw, err := r.provisioner.GetStatus(ctx, inst.OrgID)
	if err != nil {
		r.logger.Warn("reconcile_status_failed",
			zap.Error(err),
			zap.Int64("org_id", inst.OrgID),
			zap.String("status", string(inst.Status)),
		)
		return
	}

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "running":
		inst.MarkActive(inst.DesiredVersion)
		if err := r.repo.Save(ctx, inst); err != nil {
			r.logger.Warn("reconcile_mark_active_failed", zap.Error(err), zap.Int64("org_id", inst.OrgID))
		}
	case "failed", "lost", "complete":
		inst.MarkProvisionFailed("provisioner_status:" + raw)
		if err := r.repo.Save(ctx, inst); err != nil {
			r.logger.Warn("reconcile_mark_failed_failed", zap.Error(err), zap.Int64("org_id", inst.OrgID))
		}
	default:
		return
	}
}
