package railzwayclient

import (
	"context"
	"fmt"
	"net/http"
)

type Meter struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type UsageEvent struct {
	MeterCode  string                 `json:"meter_code"`
	CustomerID string                 `json:"customer_id"`
	Timestamp  string                 `json:"timestamp"`
	Value      float64                `json:"value"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// ListMeters lists all meters
func (c *Client) ListMeters(ctx context.Context) ([]Meter, error) {
	var meters []Meter
	err := c.doRequest(ctx, http.MethodGet, "/api/meters", nil, &meters)
	if err != nil {
		return nil, fmt.Errorf("failed to list meters: %w", err)
	}
	return meters, nil
}

// GetMeter retrieves a meter by ID
func (c *Client) GetMeter(ctx context.Context, id string) (*Meter, error) {
	path := fmt.Sprintf("/api/meters/%s", id)
	var meter Meter
	err := c.doRequest(ctx, http.MethodGet, path, nil, &meter)
	if err != nil {
		return nil, fmt.Errorf("failed to get meter: %w", err)
	}
	return &meter, nil
}

// CreateMeter creates a new meter
func (c *Client) CreateMeter(ctx context.Context, req any) (*Meter, error) {
	var meter Meter
	err := c.doRequest(ctx, http.MethodPost, "/api/meters", req, &meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter: %w", err)
	}
	return &meter, nil
}

// UpdateMeter updates an existing meter
func (c *Client) UpdateMeter(ctx context.Context, id string, req any) (*Meter, error) {
	path := fmt.Sprintf("/api/meters/%s", id)
	var meter Meter
	err := c.doRequest(ctx, http.MethodPatch, path, req, &meter)
	if err != nil {
		return nil, fmt.Errorf("failed to update meter: %w", err)
	}
	return &meter, nil
}

// DeleteMeter deletes a meter
func (c *Client) DeleteMeter(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/meters/%s", id)
	err := c.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete meter: %w", err)
	}
	return nil
}

// ReportUsage reports usage events
func (c *Client) ReportUsage(ctx context.Context, events []UsageEvent) error {
	err := c.doRequest(ctx, http.MethodPost, "/api/usage", events, nil)
	if err != nil {
		return fmt.Errorf("failed to report usage: %w", err)
	}
	return nil
}
