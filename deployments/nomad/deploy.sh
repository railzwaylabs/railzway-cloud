#!/bin/bash
# Quick deployment script for railzway-cloud on Nomad
# Run this on your Nomad server after infrastructure is ready

set -e

echo "🚀 Railzway Cloud - Nomad Deployment Script"
echo "============================================"
echo ""

# Configuration
DEPLOY_DIR="/opt/railzway/deployments"
ENV_FILE="/opt/railzway/.env"
VERSION="${1:-v1.2.0}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
check_command() {
  if ! command -v $1 &> /dev/null; then
    echo -e "${RED}✗ $1 not found${NC}"
    exit 1
  fi
  echo -e "${GREEN}✓ $1 found${NC}"
}

# Step 1: Verify prerequisites
echo "📋 Step 1: Verifying prerequisites..."
check_command nomad
check_command consul
check_command docker
echo ""

# Step 2: Check if .env file exists
echo "📋 Step 2: Checking environment file..."
if [ ! -f "$ENV_FILE" ]; then
  echo -e "${RED}✗ Environment file not found: $ENV_FILE${NC}"
  echo ""
  echo "Please create $ENV_FILE with required variables."
  echo "See deployments/nomad/CHECKLIST.md for template."
  exit 1
fi
echo -e "${GREEN}✓ Environment file found${NC}"
echo ""

# Step 3: Populate Consul KV
echo "📋 Step 3: Populating Consul KV..."
if [ -f "$DEPLOY_DIR/setup-consul-kv.sh" ]; then
  bash "$DEPLOY_DIR/setup-consul-kv.sh" "$ENV_FILE"
else
  echo -e "${YELLOW}⚠ setup-consul-kv.sh not found, skipping...${NC}"
fi
echo ""

# Step 4: Verify Consul KV
echo "📋 Step 4: Verifying Consul KV..."
KEYS_COUNT=$(consul kv get -recurse railzway-cloud/ | wc -l)
if [ "$KEYS_COUNT" -lt 10 ]; then
  echo -e "${YELLOW}⚠ Only $KEYS_COUNT keys found in Consul KV${NC}"
  echo "Expected at least 10 keys. Please verify setup."
else
  echo -e "${GREEN}✓ Found $KEYS_COUNT keys in Consul KV${NC}"
fi
echo ""

# Step 5: Check if Nomad job file exists
echo "📋 Step 5: Checking Nomad job file..."
if [ ! -f "$DEPLOY_DIR/railzway-cloud.nomad" ]; then
  echo -e "${RED}✗ Nomad job file not found: $DEPLOY_DIR/railzway-cloud.nomad${NC}"
  exit 1
fi
echo -e "${GREEN}✓ Nomad job file found${NC}"
echo ""

# Step 6: Deploy to Nomad
echo "📋 Step 6: Deploying to Nomad (version: $VERSION)..."
nomad job run -var="version=$VERSION" "$DEPLOY_DIR/railzway-cloud.nomad"
echo ""

# Step 7: Wait for allocation
echo "📋 Step 7: Waiting for allocation..."
sleep 5

# Step 8: Check job status
echo "📋 Step 8: Checking job status..."
nomad job status railzway-cloud
echo ""

# Step 9: Get allocation ID
echo "📋 Step 9: Getting allocation details..."
ALLOC_ID=$(nomad job allocs railzway-cloud | grep -E 'running|pending' | head -1 | awk '{print $1}')

if [ -z "$ALLOC_ID" ]; then
  echo -e "${RED}✗ No allocation found${NC}"
  echo "Check job status: nomad job status railzway-cloud"
  exit 1
fi

echo -e "${GREEN}✓ Allocation ID: $ALLOC_ID${NC}"
echo ""

# Step 10: Show logs
echo "📋 Step 10: Showing recent logs..."
echo "=================================="
nomad alloc logs "$ALLOC_ID" | tail -20
echo ""

# Step 11: Health check
echo "📋 Step 11: Performing health check..."
sleep 10

if curl -f http://localhost:8080/health &> /dev/null; then
  echo -e "${GREEN}✓ Health check passed!${NC}"
else
  echo -e "${YELLOW}⚠ Health check failed (might still be starting)${NC}"
  echo "Check logs: nomad alloc logs $ALLOC_ID"
fi
echo ""

# Summary
echo "============================================"
echo -e "${GREEN}✅ Deployment complete!${NC}"
echo ""
echo "📊 Useful commands:"
echo "  - Check status:  nomad job status railzway-cloud"
echo "  - View logs:     nomad alloc logs -f $ALLOC_ID"
echo "  - Health check:  curl http://localhost:8080/health"
echo "  - Consul UI:     http://localhost:8500/ui/dc1/services/railzway-cloud"
echo ""
