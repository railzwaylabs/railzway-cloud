package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Auth0TokenResponse represents the response from Auth0 token endpoint
type Auth0TokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Auth0Claims represents the claims in Auth0 ID token
type Auth0Claims struct {
	Sub   string `json:"sub"`   // User ID (authID)
	Email string `json:"email"` // Email
	Name  string `json:"name"`  // Full name
	jwt.RegisteredClaims
}

// HandleCallback processes the OAuth2 callback.
// Exchanges the authorization code for tokens and creates a session.
func (m *SessionManager) HandleCallback(c *gin.Context) {
	logger := m.logger.With(
		zap.String("handler", "auth.callback"),
		zap.String("request_id", c.GetString("request_id")),
	)

	code := c.Query("code")
	if code == "" {
		logger.Warn("missing_authorization_code")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_code"})
		return
	}

	// 1. Exchange code for tokens
	tokenResp, err := m.exchangeCodeForTokens(c.Request.Context(), code)
	if err != nil {
		logger.Error("token_exchange_failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_exchange_failed"})
		return
	}

	// 2. Parse ID token to get user claims
	claims, err := m.resolveUserClaims(c.Request.Context(), tokenResp)
	if err != nil {
		logger.Error("user_claims_resolution_failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid_id_token"})
		return
	}

	logger.Info("user_authenticated",
		zap.String("auth_id", claims.Sub),
		zap.String("email", claims.Email),
	)

	// 3. Sync User (Just-In-Time Provisioning)
	user, err := m.userService.EnsureUser(c.Request.Context(), claims.Sub, claims.Email)
	if err != nil {
		logger.Error("user_sync_failed",
			zap.Error(err),
			zap.String("auth_id", claims.Sub),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_sync_user"})
		return
	}

	logger.Info("user_synced", zap.Int64("user_id", user.ID))

	if user.FirstName == "" && user.LastName == "" {
		firstName, lastName := splitName(claims.Name)
		if firstName != "" || lastName != "" {
			updated, err := m.userService.UpdateProfile(c.Request.Context(), user.ID, firstName, lastName)
			if err != nil {
				logger.Warn("user_profile_sync_failed", zap.Error(err))
			} else {
				user = updated
			}
		}
	}

	// 4. Create Session
	if err := m.CreateSession(c, user.ID); err != nil {
		logger.Error("session_creation_failed",
			zap.Error(err),
			zap.Int64("user_id", user.ID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_session"})
		return
	}

	logger.Info("session_created", zap.Int64("user_id", user.ID))

	// 5. Redirect to Dashboard
	c.Redirect(http.StatusFound, "/")
}

// exchangeCodeForTokens exchanges the authorization code for access and ID tokens
func (m *SessionManager) exchangeCodeForTokens(ctx context.Context, code string) (*Auth0TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/token", m.cfg.OAuth2URI)

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", m.cfg.OAuth2ClientID)
	data.Set("client_secret", m.cfg.OAuth2ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", m.cfg.OAuth2CallbackURL)

	// Make request
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp Auth0TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

func splitName(raw string) (string, string) {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

type UserInfoClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (m *SessionManager) resolveUserClaims(ctx context.Context, tokenResp *Auth0TokenResponse) (*Auth0Claims, error) {
	if tokenResp == nil {
		return nil, fmt.Errorf("token response missing")
	}

	if strings.TrimSpace(tokenResp.IDToken) != "" {
		if claims, err := m.parseIDToken(tokenResp.IDToken); err == nil {
			if strings.TrimSpace(claims.Sub) != "" && strings.TrimSpace(claims.Email) != "" {
				return claims, nil
			}
		}
	}

	info, err := m.fetchUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(info.Sub) == "" || strings.TrimSpace(info.Email) == "" {
		return nil, fmt.Errorf("userinfo missing required claims")
	}

	return &Auth0Claims{
		Sub:   info.Sub,
		Email: info.Email,
		Name:  info.Name,
	}, nil
}

func (m *SessionManager) fetchUserInfo(ctx context.Context, accessToken string) (*UserInfoClaims, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("access token missing")
	}
	base := strings.TrimRight(m.cfg.OAuth2URI, "/")
	if base == "" {
		return nil, fmt.Errorf("oauth2 uri missing")
	}

	candidates := []string{
		fmt.Sprintf("%s/userinfo", base),
		fmt.Sprintf("%s/oauth/userinfo", base),
	}

	var lastErr error
	for _, target := range candidates {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("userinfo status %d: %s", resp.StatusCode, string(body))
			continue
		}

		var claims UserInfoClaims
		if err := json.Unmarshal(body, &claims); err != nil {
			lastErr = err
			continue
		}

		return &claims, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("userinfo request failed")
	}
	return nil, lastErr
}

// parseIDToken parses the ID token JWT and extracts claims
// Note: For production, you should verify the JWT signature
func (m *SessionManager) parseIDToken(idToken string) (*Auth0Claims, error) {
	// Parse without verification (for development)
	// In production, use jwt.ParseWithClaims with proper key verification
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, &Auth0Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse id_token: %w", err)
	}

	claims, ok := token.Claims.(*Auth0Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	return claims, nil
}

// HandleLogin initiates the OAuth2 flow by redirecting the user to the provider.
func (m *SessionManager) HandleLogin(c *gin.Context) {
	authBase := m.cfg.OAuth2URI
	clientID := m.cfg.OAuth2ClientID
	redirectURI := m.cfg.OAuth2CallbackURL
	scope := "openid profile email"

	// Construct authorization URL
	target := fmt.Sprintf("%s/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s",
		authBase, clientID, url.QueryEscape(redirectURI), url.QueryEscape(scope))

	c.Redirect(http.StatusFound, target)
}
