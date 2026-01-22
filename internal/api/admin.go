package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Router) RolloutVersion(c *gin.Context) {
	var req struct {
		Version string `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result, err := r.rolloutUC.Rollout(c.Request.Context(), req.Version)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "rollout_enqueued",
		"target_version": result.TargetVersion,
		"updated_count":  result.UpdatedCount,
		"enqueued_count": result.EnqueuedCount,
	})
}
