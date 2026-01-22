package railzwayclient

import (
	"context"
	"fmt"
	"net/http"
)

// ChangePlanRequest represents a request to change subscription plan.
type ChangePlanRequest struct {
	NewPriceID        string `json:"new_price_id"`
	ProrationBehavior string `json:"proration_behavior"` // "CREATE_PRORATION", "NONE"
	EffectiveDate     string `json:"effective_date"`     // "immediate", or specific date
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end,omitempty"`
}

// PauseSubscription pauses an active subscription (stops billing).
func (c *Client) PauseSubscription(ctx context.Context, subscriptionID string) error {
	url := fmt.Sprintf("/api/subscriptions/%s/pause", subscriptionID)

	err := c.doRequest(ctx, http.MethodPost, url, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to pause subscription: %w", err)
	}

	return nil
}

// ResumeSubscription resumes a paused subscription (restarts billing).
func (c *Client) ResumeSubscription(ctx context.Context, subscriptionID string) error {
	url := fmt.Sprintf("/api/subscriptions/%s/resume", subscriptionID)

	err := c.doRequest(ctx, http.MethodPost, url, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to resume subscription: %w", err)
	}

	return nil
}

// ChangePlan changes the subscription plan (upgrade/downgrade).
func (c *Client) ChangePlan(ctx context.Context, subscriptionID string, req *ChangePlanRequest) error {
	url := fmt.Sprintf("/api/subscriptions/%s/change-plan", subscriptionID)

	err := c.doRequest(ctx, http.MethodPost, url, req, nil)
	if err != nil {
		return fmt.Errorf("failed to change plan: %w", err)
	}

	return nil
}

// CancelSubscription cancels a subscription.
func (c *Client) CancelSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool) error {
	url := fmt.Sprintf("/api/subscriptions/%s/cancel", subscriptionID)

	reqBody := map[string]interface{}{
		"cancel_at_period_end": cancelAtPeriodEnd,
	}

	err := c.doRequest(ctx, http.MethodPost, url, reqBody, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}
