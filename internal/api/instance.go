package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"go.uber.org/zap"
)

type instanceStatusPayload struct {
	ID                 int64                    `json:"id,string"`
	OrgID              int64                    `json:"org_id,string"`
	NomadJobID         string                   `json:"nomad_job_id"`
	DesiredVersion     string                   `json:"desired_version"`
	CurrentVersion     string                   `json:"current_version"`
	Status             instance.InstanceStatus  `json:"status"`
	Role               instance.InstanceRole    `json:"role"`
	LifecycleState     instance.LifecycleState  `json:"lifecycle_state"`
	Readiness          instance.ReadinessStatus `json:"readiness"`
	ReadinessCheckedAt *time.Time               `json:"readiness_checked_at,omitempty"`
	ReadinessError     string                   `json:"readiness_error,omitempty"`
	Tier               instance.Tier            `json:"tier"`
	ComputeEngine      instance.ComputeEngine   `json:"compute_engine"`
	PlanID             string                   `json:"plan_id"`
	PriceID            string                   `json:"price_id"`
	SubscriptionID     string                   `json:"subscription_id"`
	LaunchURL          string                   `json:"launch_url"`
	LastError          string                   `json:"last_error,omitempty"`
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
}

func instanceStatusResponse(inst *instance.Instance) *instanceStatusPayload {
	if inst == nil {
		return nil
	}
	return &instanceStatusPayload{
		ID:                 inst.ID,
		OrgID:              inst.OrgID,
		NomadJobID:         inst.NomadJobID,
		DesiredVersion:     inst.DesiredVersion,
		CurrentVersion:     inst.CurrentVersion,
		Status:             inst.Status,
		Role:               inst.Role,
		LifecycleState:     inst.LifecycleState,
		Readiness:          inst.Readiness,
		ReadinessCheckedAt: inst.ReadinessCheckedAt,
		ReadinessError:     inst.ReadinessError,
		Tier:               inst.Tier,
		ComputeEngine:      inst.ComputeEngine,
		PlanID:             inst.PlanID,
		PriceID:            inst.PriceID,
		SubscriptionID:     inst.SubscriptionID,
		LaunchURL:          inst.LaunchURL,
		LastError:          inst.LastError,
		CreatedAt:          inst.CreatedAt,
		UpdatedAt:          inst.UpdatedAt,
	}
}

func (r *Router) GetInstanceStatus(c *gin.Context) {
	orgID, ok := r.resolveOrgID(c)
	if !ok {
		return
	}

	status, err := r.lifecycleUC.GetStatus(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if status == nil {
		// 404 means "Not Deployed" or "Org doesn't exist"
		c.JSON(http.StatusNotFound, gin.H{"status": "not_deployed", "org_id": orgID})
		return
	}
	c.JSON(http.StatusOK, instanceStatusResponse(status))
}

func (r *Router) StreamInstanceStatus(c *gin.Context) {
	orgID, ok := r.resolveOrgID(c)
	if !ok {
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	headers := c.Writer.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")
	headers.Set("X-Accel-Buffering", "no")

	c.Status(http.StatusOK)
	if _, err := fmt.Fprint(c.Writer, "retry: 3000\n\n"); err == nil {
		flusher.Flush()
	}

	ctx := c.Request.Context()
	pollTicker := time.NewTicker(4 * time.Second)
	heartbeatTicker := time.NewTicker(20 * time.Second)
	defer pollTicker.Stop()
	defer heartbeatTicker.Stop()

	var lastPayload string
	publish := func() bool {
		status, err := r.lifecycleUC.GetStatus(ctx, orgID)
		if err != nil {
			r.logger.Warn("stream instance status failed", zap.Error(err), zap.Int64("org_id", orgID))
			return true
		}

		var payload any
		if status == nil {
			payload = gin.H{"status": "missing", "org_id": orgID}
		} else {
			payload = instanceStatusResponse(status)
		}

		encoded, err := json.Marshal(payload)
		if err != nil {
			r.logger.Warn("stream instance status encode failed", zap.Error(err), zap.Int64("org_id", orgID))
			return true
		}

		next := string(encoded)
		if next == lastPayload {
			return true
		}
		lastPayload = next

		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", next); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	if !publish() {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			if !publish() {
				return
			}
		case <-heartbeatTicker.C:
			if _, err := fmt.Fprint(c.Writer, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (r *Router) DeployInstance(c *gin.Context) {
	var req struct {
		Version string `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	orgID, ok := r.resolveOrgID(c)
	if !ok {
		return
	}

	if err := r.deployUC.Execute(c.Request.Context(), orgID, req.Version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deployment_triggered"})
}

func (r *Router) StartInstance(c *gin.Context) {
	orgID, ok := r.resolveOrgID(c)
	if !ok {
		return
	}

	if err := r.lifecycleUC.Start(c.Request.Context(), orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "start_triggered"})
}

func (r *Router) StopInstance(c *gin.Context) {
	orgID, ok := r.resolveOrgID(c)
	if !ok {
		return
	}

	if err := r.lifecycleUC.Stop(c.Request.Context(), orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "stop_triggered"})
}

func (r *Router) PauseInstance(c *gin.Context) {
	orgID, ok := r.resolveOrgID(c)
	if !ok {
		return
	}

	if err := r.lifecycleUC.Pause(c.Request.Context(), orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "pause_triggered"})
}

func (r *Router) UpgradeInstance(c *gin.Context) {
	var req struct {
		Tier string `json:"tier"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	orgID, ok := r.resolveOrgID(c)
	if !ok {
		return
	}

	if err := r.upgradeUC.Upgrade(c.Request.Context(), orgID, instance.Tier(req.Tier)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "upgrade_initiated"})
}

func (r *Router) DowngradeInstance(c *gin.Context) {
	var req struct {
		Tier string `json:"tier"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	orgID, ok := r.resolveOrgID(c)
	if !ok {
		return
	}

	if err := r.upgradeUC.Downgrade(c.Request.Context(), orgID, instance.Tier(req.Tier)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "downgrade_scheduled"})
}
