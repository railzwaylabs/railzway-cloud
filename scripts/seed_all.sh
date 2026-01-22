#!/usr/bin/env bash
set -euo pipefail

# Master Seeding Script for Railzway Cloud
# Runs all seeding scripts in correct order

BASE_URL="${RAILZWAY_OSS_BASE_URL:-http://localhost:8080}"
API_KEY="${RAILZWAY_OSS_API_KEY:-}"

if [[ -z "$API_KEY" ]]; then
  echo "ERROR: Set RAILZWAY_OSS_API_KEY environment variable"
  exit 1
fi

echo "🚀 Railzway Cloud Complete Seeding"
echo "===================================="
echo "Target: $BASE_URL"
echo ""

# Helper functions
create_product() {
  local code="$1" name="$2" entitlements="$3"
  curl -sS -X POST "$BASE_URL/api/products" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"code\":\"$code\",\"name\":\"$name\",\"active\":true,\"metadata\":{\"entitlements\":$entitlements}}" \
    | jq -r '.data.id' 2>/dev/null || echo ""
}

create_price() {
  local product_id="$1" code="$2" name="$3" desc="$4" badge="$5" badge_color="$6" cpu="$7" ram="$8" storage="$9" isolation="${10}" provider="${11}"
  curl -sS -X POST "$BASE_URL/api/prices" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"product_id\":\"$product_id\",\"code\":\"$code\",\"lookup_key\":\"$code\",\"name\":\"$name\",\"description\":\"$desc\",\"pricing_model\":\"FLAT\",\"billing_mode\":\"LICENSED\",\"billing_interval\":\"MONTH\",\"billing_interval_count\":1,\"tax_behavior\":\"EXCLUSIVE\",\"is_default\":true,\"active\":true,\"metadata\":{\"badge\":\"$badge\",\"badge_color\":\"$badge_color\",\"type\":\"Instance\",\"description\":\"$desc\",\"compute_provider\":\"$provider\",\"specs\":{\"cpu\":\"$cpu\",\"ram\":\"$ram\",\"storage\":\"$storage\",\"isolation\":\"$isolation\"}}}" \
    | jq -r '.data.id' 2>/dev/null || echo ""
}

create_price_amount() {
  local price_id="$1" amount="$2"
  curl -sS -X POST "$BASE_URL/api/price_amounts" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"price_id\":\"$price_id\",\"currency\":\"USD\",\"unit_amount_cents\":$amount}" > /dev/null
}

create_feature() {
  local code="$1" name="$2" desc="$3"
  curl -sS -X POST "$BASE_URL/api/features" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"code\":\"$code\",\"name\":\"$name\",\"feature_type\":\"boolean\",\"description\":\"$desc\",\"active\":true}" \
    | jq -r '.data.id' 2>/dev/null || echo ""
}

get_product_id() {
  curl -sS "$BASE_URL/api/products" -H "Authorization: Bearer $API_KEY" \
    | jq -r ".data[] | select(.code == \"$1\") | .id" 2>/dev/null || echo ""
}

get_feature_id() {
  curl -sS "$BASE_URL/api/features" -H "Authorization: Bearer $API_KEY" \
    | jq -r ".data[] | select(.code == \"$1\") | .id" 2>/dev/null || echo ""
}

link_product_features() {
  local product_id="$1"
  shift
  local feature_ids=("$@")
  local json_array="["
  for i in "${!feature_ids[@]}"; do
    [[ $i -gt 0 ]] && json_array+=","
    json_array+="\"${feature_ids[$i]}\""
  done
  json_array+="]"
  curl -sS -X PUT "$BASE_URL/api/products/$product_id/features" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"feature_ids\":$json_array}" > /dev/null
}

# ============================================================================
# STEP 1: Seed Products & Prices
# ============================================================================
echo "📦 STEP 1: Seeding Products & Prices"
echo ""

# Free Trial
echo "  → Free Trial"
FREE_ID=$(create_product "free-trial" "Free Trial" '{"billing.customers.max":3,"billing.subscriptions.max":3,"billing.subscription_items.max":10,"billing.usage_events.monthly":10000,"billing.invoices.monthly":1,"billing.billing_cycles.concurrent":1,"infra.namespaces.max":1,"infra.retention_days":7,"infra.isolation_level":"shared","support.level":"community"}')
[[ -n "$FREE_ID" ]] && FREE_PRICE=$(create_price "$FREE_ID" "free-trial-monthly" "Free Trial" "14-day trial" "Free Trial" "bg-slate-600" "0.25 vCPU" "512 MB" "5 GB" "Shared" "hetzner")
[[ -n "$FREE_PRICE" ]] && create_price_amount "$FREE_PRICE" 0

# Starter
echo "  → Starter"
STARTER_ID=$(create_product "starter" "Starter" '{"billing.customers.max":100,"billing.subscriptions.max":200,"billing.subscription_items.max":1000,"billing.usage_events.monthly":10000,"billing.invoices.monthly":100,"billing.billing_cycles.concurrent":1,"infra.namespaces.max":1,"infra.retention_days":30,"infra.isolation_level":"shared","support.level":"email"}')
[[ -n "$STARTER_ID" ]] && STARTER_PRICE=$(create_price "$STARTER_ID" "starter-monthly" "Starter" "For indie developers" "Starter" "bg-blue-600" "0.5 vCPU" "1 GB" "20 GB" "Shared" "hetzner")
[[ -n "$STARTER_PRICE" ]] && create_price_amount "$STARTER_PRICE" 1900

# Pro
echo "  → Pro"
PRO_ID=$(create_product "pro" "Pro" '{"billing.customers.max":1000,"billing.subscriptions.max":2000,"billing.subscription_items.max":10000,"billing.usage_events.monthly":50000,"billing.invoices.monthly":1000,"billing.billing_cycles.concurrent":3,"infra.namespaces.max":1,"infra.retention_days":90,"infra.isolation_level":"dedicated-namespace","support.level":"priority"}')
[[ -n "$PRO_ID" ]] && PRO_PRICE=$(create_price "$PRO_ID" "pro-monthly" "Pro" "Production-grade" "Pro" "bg-indigo-600" "1.0 vCPU" "2 GB" "50 GB" "Dedicated" "hetzner")
[[ -n "$PRO_PRICE" ]] && create_price_amount "$PRO_PRICE" 3900

# Team
echo "  → Team"
TEAM_ID=$(create_product "team" "Team" '{"billing.customers.max":10000,"billing.subscriptions.max":25000,"billing.subscription_items.max":100000,"billing.usage_events.monthly":200000,"billing.invoices.monthly":10000,"billing.billing_cycles.concurrent":5,"infra.namespaces.max":3,"infra.retention_days":365,"infra.isolation_level":"isolated","support.level":"chat"}')
[[ -n "$TEAM_ID" ]] && TEAM_PRICE=$(create_price "$TEAM_ID" "team-monthly" "Team" "Scale your team" "Team" "bg-purple-600" "2.0 vCPU" "4 GB" "150 GB" "Isolated" "gcp")
[[ -n "$TEAM_PRICE" ]] && create_price_amount "$TEAM_PRICE" 9900

# Enterprise
echo "  → Enterprise"
ENTERPRISE_ID=$(create_product "enterprise" "Enterprise" '{"billing.customers.max":100000,"billing.subscriptions.max":250000,"billing.subscription_items.max":1000000,"billing.usage_events.monthly":-1,"billing.invoices.monthly":100000,"billing.billing_cycles.concurrent":10,"infra.namespaces.max":-1,"infra.retention_days":-1,"infra.isolation_level":"dedicated","support.level":"dedicated"}')
[[ -n "$ENTERPRISE_ID" ]] && ENTERPRISE_PRICE=$(create_price "$ENTERPRISE_ID" "enterprise-monthly" "Enterprise" "Unlimited scale" "Enterprise" "bg-amber-600" "Custom" "Custom" "Custom" "Dedicated" "gcp")
[[ -n "$ENTERPRISE_PRICE" ]] && create_price_amount "$ENTERPRISE_PRICE" 20000

echo "  ✓ Products & Prices seeded"
echo ""

# ============================================================================
# STEP 2: Seed Features
# ============================================================================
echo "📦 STEP 2: Seeding Features"
echo ""

FEAT_CUSTOMERS=$(create_feature "billing.customers.max" "Max Customers" "Maximum customers allowed")
FEAT_SUBS=$(create_feature "billing.subscriptions.max" "Max Subscriptions" "Maximum subscriptions allowed")
FEAT_SUB_ITEMS=$(create_feature "billing.subscription_items.max" "Max Subscription Items" "Maximum subscription items")
FEAT_USAGE=$(create_feature "billing.usage_events.monthly" "Monthly Usage Events" "Maximum usage events per month")
FEAT_INVOICES=$(create_feature "billing.invoices.monthly" "Monthly Invoices" "Maximum invoices per month")
FEAT_CYCLES=$(create_feature "billing.billing_cycles.concurrent" "Concurrent Billing Cycles" "Maximum concurrent billing cycles")
FEAT_NAMESPACES=$(create_feature "infra.namespaces.max" "Max Namespaces" "Maximum namespaces/projects")
FEAT_RETENTION=$(create_feature "infra.retention_days" "Log Retention Days" "Log retention period")
FEAT_ISOLATION=$(create_feature "infra.isolation_level" "Isolation Level" "Infrastructure isolation level")
FEAT_SUPPORT=$(create_feature "support.level" "Support Level" "Support tier")

echo "  ✓ Features seeded"
echo ""

# ============================================================================
# STEP 3: Link Products to Features
# ============================================================================
echo "📦 STEP 3: Linking Products to Features"
echo ""

# Get product IDs
FREE_ID=$(get_product_id "free-trial")
STARTER_ID=$(get_product_id "starter")
PRO_ID=$(get_product_id "pro")
TEAM_ID=$(get_product_id "team")
ENTERPRISE_ID=$(get_product_id "enterprise")

# Get feature IDs (fetch fresh IDs, not from creation step)
FEAT_CUSTOMERS=$(get_feature_id "billing.customers.max")
FEAT_SUBS=$(get_feature_id "billing.subscriptions.max")
FEAT_SUB_ITEMS=$(get_feature_id "billing.subscription_items.max")
FEAT_USAGE=$(get_feature_id "billing.usage_events.monthly")
FEAT_INVOICES=$(get_feature_id "billing.invoices.monthly")
FEAT_CYCLES=$(get_feature_id "billing.billing_cycles.concurrent")
FEAT_NAMESPACES=$(get_feature_id "infra.namespaces.max")
FEAT_RETENTION=$(get_feature_id "infra.retention_days")
FEAT_ISOLATION=$(get_feature_id "infra.isolation_level")
FEAT_SUPPORT=$(get_feature_id "support.level")

# Link all products to all features
for PRODUCT_NAME in "Free Trial:$FREE_ID" "Starter:$STARTER_ID" "Pro:$PRO_ID" "Team:$TEAM_ID" "Enterprise:$ENTERPRISE_ID"; do
  NAME="${PRODUCT_NAME%%:*}"
  ID="${PRODUCT_NAME##*:}"
  if [[ -n "$ID" ]]; then
    echo "  → Linking $NAME"
    link_product_features "$ID" "$FEAT_CUSTOMERS" "$FEAT_SUBS" "$FEAT_SUB_ITEMS" "$FEAT_USAGE" "$FEAT_INVOICES" "$FEAT_CYCLES" "$FEAT_NAMESPACES" "$FEAT_RETENTION" "$FEAT_ISOLATION" "$FEAT_SUPPORT"
  fi
done

echo "  ✓ Product-feature links created"
echo ""

echo "✅ Complete seeding finished!"
echo ""
echo "Summary:"
echo "  • 5 Products created"
echo "  • 5 Prices created"
echo "  • 10 Features created"
echo "  • All products linked to features"
