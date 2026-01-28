# Railzway Cloud - Deployment Checklist

Complete checklist untuk deploy Railzway Cloud ke Nomad.

---

## ✅ Prerequisites (Infrastructure)

### 1. Nomad Cluster
- [ ] Nomad server running
- [ ] Nomad client dengan Docker driver
- [ ] Verify: `nomad node status`

### 2. Consul
- [ ] Consul agent running
- [ ] **`client_addr = "0.0.0.0"`** configured
- [ ] Verify: `consul members`
- [ ] Verify: `curl http://localhost:8500/v1/status/leader`

### 3. Traefik (Optional)
- [ ] Traefik running
- [ ] Configured untuk Consul catalog
- [ ] Let's Encrypt resolver (untuk SSL)
- [ ] Verify: `curl http://localhost:8081/dashboard/`

### 4. Docker
- [ ] Docker installed di Nomad clients
- [ ] Docker daemon running
- [ ] Verify: `docker ps`

### 5. Database
- [ ] PostgreSQL running
- [ ] Database `cloud` created
- [ ] User dengan permissions
- [ ] Test: `psql -h <host> -U <user> -d cloud`

### 6. Redis (Optional)
- [ ] Redis running
- [ ] Test: `redis-cli -h <host> ping`

---

## 📝 Local Preparation

### 1. Environment File
- [ ] Copy `.env.production.example` → `.env.production`
- [ ] Update all `REPLACE_ME` values:
  - [ ] Database credentials
  - [ ] Redis configuration
  - [ ] OAuth2 client ID & secret
  - [ ] OAuth2 callback URL
  - [ ] Auth service credentials
- [ ] Generate secrets: `./deployments/nomad/generate-secrets.sh .env.production`
- [ ] Verify: `grep REPLACE_ME .env.production` (should be empty)

### 2. SSH Access
- [ ] SSH key generated: `~/.ssh/railzway-deploy`
- [ ] Public key added to GCP metadata
- [ ] Username matches key comment (e.g., `github-actions`)
- [ ] Test: `ssh -i ~/.ssh/railzway-deploy github-actions@<server-ip>`

---

## 🚀 Initial Setup

### 1. Run Setup Script
```bash
./deployments/nomad/initial-setup.sh
```

- [ ] Enter server IP
- [ ] Enter SSH username
- [ ] Enter SSH key path
- [ ] Script completes successfully

**Verifies:**
- [ ] Files copied to `~/railzway/` on server
- [ ] Consul KV populated
- [ ] No errors in output

### 2. Verify Server Setup
```bash
ssh -i ~/.ssh/railzway-deploy github-actions@<server-ip>

# Check files
ls -la ~/railzway/
ls -la ~/railzway/deployments/

# Check Consul KV
consul kv get -recurse railzway-cloud/
```

- [ ] `.env` file exists
- [ ] Deployment scripts exist
- [ ] Consul KV keys populated

---

## ⚙️ GitHub Configuration

### 1. GitHub Secrets
Go to: [GitHub Secrets](https://github.com/railzwaylabs/railzway-cloud/settings/secrets/actions)

Add/Update:
- [ ] `GCE_HOST_PROD_1` = `<server-ip>`
- [ ] `GCE_USERNAME_PROD_1` = `github-actions`
- [ ] `GCE_SSH_KEY_PROD_1` = `<private-key-content>`
- [ ] `GITHUB_TOKEN` = (auto-available, verify it exists)

### 2. Verify Workflow
- [ ] `.github/workflows/release-deploy.yml` exists
- [ ] Workflow uses correct paths (`~/railzway/deployments/`)
- [ ] Workflow injects `GITHUB_TOKEN` to Consul KV

---

## 🎯 Deployment

### Option A: Automated (Recommended)
- [ ] Create feature branch
- [ ] Make changes & commit
- [ ] Create PR
- [ ] Merge to `main`
- [ ] GitHub Actions auto-deploys

### Option B: Manual (Emergency)
```bash
ssh -i ~/.ssh/railzway-deploy github-actions@<server-ip>

# Inject GitHub token
consul kv put railzway-cloud/github_token "<your-token>"

# Deploy
cd ~/railzway/deployments
./deploy.sh v1.3.0
```

---

## 🔍 Verification

### 1. Nomad Job
```bash
# Check status
nomad job status railzway-cloud

# Should show:
# Status = running
# Desired = 1
# Placed = 1
# Healthy = 1
```

- [ ] Job status = `running`
- [ ] All allocations healthy

### 2. Allocation
```bash
ALLOC_ID=$(nomad job allocs railzway-cloud | grep running | head -1 | awk '{print $1}')
nomad alloc status $ALLOC_ID
```

- [ ] Task state = `running`
- [ ] Health checks passing
- [ ] No restart events

### 3. Logs
```bash
nomad alloc logs -f $ALLOC_ID
```

- [ ] No error messages
- [ ] Application started successfully
- [ ] Database connection successful

### 4. Health Check
```bash
curl http://localhost:8080/health
```

- [ ] Returns `200 OK`
- [ ] Response body valid

### 5. Consul Service
```bash
consul catalog services | grep railzway-cloud
consul catalog service railzway-cloud
```

- [ ] Service registered
- [ ] Health check passing
- [ ] Correct tags (for Traefik)

### 6. Traefik (if configured)
```bash
curl https://cloud.railzway.com/health
```

- [ ] HTTPS working
- [ ] SSL certificate valid
- [ ] Routing correct

---

## 🆘 Troubleshooting

### Docker Pull Unauthorized
```bash
# Check GITHUB_TOKEN in Consul KV
consul kv get railzway-cloud/github_token

# If missing, add it
consul kv put railzway-cloud/github_token "<token>"

# Restart job
nomad job restart railzway-cloud
```

### Consul Connection Refused
```bash
# Edit config
sudo nano /etc/consul.d/consul.hcl
# Add: client_addr = "0.0.0.0"

# Restart
sudo systemctl restart consul
```

### Health Check Failing
```bash
# Check logs
nomad alloc logs $ALLOC_ID

# Check database connection
psql -h <db-host> -U <db-user> -d cloud

# Check Consul KV
consul kv get -recurse railzway-cloud/
```

### SSH Authentication Failed
```bash
# Verify SSH key in GCP metadata
# Ensure username matches key comment

# Test connection
ssh -i ~/.ssh/railzway-deploy github-actions@<server-ip>
```

---

## 📋 Quick Commands

```bash
# Deploy
nomad job run -var="version=v1.3.0" ~/railzway/deployments/railzway-cloud.nomad

# Status
nomad job status railzway-cloud

# Logs
nomad alloc logs -f <alloc-id>

# Restart
nomad job restart railzway-cloud

# Rollback
nomad job revert railzway-cloud <version>

# Stop
nomad job stop railzway-cloud
```

---

## ✅ Post-Deployment

- [ ] Monitor logs for 5-10 minutes
- [ ] Test all critical endpoints
- [ ] Verify database migrations ran
- [ ] Check resource usage
- [ ] Document any issues
- [ ] Update runbook if needed

---

## 🎉 Success Criteria

- ✅ Nomad job running
- ✅ Health checks passing
- ✅ No errors in logs
- ✅ Database connected
- ✅ Consul service registered
- ✅ Traefik routing working
- ✅ GitHub Actions auto-deploy working

**Deployment complete!** 🚀
