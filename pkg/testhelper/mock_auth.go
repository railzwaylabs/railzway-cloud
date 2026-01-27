package testhelper

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockAuthServer creates a mock OAuth server for testing
type MockAuthServer struct {
	Server          *httptest.Server
	TokenRequests   int
	ClientRequests  int
	ShouldFailToken bool
	ShouldFailClient bool
}

// NewMockAuthServer creates a new mock auth server
func NewMockAuthServer(t *testing.T) *MockAuthServer {
	mock := &MockAuthServer{}

	mux := http.NewServeMux()

	// Token endpoint
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		mock.TokenRequests++
		if mock.ShouldFailToken {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"invalid_client"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`))
	})

	// OAuth client creation endpoint
	mux.HandleFunc("/admin/oauth/clients", func(w http.ResponseWriter, r *http.Request) {
		mock.ClientRequests++
		if mock.ShouldFailClient {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal_error"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"client_id":"test-client-id",
			"client_secret":"test-client-secret",
			"redirect_uris":["http://localhost/callback"],
			"scopes":["openid","email"],
			"grants":["authorization_code"],
			"token_endpoint_auth_methods":["client_secret_post"]
		}`))
	})

	mock.Server = httptest.NewServer(mux)
	t.Cleanup(mock.Server.Close)

	return mock
}

// URL returns the base URL of the mock server
func (m *MockAuthServer) URL() string {
	return m.Server.URL
}
