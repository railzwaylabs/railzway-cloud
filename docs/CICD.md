# CI/CD Setup & Workflows

Railzway Cloud uses **GitHub Actions** with **Semantic Release** for fully automated Continuous Integration and Continuous Deployment.

## Automated Release & Deployment Flow

```
Developer → Commit (conventional) → Push → PR → Merge to main
  ↓
Semantic Release analyzes commits
  ↓
Auto-create tag (v1.2.3)
  ↓
Build Docker image → Push to GHCR
  ↓
Deploy to GCE → Health check
  ↓
Generate CHANGELOG.md
```

## Workflows

### 1. CI - Verify (`ci-verify.yml`)
- **Triggers**: Pull Requests, Push to any branch
- **Tasks**: 
  - Linting (`golangci-lint`)
  - Testing (`go test -race`)
  - Build Verification
  - Security Check (`govulncheck`)
- **Purpose**: Ensure code quality before merge

### 2. Release & Deploy (`release-deploy.yml`)
- **Triggers**: Push to `main` (after PR merge)
- **Steps**:
  1. **Test**: Run full test suite with coverage check
  2. **Release**: Semantic Release analyzes commits and creates tag
  3. **Build**: Build Docker image and push to GitHub Container Registry
  4. **Deploy**: SSH to GCE, pull image, restart container
  5. **Verify**: Health check to confirm deployment
- **Target**: Production GCE instance
- **Versioning**: Automatic based on conventional commits

### 3. Infra - Nomad Node (`infra-nomad-init.yml`)
- **Triggers**: Manual (Workflow Dispatch)
- **Tasks**: Provisions a new Compute Engine VM with Nomad and Docker

## Conventional Commits

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for automated versioning.

**Quick Reference:**
```bash
feat: add new feature       # → v1.1.0 (minor)
fix: resolve bug            # → v1.0.1 (patch)
feat!: breaking change      # → v2.0.0 (major)
docs: update README         # → no release
test: add unit tests        # → no release
```

See [CONVENTIONAL_COMMITS.md](./CONVENTIONAL_COMMITS.md) for detailed guide.

## Setup Guide

### GitHub Secrets (Required)

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `GCE_SSH_KEY_PROD_1` | SSH private key for GCE access | (contents of `~/.ssh/railzway-deploy`) |
| `GCE_HOST_PROD_1` | Public IP of GCE instance | `34.101.xxx.xxx` |
| `GCE_USERNAME_PROD_1` | SSH username | `taufiktriantono` |

### SSH Key Setup

1. **Generate SSH key** (on local machine):
   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/railzway-deploy -C "github-actions" -N ""
   ```

2. **Add public key to GCE**:
   ```bash
   # Copy public key
   cat ~/.ssh/railzway-deploy.pub
   
   # On GCE instance (via GCP Console SSH):
   nano ~/.ssh/authorized_keys
   # Paste the public key
   ```

3. **Add secrets to GitHub**:
   - Go to: `Settings` → `Secrets and variables` → `Actions`
   - Add `GCE_SSH_KEY_PROD_1`, `GCE_HOST_PROD_1`, `GCE_USERNAME_PROD_1`

### GCE Instance Setup

Ensure `/opt/railzway/.env` exists on GCE with required environment variables.

See [DEPLOYMENT_GCE.md](./DEPLOYMENT_GCE.md) for complete GCE setup guide.

## Deployment Verification

After each deployment, the workflow automatically:
1. Waits 10 seconds for service startup
2. Runs health check: `curl http://localhost:8080/health`
3. Fails deployment if health check fails
4. Shows container logs on failure

## Manual Deployment (Fallback)

If automated deployment fails:

```bash
# SSH to GCE
ssh -i ~/.ssh/railzway-deploy user@GCE_IP

# Pull latest image
docker pull ghcr.io/railzwaylabs/railzway-cloud:v1.2.3

# Stop and remove old container
docker stop railzway-cloud
docker rm railzway-cloud

# Start new container
docker run -d \
  --name railzway-cloud \
  --restart unless-stopped \
  --env-file /opt/railzway/.env \
  --network host \
  ghcr.io/railzwaylabs/railzway-cloud:v1.2.3

# Verify
curl http://localhost:8080/health
```

## Troubleshooting

### Deployment fails with "Context access might be invalid"
- **Cause**: GitHub Secrets not configured
- **Fix**: Add all required secrets (`GCE_SSH_KEY_PROD_1`, `GCE_HOST_PROD_1`, `GCE_USERNAME_PROD_1`)

### Health check fails
- **Cause**: Service not starting or `/opt/railzway/.env` missing
- **Fix**: Check container logs: `docker logs railzway-cloud`

### No release created
- **Cause**: Commits don't follow conventional format
- **Fix**: Use `feat:`, `fix:`, `perf:`, or `refactor:` prefixes

---

For more details on deployment architecture, see [DEPLOYMENT_GCE.md](./DEPLOYMENT_GCE.md).
