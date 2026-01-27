# Zero-Downtime Upgrade (Cloud ↔ OSS)

This document specifies how Railzway Cloud upgrades OSS instances **without billing downtime**. It is the operational contract between the control-plane and OSS.

## 1. Principles

- Cloud **never** mutates billing facts (usage, rating, invoices, ledger).
- OSS is the **deterministic billing engine**.
- All upgrades are **expand-only** migrations.
- Rollback is **routing + scheduler lease**, not DB restore.

## 2. Lifecycle State Machine (Cloud)

States:
- **READY** – instance is healthy and compatible, but not receiving production traffic.
- **MIGRATING** – Cloud is running migrations/compatibility prep.
- **SERVING** – production traffic + scheduler lease.
- **DRAINING** – stop new traffic, finish inflight jobs, release lease.
- **TERMINATED** – retired.

Transitions:

```
READY -> SERVING
READY -> MIGRATING -> SERVING
SERVING -> DRAINING -> TERMINATED | READY
```

## 3. Readiness Contract (OSS)

**/health** → liveness only.  
**/ready** → must validate:

1. Schema gate is **active**.
2. TestClock is **not running**.
3. Scheduler lease is single-holder (optional; enforced by Cloud).

If `/ready.ready=false`, Cloud must **stop routing** to that instance.

## 4. Upgrade Choreography

1. Deploy new OSS as **READY** (standby, no traffic).
2. Run **expand-only** DB migrations.
3. Warm-up instance (health + readiness).
4. Shift traffic:
   - Read first (weighted).
   - Write after stability.
5. Scheduler handoff:
   - Old instance → **DRAINING**.
   - Release scheduler lease.
   - New instance acquires lease.
6. Promote new instance to **SERVING**.
7. Terminate old instance.

## 5. Rollback

- Routing and lease return to the old instance.
- No DB restore (expand-only).
- New instance can stay READY or be terminated.

## 6. Hard Rules

Cloud must **not**:
- Write usage, rating, invoices, ledger.
- Execute pricing logic.

Cloud **may**:
- Route traffic.
- Manage lifecycle state.
- Control scheduler lease.
- Collect observability/audit.

## 7. Related Code

- `railzway/internal/server/system_readiness.go` – `/ready` endpoint
- `railzway-cloud/internal/reconciler/lifecycle_reconciler.go`
- `railzway-cloud/internal/domain/instance/entity.go`
- `railzway-cloud/internal/usecase/deployment`
