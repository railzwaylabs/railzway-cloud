package authclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	if strings.TrimSpace(c.cfg.AdminToken) == "" {
		return nil, fmt.Errorf("auth service admin token missing")
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
	httpReq.Header.Set("X-Admin-Token", c.cfg.AdminToken)
	httpReq.Header.Set("X-Tenant-ID", c.cfg.TenantSlug)

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
	if strings.TrimSpace(out.ClientID) == "" || strings.TrimSpace(out.ClientSecret) == "" {
		return nil, fmt.Errorf("auth client response missing credentials")
	}

	return &out, nil
}
