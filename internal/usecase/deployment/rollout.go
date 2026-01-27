package deployment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"github.com/railzwaylabs/railzway-cloud/internal/version"
	"gorm.io/gorm"
)

type RolloutUseCase struct {
	db         *gorm.DB
	repo       instance.Repository
	versionReg *version.Registry
	cfg        *config.Config
}

type RolloutResult struct {
	TargetVersion string
	UpdatedCount  int64
	EnqueuedCount int64
}

const (
	deployEventType  = "deploy_instance"
	statusPending   = "pending"
	statusProcessing = "processing"
)

func NewRolloutUseCase(db *gorm.DB, repo instance.Repository, versionReg *version.Registry, cfg *config.Config) *RolloutUseCase {
	return &RolloutUseCase{
		db:         db,
		repo:       repo,
		versionReg: versionReg,
		cfg:        cfg,
	}
}

func (uc *RolloutUseCase) Rollout(ctx context.Context, version string) (*RolloutResult, error) {
	target, err := uc.resolveTargetVersion(ctx, version)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var updatedRows int64
	var enqueuedRows int64

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		update := tx.Model(&instance.Instance{}).
			Where("status <> ?", instance.StatusTerminated).
			Where("desired_version <> ?", target).
			Updates(map[string]any{
				"desired_version": target,
				"updated_at":      now,
			})
		if update.Error != nil {
			return update.Error
		}
		updatedRows = update.RowsAffected

		insert := tx.Exec(
			`INSERT INTO outbox_events (event_type, org_id, instance_id, status, attempts, created_at, updated_at)
			 SELECT ?, i.org_id, i.id, ?, 0, ?, ?
			 FROM instances i
			 WHERE i.status <> ?
			   AND i.desired_version = ?
			   AND (i.current_version IS NULL OR i.current_version <> i.desired_version OR i.status <> ?)
			   AND NOT EXISTS (
			     SELECT 1 FROM outbox_events e
			     WHERE e.instance_id = i.id
			       AND e.event_type = ?
			       AND e.status IN (?, ?)
			   )`,
			deployEventType,
			statusPending,
			now,
			now,
			instance.StatusTerminated,
			target,
			instance.StatusActive,
			deployEventType,
			statusPending,
			statusProcessing,
		)
		if insert.Error != nil {
			return insert.Error
		}
		enqueuedRows = insert.RowsAffected

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RolloutResult{
		TargetVersion: target,
		UpdatedCount:  updatedRows,
		EnqueuedCount: enqueuedRows,
	}, nil
}

func (uc *RolloutUseCase) resolveTargetVersion(ctx context.Context, raw string) (string, error) {
	target := strings.TrimSpace(raw)
	if target == "" {
		if uc.versionReg != nil {
			if v, err := uc.versionReg.GetDefaultVersion(ctx, "railzway"); err == nil && v != nil {
				target = v.Version
			}
		}
		if target == "" {
			target = strings.TrimSpace(uc.cfg.DefaultRailzwayOSSVersion)
		}
	}
	if target == "" {
		return "", fmt.Errorf("version is required")
	}
	if uc.versionReg != nil {
		ok, err := uc.versionReg.ValidateVersion(ctx, "railzway", target)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("version %s not found or not available", target)
		}
	}
	return target, nil
}
