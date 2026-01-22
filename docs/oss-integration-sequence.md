# Railzway Cloud to Railzway OSS Integration Sequences

This document explains the main flows for how Railzway Cloud uses Railzway OSS for pricing and billing. All OSS calls are made through `railzway-cloud/pkg/railzwayclient` configured via `RAILZWAY_CLIENT_URL` and `RAILZWAY_API_KEY`.

## 1. Pricing for Onboarding

Cloud UI calls the Cloud API endpoints, and the Cloud API proxies to the OSS pricing API.

```mermaid
sequenceDiagram
  participant UI as Cloud UI
  participant API as Cloud API
  participant OSS as Railzway OSS

  UI->>API: GET /api/prices
  API->>OSS: ListPrices()
  OSS-->>API: price list
  API-->>UI: price list

  UI->>API: GET /api/price_amounts?price_id=...
  API->>OSS: ListPriceAmounts(price_id)
  OSS-->>API: price amount list
  API-->>UI: price amount list
```

Code references:
- `railzway-cloud/internal/api/pricing.go`
- `railzway-cloud/pkg/railzwayclient/pricing.go`

## 2. Onboarding and Provisioning (Customer + Subscription + Activation)

Onboarding creates org and instance records in the Cloud DB, then side effects (OSS and Nomad) are executed by the Outbox Processor to keep DB state authoritative.

```mermaid
sequenceDiagram
  participant UI as Cloud UI
  participant API as Cloud API
  participant DB as Cloud DB
  participant Outbox as Outbox Processor
  participant OSS as Railzway OSS
  participant Nomad as Nomad

  UI->>API: POST /user/onboarding/initialize (price_id)
  API->>DB: create org + instance
  API->>DB: create outbox event
  API-->>UI: 200 OK

  Outbox->>DB: fetch pending event
  Outbox->>OSS: EnsureCustomer(name, email, external_id)
  OSS-->>Outbox: customer_id
  Outbox->>DB: update org.oss_customer_id

  Outbox->>OSS: CreateSubscription(customer_id, price_id)
  OSS-->>Outbox: subscription_id
  Outbox->>DB: update instance.subscription_id

  Outbox->>Nomad: Deploy(job spec)
  Nomad-->>Outbox: job registered

  Outbox->>OSS: ActivateSubscription(subscription_id)
  OSS-->>Outbox: activated
```

Code references:
- `railzway-cloud/internal/onboarding/service.go`
- `railzway-cloud/internal/outbox/processor.go`
- `railzway-cloud/pkg/railzwayclient/subscription.go`

## 3. Upgrade / Downgrade Plan

Upgrade and downgrade go through OSS for subscription changes, while infrastructure changes are done via Nomad.

```mermaid
sequenceDiagram
  participant UI as Cloud UI
  participant API as Cloud API
  participant UC as UpgradeUseCase
  participant OSS as Railzway OSS
  participant Nomad as Nomad

  UI->>API: POST /user/instance/upgrade (tier)
  API->>UC: Upgrade(org_id, target_tier)
  UC->>UC: ResolvePriceID(target_tier)
  UC->>Nomad: Deploy(new tier)
  UC->>OSS: ChangePlan(subscription_id, new_price_id)
  OSS-->>UC: ok
  UC-->>API: ok
```

Code references:
- `railzway-cloud/internal/usecase/deployment/upgrade.go`
- `railzway-cloud/internal/adapter/billing/railzway_oss/adapter.go`

## 4. Pause / Resume Billing on Stop/Start

Stopping an instance pauses the OSS subscription; starting resumes it.

```mermaid
sequenceDiagram
  participant UI as Cloud UI
  participant API as Cloud API
  participant UC as LifecycleUseCase
  participant OSS as Railzway OSS
  participant Nomad as Nomad

  UI->>API: POST /user/instance/stop
  API->>UC: Stop(org_id)
  UC->>Nomad: Stop(job)
  UC->>OSS: PauseSubscription(subscription_id)
  OSS-->>UC: ok
  UC-->>API: ok

  UI->>API: POST /user/instance/start
  API->>UC: Start(org_id)
  UC->>Nomad: Deploy(job)
  UC->>OSS: ResumeSubscription(subscription_id)
  OSS-->>UC: ok
  UC-->>API: ok
```

Code references:
- `railzway-cloud/internal/usecase/deployment/lifecycle.go`
- `railzway-cloud/internal/adapter/billing/railzway_oss/adapter.go`
