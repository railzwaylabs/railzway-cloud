package reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"go.uber.org/zap"
)

type LifecycleReconciler struct {
	repo      instance.Repository
	logger    *zap.Logger
	interval  time.Duration
	batchSize int
	client    *http.Client
}

func NewLifecycleReconciler(repo instance.Repository, logger *zap.Logger) *LifecycleReconciler {
	return &LifecycleReconciler{
		repo:      repo,
		logger:    logger.Named("lifecycle.reconciler"),
		interval:  15 * time.Second,
		batchSize: 50,
		client: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

func (r *LifecycleReconciler) Run(ctx context.Context) {
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

func (r *LifecycleReconciler) reconcile(ctx context.Context) error {
	items, err := r.repo.ListByStatus(ctx, instance.AllStatuses(), r.batchSize)
	if err != nil {
		return err
	}

	for _, inst := range items {
		r.reconcileInstance(ctx, inst)
	}
	return nil
}

func (r *LifecycleReconciler) reconcileInstance(ctx context.Context, inst *instance.Instance) {
	if inst == nil {
		return
	}

	readiness, readyErr := r.checkReadiness(ctx, inst.LaunchURL)
	now := time.Now().UTC()
	inst.Readiness = readiness
	inst.ReadinessCheckedAt = &now
	inst.ReadinessError = ""
	if readyErr != nil {
		inst.ReadinessError = readyErr.Error()
	}

	desired := r.computeLifecycle(inst)
	if desired != "" && desired != inst.LifecycleState {
		if !instance.CanTransitionLifecycle(inst.LifecycleState, desired) {
			r.logger.Warn("lifecycle_transition_blocked",
				zap.Int64("org_id", inst.OrgID),
				zap.String("current", string(inst.LifecycleState)),
				zap.String("target", string(desired)),
			)
		} else {
			inst.LifecycleState = desired
		}
	}

	if err := r.repo.Save(ctx, inst); err != nil {
		r.logger.Warn("lifecycle_reconcile_save_failed", zap.Error(err), zap.Int64("org_id", inst.OrgID))
	}
}

func (r *LifecycleReconciler) computeLifecycle(inst *instance.Instance) instance.LifecycleState {
	if inst == nil {
		return ""
	}

	switch inst.Status {
	case instance.StatusProvisioning, instance.StatusUpgrading:
		return instance.LifecycleMigrating
	case instance.StatusTerminated:
		return instance.LifecycleTerminated
	case instance.StatusStopped:
		return instance.LifecycleDraining
	}

	if inst.Readiness == instance.ReadinessReady {
		if inst.Role == instance.RolePrimary {
			return instance.LifecycleServing
		}
		return instance.LifecycleReady
	}

	if inst.Readiness == instance.ReadinessNotReady {
		if inst.Role == instance.RolePrimary {
			return instance.LifecycleDraining
		}
		return instance.LifecycleReady
	}

	return inst.LifecycleState
}

func (r *LifecycleReconciler) checkReadiness(ctx context.Context, launchURL string) (instance.ReadinessStatus, error) {
	base, err := baseURL(launchURL)
	if err != nil {
		return instance.ReadinessUnknown, err
	}

	healthOK, err := r.checkEndpoint(ctx, base, "/health")
	if err != nil || !healthOK {
		if err == nil {
			err = fmt.Errorf("health check failed")
		}
		return instance.ReadinessNotReady, err
	}

	readyOK, readyErr := r.checkReadyEndpoint(ctx, base)
	if readyErr != nil {
		if errorsIsNotFound(readyErr) {
			// Fallback: if /ready is missing, treat health as ready for now.
			return instance.ReadinessReady, nil
		}
		return instance.ReadinessNotReady, readyErr
	}
	if !readyOK {
		return instance.ReadinessNotReady, fmt.Errorf("ready=false")
	}
	return instance.ReadinessReady, nil
}

func (r *LifecycleReconciler) checkEndpoint(ctx context.Context, base string, path string) (bool, error) {
	target := strings.TrimRight(base, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return false, err
	}
	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("status=%d", resp.StatusCode)
	}
	return true, nil
}

func (r *LifecycleReconciler) checkReadyEndpoint(ctx context.Context, base string) (bool, error) {
	target := strings.TrimRight(base, "/") + "/ready"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return false, fmt.Errorf("not_found")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("status=%d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		// Treat 200 with unknown payload as ready.
		return true, nil
	}

	if val, ok := payload["ready"]; ok {
		if ready, ok := val.(bool); ok {
			return ready, nil
		}
	}

	if val, ok := payload["system_state"]; ok {
		if state, ok := val.(string); ok {
			return strings.EqualFold(state, "ready"), nil
		}
	}

	return true, nil
}

func baseURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("launch_url missing")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid launch_url")
	}
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host), nil
}

func errorsIsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not_found")
}
