#!/bin/bash
# Initial setup script for railzway-cloud deployment
# This script is for ONE-TIME initial setup only
# After this, deployments will be handled by CI/CD (GitHub Actions)

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🚀 Railzway Cloud - Initial Server Setup"
echo "=========================================="
echo ""

# Configuration
read -p "Enter server IP [34.87.70.45]: " SERVER_IP
SERVER_IP=${SERVER_IP:-34.87.70.45}
read -p "Enter SSH username [github-actions]: " SERVER_USER
SERVER_USER=${SERVER_USER:-github-actions}
read -p "Enter path to SSH private key [~/.ssh/railzway-deploy]: " SSH_KEY
SSH_KEY=${SSH_KEY:-~/.ssh/railzway-deploy}

# Expand tilde
SSH_KEY="${SSH_KEY/#\~/$HOME}"

# Validate inputs
if [ -z "$SERVER_IP" ]; then
  echo -e "${RED}❌ Error: Server IP is required${NC}"
  exit 1
fi

if [ ! -f "$SSH_KEY" ]; then
  echo -e "${RED}❌ Error: SSH key not found: $SSH_KEY${NC}"
  exit 1
fi

if [ ! -f ".env.production" ]; then
  echo -e "${RED}❌ Error: .env.production not found${NC}"
  echo "Please run this script from the project root directory"
  exit 1
fi

echo ""
echo "📋 Configuration:"
echo "  Server IP: $SERVER_IP"
echo "  SSH User: $SERVER_USER"
echo "  SSH Key: $SSH_KEY"
echo ""

read -p "Continue with setup? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Setup cancelled."
  exit 0
fi

echo ""
echo "📦 Step 1: Copying .env.production to server..."
scp -i "$SSH_KEY" .env.production ${SERVER_USER}@${SERVER_IP}:/tmp/.env
echo -e "${GREEN}✓ .env.production copied${NC}"

echo ""
echo "📦 Step 2: Copying deployment scripts..."
scp -i "$SSH_KEY" \
  deployments/nomad/setup-consul-kv.sh \
  deployments/nomad/deploy.sh \
  deployments/nomad/railzway-cloud.nomad \
  ${SERVER_USER}@${SERVER_IP}:/tmp/
echo -e "${GREEN}✓ Deployment scripts copied${NC}"

echo ""
echo "📦 Step 3: Setting up directories on server..."
ssh -i "$SSH_KEY" ${SERVER_USER}@${SERVER_IP} << 'EOF'
  # Create directories with sudo
  sudo mkdir -p /opt/railzway/deployments
  sudo chown -R ${USER}:${USER} /opt/railzway
  
  # Move .env file
  sudo mv /tmp/.env /opt/railzway/.env
  
  # Move deployment scripts
  sudo mv /tmp/setup-consul-kv.sh /opt/railzway/deployments/
  sudo mv /tmp/deploy.sh /opt/railzway/deployments/
  sudo mv /tmp/railzway-cloud.nomad /opt/railzway/deployments/
  
  # Set permissions
  sudo chmod +x /opt/railzway/deployments/*.sh
  
  # Verify
  echo ""
  echo "✅ Files setup complete:"
  sudo ls -lh /opt/railzway/
  echo ""
  sudo ls -lh /opt/railzway/deployments/
EOF
echo -e "${GREEN}✓ Server directories setup complete${NC}"

echo ""
echo "📦 Step 4: Setting up Consul KV..."
ssh -i "$SSH_KEY" ${SERVER_USER}@${SERVER_IP} << 'EOF'
  cd /opt/railzway/deployments
  ./setup-consul-kv.sh /opt/railzway/.env
EOF
echo -e "${GREEN}✓ Consul KV populated${NC}"

echo ""
echo "=========================================="
echo -e "${GREEN}✅ Initial setup complete!${NC}"
echo ""
echo "📋 Next steps:"
echo ""
echo "1. Verify server setup:"
echo "   ssh -i $SSH_KEY ${SERVER_USER}@${SERVER_IP}"
echo "   sudo ls -la /opt/railzway"
echo "   consul kv get -recurse railzway-cloud/"
echo ""
echo "2. Setup GitHub Secrets for CI/CD:"
echo "   Go to: https://github.com/railzwaylabs/railzway-cloud/settings/secrets/actions"
echo "   Add/Update these secrets:"
echo "   - GCE_HOST_PROD_1 = $SERVER_IP"
echo "   - GCE_USERNAME_PROD_1 = $SERVER_USER"
echo "   - GCE_SSH_KEY_PROD_1 = (content of $SSH_KEY)"
echo "   - GITHUB_TOKEN = (already exists, used for Docker registry auth)"
echo ""
echo "3. Deploy via GitHub Actions:"
echo "   Option A: Merge PR to 'main' → Auto-deploy"
echo "   Option B: Manually trigger workflow with specific version"
echo ""
echo "4. Monitor deployment:"
echo "   - GitHub Actions: https://github.com/railzwaylabs/railzway-cloud/actions"
echo "   - Nomad UI: http://$SERVER_IP:4646"
echo ""
echo "🎉 All future deployments will be automatic via GitHub Actions!"
echo ""

