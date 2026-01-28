# Railzway Cloud - Quick Deployment Guide 🚀

## ✅ Prerequisites

Infrastructure yang sudah ready:
- ✅ Nomad cluster
- ✅ Consul (with `client_addr = "0.0.0.0"`)
- ✅ Traefik
- ✅ Docker runtime
- ✅ PostgreSQL (Nomad job)
- ✅ Redis (Nomad job)

---

## 🎯 Deployment Steps

### **Step 1: Prepare .env.production Locally**

Edit `.env.production` di local machine Anda:

```bash
# Update values yang masih REPLACE_ME:
# - OAUTH2_CLIENT_ID
# - OAUTH2_CLIENT_SECRET
# - OAUTH2_CALLBACK_URL
# - AUTH_SERVICE_CLIENT_ID
# - AUTH_SERVICE_CLIENT_SECRET
# - Database credentials
# - Redis credentials
```

**Generate secrets** (jika belum):
```bash
./deployments/nomad/generate-secrets.sh .env.production
```

---

### **Step 2: Run Initial Setup**

```bash
# Dari project root
./deployments/nomad/initial-setup.sh
```

**Prompts** (semua ada default):
- Server IP: `34.87.70.45`
- SSH username: `github-actions`
- SSH key: `~/.ssh/railzway-deploy`

**Script akan:**
1. ✅ Copy `.env.production` → `~/railzway/.env` di server
2. ✅ Copy deployment scripts ke `~/railzway/deployments/`
3. ✅ Populate Consul KV dari `.env`

---

### **Step 3: Setup GitHub Secrets**

Go to: [GitHub Secrets](https://github.com/railzwaylabs/railzway-cloud/settings/secrets/actions)

Add/Update:
```
GCE_HOST_PROD_1 = 34.87.70.45
GCE_USERNAME_PROD_1 = github-actions
GCE_SSH_KEY_PROD_1 = <content of ~/.ssh/railzway-deploy>
```

> **Note:** `GITHUB_TOKEN` sudah auto-available, digunakan untuk Docker registry auth

---

### **Step 4: Deploy via GitHub Actions**

**Option A: Merge to main (Recommended)**
```bash
git checkout main
git merge your-branch
git push
```

**Option B: Manual trigger**
- Go to Actions tab
- Select "Release and Deploy"
- Click "Run workflow"

---

## 📊 Verify Deployment

```bash
# SSH to server
ssh -i ~/.ssh/railzway-deploy github-actions@34.87.70.45

# Check Nomad job
nomad job status railzway-cloud

# Check logs
ALLOC_ID=$(nomad job allocs railzway-cloud | grep running | head -1 | awk '{print $1}')
nomad alloc logs -f $ALLOC_ID

# Health check
curl http://localhost:8080/health
```

---

## 🔧 Troubleshooting

### Docker pull unauthorized

**Cause:** GITHUB_TOKEN tidak ada di Consul KV

**Fix:**
```bash
# SSH to server
ssh -i ~/.ssh/railzway-deploy github-actions@34.87.70.45

# Check if token exists
consul kv get railzway-cloud/github_token

# If missing, GitHub Actions will inject it on next deployment
# Or manually add:
consul kv put railzway-cloud/github_token "your-github-token"
```

### PostgreSQL connection error

```bash
# Check PostgreSQL service
consul catalog service postgres

# Test connection
psql -h postgres.service.consul -U postgres -d cloud -c "SELECT 1"
```

### Consul connection refused

**Cause:** Consul tidak listening di localhost

**Fix:**
```bash
# Edit Consul config
sudo nano /etc/consul.d/consul.hcl

# Add:
# client_addr = "0.0.0.0"

# Restart
sudo systemctl restart consul
```

---

## 🎉 Success!

Setelah deploy sukses:
- ✅ Future deployments otomatis via GitHub Actions
- ✅ Merge to `main` → Auto-deploy
- ✅ Monitor via Nomad UI: `http://<server-ip>:4646`
- ✅ Monitor via GitHub Actions

**Ready to ship!** 🚀
