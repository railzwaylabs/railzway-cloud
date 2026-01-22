package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smallbiznis/railzway-cloud/internal/onboarding"
	"go.uber.org/zap"
)

func (r *Router) CheckOrgName(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name parameter required"})
		return
	}

	available, err := r.onboardingSvc.CheckOrgName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"available": available,
		"name":      name,
	})
}

func (r *Router) InitializeOrganization(c *gin.Context) {
	var req struct {
		PlanID  string `json:"plan_id"`  // Deprecated: use price_id
		PriceID string `json:"price_id"` // Actual price ID from pricing API
		OrgName string `json:"org_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	// Get User ID from Session/Context
	// Middleware sets "UserID" (int64)
	val, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(int64)

	// Frontend sends actual price_id in plan_id field for now
	priceID := req.PriceID
	if priceID == "" {
		priceID = req.PlanID
	}

	initReq := onboarding.InitRequest{
		UserID:  userID,
		PlanID:  req.PlanID,
		PriceID: priceID,
		OrgName: req.OrgName,
	}

	org, err := r.onboardingSvc.InitializeOrganization(c.Request.Context(), initReq)
	if err != nil {
		r.logger.Error("organization_initialization_failed",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.String("org_name", req.OrgName),
			zap.String("plan_id", req.PlanID),
			zap.String("request_id", c.GetString("request_id")),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "organization_initializing",
		"data":   org,
	})
}
