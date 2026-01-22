package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Router) GetUserOrganizations(c *gin.Context) {
	val, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(int64)

	orgs, err := r.onboardingSvc.GetOrganizationsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orgs})
}
