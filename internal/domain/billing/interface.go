package billing

import (
	"context"
)

// ProrationBehavior defines how plan changes affect billing immediately.
type ProrationBehavior string

const (
	CreateProration ProrationBehavior = "CREATE_PRORATION"
	None            ProrationBehavior = "NONE"
)

// ChangePlanParams encapsulates parameters for changing a subscription plan.
type ChangePlanParams struct {
	SubscriptionID    string
	NewPriceID        string
	ProrationBehavior ProrationBehavior
	EffectiveDate     string // "immediate" or specific date
	CancelAtPeriodEnd bool
}

// PriceResolver defines the interface for resolving Price IDs for given tiers.
type PriceResolver interface {
	ResolvePriceID(ctx context.Context, tier string) (string, error)
}

// Engine defines the interface for interacting with the billing system (Railzway OSS).
type Engine interface {
	// PauseSubscription pauses billing for a subscription.
	PauseSubscription(ctx context.Context, subscriptionID string) error

	// ResumeSubscription resumes billing for a subscription.
	ResumeSubscription(ctx context.Context, subscriptionID string) error

	// ChangePlan updates the subscription plan/price.
	ChangePlan(ctx context.Context, params ChangePlanParams) error
}
