package authclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_EnsureOAuthClient_Success(t *testing.T) {
	// Setup mock auth server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			// Mock token endpoint
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			assert.Equal(t, "test-tenant", r.Header.Get("X-Tenant-ID"))
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokenResponse{
				AccessToken: "test-access-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			})

		case "/admin/oauth/clients":
			// Mock OAuth client creation endpoint
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "test-tenant", r.Header.Get("X-Tenant-ID"))
			assert.Equal(t, "Bearer test-access-token", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(OAuthClient{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURIs: []string{"https://example.com/callback"},
			})

		default:
			t.Fatalf("unexpected request: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	// Create client
	client := New(Config{
		BaseURL:      server.URL,
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	// Execute
	result, err := client.EnsureOAuthClient(context.Background(), EnsureOAuthClientRequest{
		ExternalOrgID: 123,
		RedirectURIs:  []string{"https://example.com/callback"},
	})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "test-client-id", result.ClientID)
	assert.Equal(t, "test-client-secret", result.ClientSecret)
	assert.Equal(t, []string{"https://example.com/callback"}, result.RedirectURIs)
}

func TestClient_EnsureOAuthClient_NilClient(t *testing.T) {
	var client *Client
	_, err := client.EnsureOAuthClient(context.Background(), EnsureOAuthClientRequest{
		ExternalOrgID: 123,
		RedirectURIs:  []string{"https://example.com/callback"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth client not configured")
}

func TestClient_EnsureOAuthClient_MissingBaseURL(t *testing.T) {
	client := New(Config{
		BaseURL:      "",
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.EnsureOAuthClient(context.Background(), EnsureOAuthClientRequest{
		ExternalOrgID: 123,
		RedirectURIs:  []string{"https://example.com/callback"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth service url missing")
}

func TestClient_EnsureOAuthClient_MissingTenantSlug(t *testing.T) {
	client := New(Config{
		BaseURL:      "https://auth.example.com",
		TenantSlug:   "",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.EnsureOAuthClient(context.Background(), EnsureOAuthClientRequest{
		ExternalOrgID: 123,
		RedirectURIs:  []string{"https://example.com/callback"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth service tenant missing")
}

func TestClient_EnsureOAuthClient_MissingExternalOrgID(t *testing.T) {
	client := New(Config{
		BaseURL:      "https://auth.example.com",
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.EnsureOAuthClient(context.Background(), EnsureOAuthClientRequest{
		ExternalOrgID: 0,
		RedirectURIs:  []string{"https://example.com/callback"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "external org id missing")
}

func TestClient_EnsureOAuthClient_MissingRedirectURIs(t *testing.T) {
	client := New(Config{
		BaseURL:      "https://auth.example.com",
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.EnsureOAuthClient(context.Background(), EnsureOAuthClientRequest{
		ExternalOrgID: 123,
		RedirectURIs:  []string{},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redirect uris missing")
}

func TestClient_EnsureOAuthClient_TokenRequestFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error":             "invalid_client",
				"error_description": "Invalid client credentials",
			})
		}
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:      server.URL,
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.EnsureOAuthClient(context.Background(), EnsureOAuthClientRequest{
		ExternalOrgID: 123,
		RedirectURIs:  []string{"https://example.com/callback"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token request error (401)")
}

func TestClient_EnsureOAuthClient_OAuthClientCreationFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			json.NewEncoder(w).Encode(tokenResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
			})
		case "/admin/oauth/clients":
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Insufficient permissions"))
		}
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:      server.URL,
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.EnsureOAuthClient(context.Background(), EnsureOAuthClientRequest{
		ExternalOrgID: 123,
		RedirectURIs:  []string{"https://example.com/callback"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth client error (403)")
}

func TestClient_EnsureOAuthClient_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			json.NewEncoder(w).Encode(tokenResponse{
				AccessToken: "test-token",
			})
		case "/admin/oauth/clients":
			w.Write([]byte("invalid json"))
		}
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:      server.URL,
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.EnsureOAuthClient(context.Background(), EnsureOAuthClientRequest{
		ExternalOrgID: 123,
		RedirectURIs:  []string{"https://example.com/callback"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode auth client response")
}

func TestClient_EnsureOAuthClient_MissingClientID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			json.NewEncoder(w).Encode(tokenResponse{
				AccessToken: "test-token",
			})
		case "/admin/oauth/clients":
			// Return response without client_id
			json.NewEncoder(w).Encode(OAuthClient{
				ClientSecret: "test-secret",
			})
		}
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:      server.URL,
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.EnsureOAuthClient(context.Background(), EnsureOAuthClientRequest{
		ExternalOrgID: 123,
		RedirectURIs:  []string{"https://example.com/callback"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth client response missing client_id")
}

func TestClient_ClientCredentialsToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/token", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		// Verify form data
		r.ParseForm()
		assert.Equal(t, "client_credentials", r.Form.Get("grant_type"))
		assert.Equal(t, "admin", r.Form.Get("scope"))

		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:      server.URL,
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	token, err := client.clientCredentialsToken(context.Background(), server.URL)

	require.NoError(t, err)
	assert.Equal(t, "test-access-token", token)
}

func TestClient_ClientCredentialsToken_MissingCredentials(t *testing.T) {
	client := New(Config{
		BaseURL:    "https://auth.example.com",
		TenantSlug: "test-tenant",
		// Missing ClientID and ClientSecret
	})

	_, err := client.clientCredentialsToken(context.Background(), "https://auth.example.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth service client credentials missing")
}

func TestClient_ClientCredentialsToken_CustomScope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		assert.Equal(t, "custom-scope", r.Form.Get("scope"))

		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "test-token",
		})
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:      server.URL,
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		ClientScope:  "custom-scope",
	})

	_, err := client.clientCredentialsToken(context.Background(), server.URL)
	require.NoError(t, err)
}

func TestClient_ClientCredentialsToken_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			Error:            "invalid_grant",
			ErrorDescription: "Invalid client credentials",
		})
	}))
	defer server.Close()

	client := New(Config{
		BaseURL:      server.URL,
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.clientCredentialsToken(context.Background(), server.URL)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid client credentials")
}

func TestNew_DefaultTimeout(t *testing.T) {
	client := New(Config{
		BaseURL: "https://auth.example.com",
	})

	assert.NotNil(t, client)
	assert.NotNil(t, client.http)
	// Default timeout should be 10 seconds
	assert.Equal(t, 10*1000000000, int(client.http.Timeout)) // 10 seconds in nanoseconds
}

func TestNew_CustomTimeout(t *testing.T) {
	client := New(Config{
		BaseURL: "https://auth.example.com",
		Timeout: 5000000000, // 5 seconds
	})

	assert.NotNil(t, client)
	assert.Equal(t, 5000000000, int(client.http.Timeout))
}
