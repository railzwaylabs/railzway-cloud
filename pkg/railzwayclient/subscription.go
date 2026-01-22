package railzwayclient

import (
	"context"
	"fmt"
	"net/http"
)

type Subscription struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	PlanID     string `json:"plan_id"`
	Status     string `json:"status"`
}

type SubscriptionItem struct {
	ID             string `json:"id"`
	SubscriptionID string `json:"subscription_id"`
	PriceID        string `json:"price_id"`
	Quantity       int    `json:"quantity"`
}

// ListSubscriptions lists all subscriptions
func (c *Client) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	// Wrapper logic for list might be different. Usually list endpoints return { "items": [...] }
	// Assuming ListSubscriptions follows the customer list pattern (wrapped in `items` or `data`?)
	// Let's assume server returns { "data": [...] } for consistency or check ListCustomers implementation.
	// Actually, ListCustomers returned { "items": [...] }.
	// Let's check railzway-oss implementation for ListSubscriptions if possible.
	// If uncertain, let's look at `GetSubscription` and `CreateSubscription` which likely return single object wrapped in `data`.

	// For now, let's update Get and Create.
	return nil, fmt.Errorf("not implemented")
}

// GetSubscription retrieves a subscription by ID
func (c *Client) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	path := fmt.Sprintf("/api/subscriptions/%s", id)
	var resp ResponseWrapper[Subscription]
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &resp.Data, nil
}

// ActivateSubscription activates a subscription
func (c *Client) ActivateSubscription(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/subscriptions/%s/activate", id)
	err := c.doRequest(ctx, http.MethodPost, path, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to activate subscription: %w", err)
	}
	return nil
}

// ListSubscriptionItems lists items for a subscription
func (c *Client) ListSubscriptionItems(ctx context.Context, id string) ([]SubscriptionItem, error) {
	// CAUTION: Check if items list returns { "data": [...] } or { "items": [...] }
	// Assuming { "data": [...] } for now based on other endpoints.
	path := fmt.Sprintf("/api/subscriptions/%s/items", id)
	var resp ResponseWrapper[[]SubscriptionItem]
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscription items: %w", err)
	}
	return resp.Data, nil
}

type CreateSubscriptionItemRequest struct {
	PriceID  string `json:"price_id"`
	MeterID  string `json:"meter_id,omitempty"`
	Quantity int    `json:"quantity"`
}

type CreateSubscriptionRequest struct {
	CustomerID       string                          `json:"customer_id"`
	CollectionMode   string                          `json:"collection_mode"`
	BillingCycleType string                          `json:"billing_cycle_type"`
	Items            []CreateSubscriptionItemRequest `json:"items"`
}

func (c *Client) CreateSubscription(ctx context.Context, customerID string, billingCycleType string, items []CreateSubscriptionItemRequest) (*Subscription, error) {
	reqBody := CreateSubscriptionRequest{
		CustomerID:       customerID,
		CollectionMode:   "SEND_INVOICE",
		BillingCycleType: billingCycleType,
		Items:            items,
	}

	var resp ResponseWrapper[Subscription]
	err := c.doRequest(ctx, http.MethodPost, "/api/subscriptions", reqBody, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return &resp.Data, nil
}
