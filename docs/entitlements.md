# Product Entitlements (Railzway Cloud)

Railzway Cloud seeds Railzway OSS products with entitlement metadata during startup.
Each product is created under `/api/products` with `metadata.entitlements`.
Prices are seeded as flat, monthly, USD amounts under `/api/prices` with `/api/price_amounts`.

Products created:
- Evaluation
- Hobby
- Production
- Performance

Price codes:
- evaluation-monthly
- hobby-monthly
- production-monthly
- performance-monthly

Metadata shape:

```json
{
  "entitlements": {
    "billing.customers.max": 3,
    "billing.subscriptions.max": 3,
    "billing.subscription_items.max": 10,
    "billing.usage_events.monthly": 50000,
    "billing.invoices.monthly": 1,
    "billing.billing_cycles.concurrent": 1,
    "infra.namespaces.max": 1,
    "infra.retention_days": 7,
    "infra.isolation_level": "shared",
    "support.level": "community"
  }
}
```

## Evaluation ($0)
- billing.customers.max: 3
- billing.subscriptions.max: 3
- billing.subscription_items.max: 10
- billing.usage_events.monthly: 50000
- billing.invoices.monthly: 1
- billing.billing_cycles.concurrent: 1
- infra.namespaces.max: 1
- infra.retention_days: 7
- infra.isolation_level: shared
- support.level: community

## Hobby ($19)
- billing.customers.max: 1000
- billing.subscriptions.max: 2000
- billing.subscription_items.max: 10000
- billing.usage_events.monthly: 2000000
- billing.invoices.monthly: 1000
- billing.billing_cycles.concurrent: 1
- infra.namespaces.max: 1
- infra.retention_days: 30
- infra.isolation_level: dedicated-namespace
- support.level: community

## Production ($39)
- billing.customers.max: 10000
- billing.subscriptions.max: 25000
- billing.subscription_items.max: 100000
- billing.usage_events.monthly: 10000000
- billing.invoices.monthly: 10000
- billing.billing_cycles.concurrent: 3
- infra.namespaces.max: 1
- infra.retention_days: 90
- infra.isolation_level: dedicated-namespace
- support.level: email

## Performance ($99)
- billing.customers.max: 50000
- billing.subscriptions.max: 150000
- billing.subscription_items.max: 500000
- billing.usage_events.monthly: 50000000
- billing.invoices.monthly: 50000
- billing.billing_cycles.concurrent: 5
- infra.namespaces.max: 3
- infra.retention_days: 365
- infra.isolation_level: isolated
- support.level: priority
