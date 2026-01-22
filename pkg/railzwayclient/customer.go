package railzwayclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Customer struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	ExternalID string `json:"external_id"`
}

type ListCustomersParams struct {
	Name        string
	Email       string
	ExternalID  string
	Currency    string
	CreatedFrom string
	CreatedTo   string
	PageToken   string
	PageSize    int
}

type customerListResponse struct {
	Items         []Customer `json:"items"`
	NextPageToken string     `json:"next_page_token"`
}

// ListCustomers lists customers with optional filtering
func (c *Client) ListCustomers(ctx context.Context, params ListCustomersParams) ([]Customer, error) {
	path := "/api/customers"

	// Build query string
	query := url.Values{}
	if params.Name != "" {
		query.Set("name", params.Name)
	}
	if params.Email != "" {
		query.Set("email", params.Email)
	}
	if params.ExternalID != "" {
		query.Set("external_id", params.ExternalID)
	}
	if params.Currency != "" {
		query.Set("currency", params.Currency)
	}
	if params.CreatedFrom != "" {
		query.Set("created_from", params.CreatedFrom)
	}
	if params.CreatedTo != "" {
		query.Set("created_to", params.CreatedTo)
	}
	if params.PageToken != "" {
		query.Set("page_token", params.PageToken)
	}
	if params.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(params.PageSize))
	}

	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var resp customerListResponse
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// GetCustomer retrieves a customer by ID
func (c *Client) GetCustomer(ctx context.Context, id string) (*Customer, error) {
	path := fmt.Sprintf("/api/customers/%s", id)
	var resp ResponseWrapper[Customer]
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	return &resp.Data, nil
}

// EnsureCustomer ensures a Customer exists in Railzway OSS.
// Idempotent by externalID.
// If customer exists → return it.
// If not → create it.
func (c *Client) EnsureCustomer(
	ctx context.Context,
	name string,
	email string,
	externalID string,
) (*Customer, error) {

	// 1. Try to find existing customer by external_id
	customer, err := c.getCustomerByExternalID(ctx, externalID)
	if err == nil && customer != nil {
		return customer, nil
	}

	if err != nil && !errors.Is(err, ErrNotFound) {
		// Real error (OSS down, permission, etc)
		return nil, err
	}

	// 2. Not found → create customer
	createReq := map[string]any{
		"email": email,
		"name":  name,
		"metadata": map[string]any{
			"external_id": externalID,
			"source":      "railzway-cloud",
		},
	}

	var resp ResponseWrapper[Customer]
	err = c.doRequest(ctx, http.MethodPost, "/api/customers", createReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// ErrNotFound is returned when a resource is not found in OSS.
var ErrNotFound = errors.New("resource not found")

func (c *Client) getCustomerByExternalID(ctx context.Context, externalID string) (*Customer, error) {
	customers, err := c.ListCustomers(ctx, ListCustomersParams{
		ExternalID: externalID,
		PageSize:   1,
	})
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, ErrNotFound
	}

	return &customers[0], nil
}

func (c *Client) CreateCustomer(ctx context.Context, email, name, externalID string) (*Customer, error) {
	reqBody := map[string]string{
		"email":       email,
		"name":        name,
		"external_id": externalID,
	}

	var resp Customer
	if err := c.doRequest(ctx, http.MethodPost, "/api/customers", reqBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return &resp, nil
}
