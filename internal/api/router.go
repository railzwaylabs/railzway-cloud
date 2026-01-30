package api

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/railzwaylabs/railzway-cloud/internal/api/middleware"
	"github.com/railzwaylabs/railzway-cloud/internal/auth"
	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/billing"
	"github.com/railzwaylabs/railzway-cloud/internal/onboarding"
	"github.com/railzwaylabs/railzway-cloud/internal/usecase/deployment"
	"github.com/railzwaylabs/railzway-cloud/internal/user"
	"github.com/railzwaylabs/railzway-cloud/pkg/railzwayclient"
	"go.uber.org/zap"
)

type Router struct {
	engine        *gin.Engine
	server        *http.Server
	cfg           *config.Config
	deployUC      *deployment.DeployUseCase
	lifecycleUC   *deployment.LifecycleUseCase
	upgradeUC     *deployment.UpgradeUseCase
	rolloutUC     *deployment.RolloutUseCase
	onboardingSvc *onboarding.Service
	userSvc       *user.Service
	sessionMgr    *auth.SessionManager
	billingEngine billing.Engine
	client        *railzwayclient.Client
	logger        *zap.Logger
}

func NewRouter(
	cfg *config.Config,
	deployUC *deployment.DeployUseCase,
	lifecycleUC *deployment.LifecycleUseCase,
	upgradeUC *deployment.UpgradeUseCase,
	rolloutUC *deployment.RolloutUseCase,
	onboardingSvc *onboarding.Service,
	userSvc *user.Service,
	sessionMgr *auth.SessionManager,
	billingEngine billing.Engine,
	client *railzwayclient.Client,
	logger *zap.Logger,
) *Router {
	// Disable GIN default logger
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Add custom middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Metrics())
	r.Use(middleware.Logger(logger))

	api := &Router{
		engine:        r,
		cfg:           cfg,
		deployUC:      deployUC,
		lifecycleUC:   lifecycleUC,
		upgradeUC:     upgradeUC,
		rolloutUC:     rolloutUC,
		onboardingSvc: onboardingSvc,
		userSvc:       userSvc,
		sessionMgr:    sessionMgr,
		billingEngine: billingEngine,
		client:        client,
		logger:        logger,
	}

	api.RegisterRoutes()
	return api
}

func (r *Router) RegisterRoutes() {
	// Simple health check
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Prometheus metrics endpoint
	r.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Auth Routes
	auth := r.engine.Group("/auth")
	{
		auth.GET("/login", r.sessionMgr.HandleLogin)
		auth.GET("/callback", r.sessionMgr.HandleCallback)
		auth.GET("/session", r.sessionMgr.HandleSession)
		auth.GET("/logout", r.sessionMgr.HandleLogout)
	}

	// Pricing Routes (Proxied to Railzway OSS)
	// These are public (for onboarding) or session-based? Onboarding is public-ish.
	// But ListPrices in client requires API Key which we have on server.
	// We expose them publicly or protected? Frontend calls them in onboarding (public/unauth).
	api := r.engine.Group("/api")
	{
		api.GET("/prices", r.ListPrices)
		api.GET("/price_amounts", r.ListPriceAmounts)
	}

	// User Routes (Protected)
	user := r.engine.Group("/user")
	user.Use(r.sessionMgr.Middleware())
	{
		user.GET("/organizations", r.GetUserOrganizations)
		user.GET("/profile", r.GetUserProfile)
		user.PUT("/profile", r.UpdateUserProfile)
		user.GET("/instance", r.GetInstanceStatus)
		user.GET("/instance/stream", r.StreamInstanceStatus)
		user.POST("/instance/deploy", r.DeployInstance)
		user.POST("/instance/start", r.StartInstance)
		user.POST("/instance/pause", r.PauseInstance)
		user.POST("/instance/stop", r.StopInstance)
		user.POST("/instance/upgrade", r.UpgradeInstance)
		user.POST("/instance/downgrade", r.DowngradeInstance)

		// Onboarding Endpoints (Protected)
		onboardGroup := user.Group("/onboarding")
		{
			onboardGroup.GET("/check-org-name", r.CheckOrgName)
			onboardGroup.POST("/initialize", r.InitializeOrganization)
		}
	}

	// Admin Routes (Protected by ADMIN_API_TOKEN)
	admin := r.engine.Group("/admin")
	admin.Use(r.adminAuth())
	{
		admin.POST("/rollout", r.RolloutVersion)
	}

	// SPA Fallback
	r.RegisterFallback()
}

func (r *Router) Run() error {
	r.server = &http.Server{
		Addr:         ":" + r.cfg.Port,
		Handler:      r.engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return r.server.ListenAndServe()
}

func (r *Router) adminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := strings.TrimSpace(r.cfg.AdminAPIToken)
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin_token_not_configured"})
			return
		}

		provided := strings.TrimSpace(c.GetHeader("X-Admin-Token"))
		if provided == "" {
			authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				provided = strings.TrimSpace(authHeader[7:])
			}
		}

		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

// Shutdown gracefully shuts down the HTTP server
func (r *Router) Shutdown(ctx context.Context) error {
	if r.server == nil {
		return nil
	}
	return r.server.Shutdown(ctx)
}

func (r *Router) resolveOrgID(c *gin.Context) (int64, bool) {
	// 1. Get User ID (Safety check)
	val, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, false
	}
	userID, ok := val.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, false
	}

	// 2. Determine OrgID
	// Priority: Query Param only.
	queryOrg := c.Query("org_id")
	if queryOrg != "" {
		parsed, err := strconv.ParseInt(queryOrg, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org_id"})
			return 0, false
		}
		owns, err := r.onboardingSvc.UserOwnsOrg(c.Request.Context(), userID, parsed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return 0, false
		}
		if !owns {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return 0, false
		}
		return parsed, true
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "org_id is required"})
	return 0, false
}
