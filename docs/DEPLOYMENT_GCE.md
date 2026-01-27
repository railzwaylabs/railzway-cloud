# Deployment Guide - Single GCE Instance

This guide covers deploying `railzway-cloud` to a single GCE instance running Nomad, Consul, Traefik, PostgreSQL, and Redis.

## Prerequisites

- GCE instance with:
  - Nomad (server + client mode)
  - Consul
  - Traefik
  - PostgreSQL
  - Redis
- SSH access to the instance
- GitHub repository with Actions enabled

## Network Configuration

Since all services run on a single node, containers cannot use `localhost` to access host services. Use one of these approaches:

### Option 1: Docker Bridge Gateway (Recommended)
Set `DB_HOST` and `PROVISION_DB_HOST` to `172.17.0.1` (Docker's default bridge gateway).

### Option 2: GCE Internal IP
Use the instance's internal IP address (e.g., `10.128.0.2`).

### Option 3: Host Network Mode
Run containers with `--network host` (less isolated but simpler).

## Environment Variables

Create `/opt/railzway/.env` on your GCE instance:

```bash
# Application
APP_SERVICE=railzway-cloud
APP_VERSION=v1.0.0
PORT=8080
ENVIRONMENT=production

# Database (Control Plane)
DB_HOST=172.17.0.1
DB_PORT=5432
DB_NAME=cloud
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_SSL_MODE=disable

# Database (Tenant Provisioning)
PROVISION_DB_HOST=172.17.0.1
PROVISION_DB_PORT=5432
PROVISION_DB_NAME=postgres
PROVISION_DB_USER=postgres
PROVISION_DB_PASSWORD=your_secure_password
PROVISION_DB_SSL_MODE=disable

# Redis (Rate Limiting for Tenants)
PROVISION_RATE_LIMIT_REDIS_ADDR=172.17.0.1:6379
PROVISION_RATE_LIMIT_REDIS_PASSWORD=
PROVISION_RATE_LIMIT_REDIS_DB=0

# OAuth (Control Plane)
OAUTH2_CLIENT_ID=your_oauth_client_id
OAUTH2_CLIENT_SECRET=your_oauth_client_secret
OAUTH2_URI=https://your-auth-provider.com
OAUTH2_CALLBACK_URL=https://cloud.railzway.com/auth/callback

# Auth Secrets
AUTH_JWT_SECRET=your_jwt_secret_here
ADMIN_API_TOKEN=your_admin_token_here
TENANT_AUTH_JWT_SECRET_KEY=your_tenant_master_key

# Nomad
NOMAD_ADDR=http://localhost:4646

# Domain Configuration
APP_ROOT_DOMAIN=railzway.com
APP_ROOT_SCHEME=https

# Railzway OSS Version
RAILZWAY_OSS_VERSION=v1.6.0
```

## GitHub Actions Setup

### 1. Generate SSH Key

On your local machine:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/railzway-deploy -C "github-actions" -N ""
```

### 2. Add Public Key to GCE

Copy the public key:

```bash
cat ~/.ssh/railzway-deploy.pub
```

On your GCE instance (via GCP Console SSH):

```bash
nano ~/.ssh/authorized_keys
# Paste the public key on a new line
# Save and exit (Ctrl+O, Enter, Ctrl+X)
```

### 3. Configure GitHub Secrets

Go to your repository: `Settings` → `Secrets and variables` → `Actions` → `New repository secret`

Add these secrets:

- **GCE_SSH_KEY**: Content of `~/.ssh/railzway-deploy` (private key)
- **GCE_HOST**: Your GCE instance's public IP
- **GCE_USERNAME**: Your SSH username (usually your Google account name or `root`)

### 4. Deploy

Push a new tag to trigger deployment:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions will:
1. Run tests
2. Build Docker image
3. Push to GitHub Container Registry
4. SSH to GCE and restart the service
5. Verify deployment via health check

## Manual Deployment (Fallback)

If you need to deploy manually:

```bash
# SSH to GCE
gcloud compute ssh your-instance-name

# Pull latest image
docker pull ghcr.io/railzwaylabs/railzway-cloud:v1.0.0

# Stop existing container
docker stop railzway-cloud || true
docker rm railzway-cloud || true

# Run new version
docker run -d \
  --name railzway-cloud \
  --restart unless-stopped \
  --env-file /opt/railzway/.env \
  -p 8080:8080 \
  ghcr.io/railzwaylabs/railzway-cloud:v1.0.0

# Check logs
docker logs -f railzway-cloud
```

## Traefik Configuration

Ensure Traefik is configured to route traffic to the control plane:

```yaml
# /etc/traefik/dynamic/railzway-cloud.yml
http:
  routers:
    railzway-cloud:
      rule: "Host(`cloud.railzway.com`)"
      service: railzway-cloud
      entryPoints:
        - web
        - websecure
      tls:
        certResolver: letsencrypt

  services:
    railzway-cloud:
      loadBalancer:
        servers:
          - url: "http://localhost:8080"
```

## Troubleshooting

### Container can't connect to PostgreSQL

**Problem**: `dial tcp 127.0.0.1:5432: connect: connection refused`

**Solution**: Change `DB_HOST` from `localhost` to `172.17.0.1` or your GCE internal IP.

### Nomad jobs fail to start

**Problem**: Deployed instances can't access database

**Solution**: Ensure `PROVISION_DB_HOST` uses the correct IP (not `localhost`).

### GitHub Actions deployment fails

**Problem**: SSH connection refused

**Solution**: 
1. Verify `GCE_HOST` is the correct public IP
2. Check GCE firewall allows SSH (port 22)
3. Verify SSH key is correctly added to `~/.ssh/authorized_keys`

## Monitoring

Check service status:

```bash
# Docker
docker ps | grep railzway-cloud
docker logs railzway-cloud

# Health check
curl http://localhost:8080/health

# Nomad (for tenant instances)
nomad status
nomad job status railzway-org-1
```
