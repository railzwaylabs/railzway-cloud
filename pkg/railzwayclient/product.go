package railzwayclient

import (
	"context"
	"fmt"
	"net/http"
)

type Product struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateProductRequest struct {
	Code        string         `json:"code"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Active      *bool          `json:"active,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

func (c *Client) ListProducts(ctx context.Context) ([]Product, error) {
	var out []Product
	err := c.doRequest(ctx, http.MethodGet, "/api/products", nil, &out)
	return out, err
}

func (c *Client) CreateProduct(ctx context.Context, req any) (*Product, error) {
	var out Product
	err := c.doRequest(ctx, http.MethodPost, "/api/products", req, &out)
	return &out, err
}

// GetProduct retrieves a product by ID
func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	path := fmt.Sprintf("/api/products/%s", id)
	var product Product
	err := c.doRequest(ctx, http.MethodGet, path, nil, &product)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}

// ArchiveProduct archives a product
func (c *Client) ArchiveProduct(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/products/%s/archive", id)
	err := c.doRequest(ctx, http.MethodPost, path, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to archive product: %w", err)
	}
	return nil
}
