package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/smallbiznis/railzway-cloud/internal/config"
	"github.com/smallbiznis/railzway-cloud/internal/user"
	"go.uber.org/zap"
)

const (
	SessionName = "railzway_cloud_session"
	UserIDXKey  = "user_id"
)

type SessionManager struct {
	store       *sessions.CookieStore
	userService *user.Service
	cfg         *config.Config
	logger      *zap.Logger
}

func NewSessionManager(cfg *config.Config, userService *user.Service, logger *zap.Logger) *SessionManager {
	store := sessions.NewCookieStore([]byte(cfg.AuthJWTSecret)) // Reusing JWT secret for session encryption
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   cfg.AuthCookieSecure,
	}

	return &SessionManager{
		store:       store,
		userService: userService,
		cfg:         cfg,
		logger:      logger,
	}
}

func (m *SessionManager) CreateSession(c *gin.Context, userID int64) error {
	session, _ := m.store.Get(c.Request, SessionName)
	session.Values[UserIDXKey] = userID

	// Set a non-HttpOnly cookie for frontend visibility (Marketing site Navbar)
	// This does not contain sensitive data, just a flag.
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "railzway_is_logged_in",
		Value:    "true",
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: false, // Accessible by JS
		Secure:   m.cfg.AuthCookieSecure,
	})

	return session.Save(c.Request, c.Writer)
}

func (m *SessionManager) HandleSession(c *gin.Context) {
	session, err := m.store.Get(c.Request, SessionName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "session_invalid"})
		return
	}

	rawID, ok := session.Values[UserIDXKey]
	if !ok || rawID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "unauthorized"})
		return
	}

	var userID int64
	switch value := rawID.(type) {
	case int64:
		userID = value
	case int:
		userID = int64(value)
	case float64:
		userID = int64(value)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "unauthorized"})
		return
	}

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"authenticated": true, "user_id": userID})
}

func (m *SessionManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := m.store.Get(c.Request, SessionName)
		if err != nil {
			// Session invalid/expired by cookie store
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session_invalid"})
			return
		}

		rawID, ok := session.Values[UserIDXKey]
		if !ok || rawID == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var userID int64
		switch value := rawID.(type) {
		case int64:
			userID = value
		case int:
			userID = int64(value)
		case float64:
			userID = int64(value)
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Optional: Hydrate User into Context (may skip querying DB if only ID needed)
		// For now, let's put simple struct
		c.Set("UserID", userID)
		c.Next()
	}
}
