package railzwayclient

import (
	"context"
	"fmt"
	"net/http"
)

// Price represents a price entity
type Price struct {
	ID              string         `json:"id"`
	ProductID       string         `json:"product_id"`
	Code            string         `json:"code"`
	Name            string         `json:"name"`
	PricingModel    string         `json:"pricing_model"`
	BillingMode     string         `json:"billing_mode"`
	BillingInterval string         `json:"billing_interval"`
	BillingUnit     *string        `json:"billing_unit,omitempty"`
	TaxBehavior     string         `json:"tax_behavior"`
	Active          bool           `json:"active"`
	Metadata        map[string]any `json:"metadata"`
}

// PriceAmount represents a price amount
type PriceAmount struct {
	ID              string `json:"id"`
	PriceID         string `json:"price_id"`
	UnitAmountCents int64  `json:"unit_amount_cents"`
}

// PriceTier represents a pricing tier
type PriceTier struct {
	ID      string `json:"id"`
	PriceID string `json:"price_id"`
	UpTo    int64  `json:"up_to"`
	FlatFee int64  `json:"flat_fee"`
	UnitFee int64  `json:"unit_fee"`
}

// Pricing represents a pricing configuration
type Pricing struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// === Prices ===

// PriceListOptions options for listing prices
type PriceListOptions struct {
	Code string `url:"code,omitempty"`
}

// ListPrices lists all prices
func (c *Client) ListPrices(ctx context.Context, opts *PriceListOptions) ([]Price, error) {
	path := "/api/prices"
	if opts != nil && opts.Code != "" {
		path += "?code=" + opts.Code
	}

	var resp ResponseWrapper[[]Price]
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list prices: %w", err)
	}
	return resp.Data, nil
}

// GetPriceByCode retrieves a price by its unique code
func (c *Client) GetPriceByCode(ctx context.Context, code string) (*Price, error) {
	prices, err := c.ListPrices(ctx, &PriceListOptions{Code: code})
	if err != nil {
		return nil, err
	}
	if len(prices) == 0 {
		return nil, fmt.Errorf("price with code %s not found", code)
	}
	return &prices[0], nil
}

// GetPrice retrieves a price by ID
func (c *Client) GetPrice(ctx context.Context, id string) (*Price, error) {
	path := fmt.Sprintf("/api/prices/%s", id)
	var price Price
	err := c.doRequest(ctx, http.MethodGet, path, nil, &price)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}
	return &price, nil
}

// CreatePrice creates a new price
func (c *Client) CreatePrice(ctx context.Context, req any) (*Price, error) {
	var price Price
	err := c.doRequest(ctx, http.MethodPost, "/api/prices", req, &price)
	if err != nil {
		return nil, fmt.Errorf("failed to create price: %w", err)
	}
	return &price, nil
}

// === Price Amounts ===

// ListPriceAmounts lists price amounts, optionally filtered by price ID
func (c *Client) ListPriceAmounts(ctx context.Context, priceID string) ([]PriceAmount, error) {
	path := "/api/price_amounts"
	if priceID != "" {
		path += "?price_id=" + priceID
	}

	var resp ResponseWrapper[[]PriceAmount]
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list price amounts: %w", err)
	}
	return resp.Data, nil
}

// GetPriceAmount retrieves a price amount by ID
func (c *Client) GetPriceAmount(ctx context.Context, id string) (*PriceAmount, error) {
	path := fmt.Sprintf("/api/price_amounts/%s", id)
	var amount PriceAmount
	err := c.doRequest(ctx, http.MethodGet, path, nil, &amount)
	if err != nil {
		return nil, fmt.Errorf("failed to get price amount: %w", err)
	}
	return &amount, nil
}

// CreatePriceAmount creates a new price amount
func (c *Client) CreatePriceAmount(ctx context.Context, req any) (*PriceAmount, error) {
	var amount PriceAmount
	err := c.doRequest(ctx, http.MethodPost, "/api/price_amounts", req, &amount)
	if err != nil {
		return nil, fmt.Errorf("failed to create price amount: %w", err)
	}
	return &amount, nil
}

// === Price Tiers ===

// ListPriceTiers lists all price tiers
func (c *Client) ListPriceTiers(ctx context.Context) ([]PriceTier, error) {
	var resp ResponseWrapper[[]PriceTier]
	err := c.doRequest(ctx, http.MethodGet, "/api/price_tiers", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list price tiers: %w", err)
	}
	return resp.Data, nil
}

// GetPriceTier retrieves a price tier by ID
func (c *Client) GetPriceTier(ctx context.Context, id string) (*PriceTier, error) {
	path := fmt.Sprintf("/api/price_tiers/%s", id)
	var tier PriceTier
	err := c.doRequest(ctx, http.MethodGet, path, nil, &tier)
	if err != nil {
		return nil, fmt.Errorf("failed to get price tier: %w", err)
	}
	return &tier, nil
}

// CreatePriceTier creates a new price tier
func (c *Client) CreatePriceTier(ctx context.Context, req any) (*PriceTier, error) {
	var tier PriceTier
	err := c.doRequest(ctx, http.MethodPost, "/api/price_tiers", req, &tier)
	if err != nil {
		return nil, fmt.Errorf("failed to create price tier: %w", err)
	}
	return &tier, nil
}

// === Pricings ===

// ListPricings lists all pricings
func (c *Client) ListPricings(ctx context.Context) ([]Pricing, error) {
	var resp ResponseWrapper[[]Pricing]
	err := c.doRequest(ctx, http.MethodGet, "/api/pricings", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list pricings: %w", err)
	}
	return resp.Data, nil
}

// GetPricing retrieves a pricing by ID
func (c *Client) GetPricing(ctx context.Context, id string) (*Pricing, error) {
	path := fmt.Sprintf("/api/pricings/%s", id)
	var pricing Pricing
	err := c.doRequest(ctx, http.MethodGet, path, nil, &pricing)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing: %w", err)
	}
	return &pricing, nil
}

// CreatePricing creates a new pricing
func (c *Client) CreatePricing(ctx context.Context, req any) (*Pricing, error) {
	var pricing Pricing
	err := c.doRequest(ctx, http.MethodPost, "/api/pricings", req, &pricing)
	if err != nil {
		return nil, fmt.Errorf("failed to create pricing: %w", err)
	}
	return &pricing, nil
}
