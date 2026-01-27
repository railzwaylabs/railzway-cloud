package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxNameLength = 100

type userProfileResponse struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type updateUserProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (r *Router) GetUserProfile(c *gin.Context) {
	userID, ok := resolveUserID(c)
	if !ok {
		return
	}

	user, err := r.userSvc.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_load_user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userProfileResponse{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}})
}

func (r *Router) UpdateUserProfile(c *gin.Context) {
	userID, ok := resolveUserID(c)
	if !ok {
		return
	}

	var payload updateUserProfileRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	firstName := strings.TrimSpace(payload.FirstName)
	lastName := strings.TrimSpace(payload.LastName)

	if len(firstName) > maxNameLength || len(lastName) > maxNameLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name_too_long"})
		return
	}

	user, err := r.userSvc.UpdateProfile(c.Request.Context(), userID, firstName, lastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userProfileResponse{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}})
}

func resolveUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, false
	}
	userID, ok := val.(int64)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, false
	}
	return userID, true
}
