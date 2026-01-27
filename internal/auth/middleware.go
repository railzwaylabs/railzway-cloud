package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/internal/user"
)

type Middleware struct {
	cfg         *config.Config
	userService *user.Service
}

func NewMiddleware(cfg *config.Config, userService *user.Service) *Middleware {
	return &Middleware{
		cfg:         cfg,
		userService: userService,
	}
}

func (m *Middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_token"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate JWT (Stub for now, use proper validation lib/secret)
		// In reality, verify signature with m.cfg.AuthJWTSecret
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			// For strict mode, fail here. For stub, we'll try to proceed if we can extract claims.
			// c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			// return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_claims"})
			return
		}

		// Extract Auth Identity
		sub, _ := claims.GetSubject()
		email, _ := claims["email"].(string)

		if sub == "" || email == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_claims"})
			return
		}

		// Orchestration: Just-in-Time User Provisioning
		user, err := m.userService.EnsureUser(c.Request.Context(), sub, email)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to sync user: %v", err)})
			return
		}

		// Inject User into Context
		c.Set("user", user)
		c.Next()
	}
}
