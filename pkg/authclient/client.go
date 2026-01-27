package authclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	cfg  Config
	http *http.Client
}

func NewFromEnv() *Client {
	return New(LoadFromEnv())
}

func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: timeout},
	}
}

type EnsureOAuthClientRequest struct {
	ExternalOrgID int64
	RedirectURIs  []string
}

type OAuthClient struct {
	ClientID                 string   `json:"client_id"`
	ClientSecret             string   `json:"client_secret"`
	RedirectURIs             []string `json:"redirect_uris"`
	Scopes                   []string `json:"scopes"`
	Grants                   []string `json:"grants"`
	TokenEndpointAuthMethods []string `json:"token_endpoint_auth_methods"`
}

type upsertOAuthClientPayload struct {
	ExternalOrgID string   `json:"external_org_id"`
	RedirectURIs  []string `json:"redirect_uris"`
}

func (c *Client) EnsureOAuthClient(ctx context.Context, req EnsureOAuthClientRequest) (*OAuthClient, error) {
	if c == nil {
		return nil, fmt.Errorf("auth client not configured")
	}
	base := strings.TrimRight(c.cfg.BaseURL, "/")
	if base == "" {
		return nil, fmt.Errorf("auth service url missing")
	}
	if strings.TrimSpace(c.cfg.TenantSlug) == "" {
		return nil, fmt.Errorf("auth service tenant missing")
	}
	if req.ExternalOrgID == 0 {
		return nil, fmt.Errorf("external org id missing")
	}
	if len(req.RedirectURIs) == 0 {
		return nil, fmt.Errorf("redirect uris missing")
	}

	payload := upsertOAuthClientPayload{
		ExternalOrgID: fmt.Sprintf("%d", req.ExternalOrgID),
		RedirectURIs:  req.RedirectURIs,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode auth client payload: %w", err)
	}

	url := fmt.Sprintf("%s/admin/oauth/clients", base)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build auth client request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", c.cfg.TenantSlug)

	token, err := c.clientCredentialsToken(ctx, base)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("auth client request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("auth client error (%d): %s", resp.StatusCode, string(raw))
	}

	var out OAuthClient
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode auth client response: %w", err)
	}
	if strings.TrimSpace(out.ClientID) == "" {
		return nil, fmt.Errorf("auth client response missing client_id")
	}

	return &out, nil
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (c *Client) clientCredentialsToken(ctx context.Context, base string) (string, error) {
	clientID := strings.TrimSpace(c.cfg.ClientID)
	clientSecret := strings.TrimSpace(c.cfg.ClientSecret)
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("auth service client credentials missing")
	}
	scope := strings.TrimSpace(c.cfg.ClientScope)
	if scope == "" {
		scope = "admin"
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", scope)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	tokenURL := fmt.Sprintf("%s/token", strings.TrimRight(base, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Tenant-ID", c.cfg.TenantSlug)
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("token request error (%d): %s", resp.StatusCode, string(raw))
	}

	var out tokenResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		if out.Error != "" {
			return "", fmt.Errorf("token error: %s", out.ErrorDescription)
		}
		return "", fmt.Errorf("token response missing access_token")
	}
	return out.AccessToken, nil
}
