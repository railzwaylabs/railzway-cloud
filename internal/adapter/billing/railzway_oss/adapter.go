package railzwayoss

import (
	"context"
	"fmt"

	"github.com/smallbiznis/railzway-cloud/internal/domain/billing"
	"github.com/smallbiznis/railzway-cloud/pkg/railzwayclient"
)

type Adapter struct {
	client *railzwayclient.Client
}

func NewAdapter(client *railzwayclient.Client) *Adapter {
	return &Adapter{client: client}
}

func (a *Adapter) PauseSubscription(ctx context.Context, subscriptionID string) error {
	return a.client.PauseSubscription(ctx, subscriptionID)
}

func (a *Adapter) ResumeSubscription(ctx context.Context, subscriptionID string) error {
	return a.client.ResumeSubscription(ctx, subscriptionID)
}

func (a *Adapter) ChangePlan(ctx context.Context, params billing.ChangePlanParams) error {
	// The caller (UseCase) provides the NewPriceID?
	// OR does the domain provide the Tier, and WE map it to PriceID?
	// The Proposal said: "Move `tierToPriceID` logic here as configuration."
	// BUT the `ChangePlanParams` in domain/billing/interface.go accepts `NewPriceID`.
	// If the Domain knows about PriceIDs, that leaks OSS details.
	// REFINEMENT: The interface should accept TIER or we provide a helper to map it.
	// However, to satisfy the interface strictly defined in Phase 1 (which used NewPriceID), we must use NewPriceID.
	// But wait, I defined Phase 1 `interface.go` with `NewPriceID string`.
	// This means the Use Case must know the Price ID.
	// BETTER APPROACH: The Use Case should ask an "ConfigService" or "PriceMapper" for the ID,
	// OR this Adapter should offer a specialized method `GetPriceIDForTier`?
	// Let's stick to the interface implementation. The logic to MAP Tier->PriceID
	// should probably live in a `Config` or `Registry` passed to the **Use Case**,
	// or encapsulated here if we modify the interface.
	// Given the constraints, I will implement the interface AS IS.
	// For the "Move tierToPriceID logic" requirement, I will add a method `GetPriceID(tier string) string`
	// to this adapter, even if not strictly in the `Engine` interface, or just export a map.
	// Actually, the best place for Tier->PriceID mapping is likely a separate Config component/Use Case logic.
	// But for now, let's implement the interface methods.

	req := &railzwayclient.ChangePlanRequest{
		NewPriceID:        params.NewPriceID,
		ProrationBehavior: string(params.ProrationBehavior),
		EffectiveDate:     params.EffectiveDate,
		CancelAtPeriodEnd: params.CancelAtPeriodEnd,
	}

	return a.client.ChangePlan(ctx, params.SubscriptionID, req)
}

func (a *Adapter) ResolvePriceID(ctx context.Context, tier string) (string, error) {
	code, ok := PriceMap[tier]
	if !ok {
		return "", fmt.Errorf("no price mapping found for tier: %s", tier)
	}

	// Dynamic Resolution: Fetch Price ID from Code via API
	// TODO: Consider caching this to reduce API calls
	price, err := a.client.GetPriceByCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to resolve price id for code %s: %w", code, err)
	}

	return price.ID, nil
}

// PriceMap defines the configuration for Tier -> Price Code.
var PriceMap = map[string]string{
	"FREE_TRIAL": "free-trial-monthly",
	"STARTER":    "starter-monthly",
	"PRO":        "pro-monthly",
	"TEAM":       "team-monthly",
	"ENTERPRISE": "enterprise-monthly", // Assuming this pattern, or keep commented if not ready
}
