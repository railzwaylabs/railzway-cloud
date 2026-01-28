# Railzway Cloud - Nomad Deployment Setup

Complete setup guide for deploying Railzway Cloud to Nomad.

---

## Prerequisites

1. ✅ Nomad cluster running
2. ✅ Consul with `client_addr = "0.0.0.0"`
3. ✅ Traefik as ingress controller
4. ✅ Docker runtime on Nomad clients
5. ✅ PostgreSQL (Nomad job or external)
6. ✅ Redis (Nomad job or external)

---

## Initial Setup

### Step 1: Prepare Environment File

Edit `.env.production` locally:

```bash
# Update all REPLACE_ME values:
# - Database credentials
# - Redis configuration
# - OAuth2 credentials
# - Auth service credentials

# Generate secrets
./deployments/nomad/generate-secrets.sh .env.production
```

### Step 2: Run Initial Setup Script

```bash
./deployments/nomad/initial-setup.sh
```

**What it does:**
1. Copy `.env.production` → `~/railzway/.env` on server
2. Copy deployment scripts to `~/railzway/deployments/`
3. Populate Consul KV from `.env`

**Prompts:**
- Server IP (default: `34.87.70.45`)
- SSH username (default: `github-actions`)
- SSH key path (default: `~/.ssh/railzway-deploy`)

### Step 3: Configure GitHub Secrets

Go to: [GitHub Secrets](https://github.com/railzwaylabs/railzway-cloud/settings/secrets/actions)

Add/Update:
```
GCE_HOST_PROD_1 = <server-ip>
GCE_USERNAME_PROD_1 = github-actions
GCE_SSH_KEY_PROD_1 = <private-key-content>
```

> **Note:** `GITHUB_TOKEN` is auto-available and used for Docker registry authentication

---

## Deployment

### Automated (Recommended)

Merge to `main` branch → GitHub Actions auto-deploys:

1. Semantic release creates version tag
2. Docker image built and pushed to GHCR
3. GitHub Actions injects `GITHUB_TOKEN` to Consul KV
4. Nomad job deployed via SSH
5. Rolling update with health checks

### Manual (Emergency)

```bash
ssh -i ~/.ssh/railzway-deploy github-actions@<server-ip>

# Ensure GITHUB_TOKEN is in Consul KV
consul kv put railzway-cloud/github_token "<your-github-token>"

# Deploy specific version
cd ~/railzway/deployments
./deploy.sh v1.3.0
```

---

## Verification

```bash
# Check job status
nomad job status railzway-cloud

# Check allocation
ALLOC_ID=$(nomad job allocs railzway-cloud | grep running | head -1 | awk '{print $1}')
nomad alloc status $ALLOC_ID

# View logs
nomad alloc logs -f $ALLOC_ID

# Health check
curl http://localhost:8080/health
```

---

## Troubleshooting

### SSH Authentication Failed

```bash
# Verify SSH key is added to GCP metadata
# Username must match SSH key comment (e.g., github-actions)

# Test connection
ssh -i ~/.ssh/railzway-deploy github-actions@<server-ip>
```

### Docker Pull Unauthorized

**Cause:** Missing GITHUB_TOKEN in Consul KV

**Fix:**
```bash
# GitHub Actions will inject it automatically
# Or manually add:
consul kv put railzway-cloud/github_token "<token>"
```

### Consul Connection Refused

**Cause:** Consul not listening on localhost

**Fix:**
```bash
# Edit Consul config
sudo nano /etc/consul.d/consul.hcl

# Add:
client_addr = "0.0.0.0"

# Restart
sudo systemctl restart consul
```

### Nomad Job Failed

```bash
# Check events
nomad alloc status <alloc-id> | grep Events -A 20

# View logs
nomad alloc logs <alloc-id>

# Check Consul KV
consul kv get -recurse railzway-cloud/
```

---

## Rollback

```bash
# Revert to previous version
nomad job revert railzway-cloud <version-number>

# Or deploy specific version
nomad job run -var="version=v1.2.0" ~/railzway/deployments/railzway-cloud.nomad
```

---

## Monitoring

```bash
# Watch deployment
watch -n 2 'nomad job status railzway-cloud'

# Stream logs
nomad alloc logs -f <alloc-id>

# Resource usage
nomad alloc status <alloc-id> | grep Resources -A 10

# Nomad UI
http://<server-ip>:4646
```

---

## File Locations

### On Server
```
~/railzway/
├── .env                     # Environment variables
└── deployments/
    ├── setup-consul-kv.sh  # Consul KV setup
    ├── deploy.sh           # Deployment script
    └── railzway-cloud.nomad # Nomad job definition
```

### In Repository
```
deployments/nomad/
├── initial-setup.sh        # One-time setup script
├── setup-consul-kv.sh     # Consul KV population
├── deploy.sh              # Deployment script
├── railzway-cloud.nomad   # Nomad job file
├── generate-secrets.sh    # Secret generation helper
├── QUICKSTART.md          # Quick start guide
├── SETUP.md               # This file
└── CHECKLIST.md           # Deployment checklist
```
