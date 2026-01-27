# Checklist Persiapan Deployment Nomad

## ✅ **Prerequisites (Infrastructure)**

### 1. **Nomad Cluster**
- [ ] Nomad server running (minimal 1 server)
- [ ] Nomad client running (minimal 1 client dengan Docker driver)
- [ ] Verify: `nomad node status`

### 2. **Consul**
- [ ] Consul agent running (untuk service discovery & KV storage)
- [ ] Verify: `consul members`
- [ ] Verify KV access: `consul kv put test/key value && consul kv get test/key`

### 3. **Traefik (Optional, untuk HTTPS)**
- [ ] Traefik running sebagai Nomad job atau systemd service
- [ ] Traefik configured untuk read dari Consul catalog
- [ ] Let's Encrypt resolver configured (untuk auto SSL)
- [ ] Verify: `curl http://localhost:8081/dashboard/`

### 4. **Docker**
- [ ] Docker installed di Nomad client nodes
- [ ] Docker daemon running
- [ ] Verify: `docker ps`
- [ ] Access ke GHCR: `docker login ghcr.io`

---

## 📦 **Database Setup**

### 1. **PostgreSQL Database**
- [ ] PostgreSQL server running
- [ ] Database `cloud` created
- [ ] User dengan permissions created
- [ ] Connection test: `psql -h <host> -U <user> -d cloud`

### 2. **Provision Database (untuk customer instances)**
- [ ] Separate PostgreSQL database (atau sama dengan DB utama)
- [ ] Database created untuk customer instances
- [ ] User dengan CREATE DATABASE permission

### 3. **Redis (Optional, untuk rate limiting)**
- [ ] Redis server running
- [ ] Connection test: `redis-cli -h <host> ping`

---

## 🔐 **Secrets & Configuration**

### 1. **Environment Variables (.env file)**
Buat file `/opt/railzway/.env` dengan content:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=cloud
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_SSL_MODE=disable

# Provision Database (for customer instances)
PROVISION_DB_HOST=localhost
PROVISION_DB_PORT=5432
PROVISION_DB_NAME=provision
PROVISION_DB_USER=postgres
PROVISION_DB_PASSWORD=your_secure_password
PROVISION_DB_SSL_MODE=disable

# Rate Limiting Redis (optional)
PROVISION_RATE_LIMIT_REDIS_ADDR=localhost:6379
PROVISION_RATE_LIMIT_REDIS_PASSWORD=
PROVISION_RATE_LIMIT_REDIS_DB=0

# OAuth2 Configuration (for Cloud UI)
OAUTH2_CLIENT_ID=your_oauth2_client_id
OAUTH2_CLIENT_SECRET=your_oauth2_client_secret
OAUTH2_URI=https://your-auth-provider.com
OAUTH2_CALLBACK_URL=https://cloud.railzway.com/auth/callback

# Tenant OAuth (for deployed OSS instances)
TENANT_OAUTH2_CLIENT_ID=tenant_client_id
TENANT_OAUTH2_CLIENT_SECRET=tenant_client_secret
TENANT_AUTH_JWT_SECRET_KEY=your_master_jwt_secret_key

# Security
AUTH_JWT_SECRET=your_jwt_secret_for_cloud_ui
ADMIN_API_TOKEN=your_admin_api_token

# Application
APP_ROOT_DOMAIN=railzway.com
APP_ROOT_SCHEME=https
```

### 2. **Populate Consul KV**
```bash
# Copy setup script to server
scp deployments/nomad/setup-consul-kv.sh user@server:/opt/railzway/deployments/

# SSH to server
ssh user@server

# Run setup script
cd /opt/railzway/deployments
chmod +x setup-consul-kv.sh
./setup-consul-kv.sh /opt/railzway/.env

# Verify
consul kv get -recurse railzway-cloud/
```

---

## 📁 **File Deployment**

### 1. **Copy Nomad Job File**
```bash
# From local machine
scp deployments/nomad/railzway-cloud.nomad user@server:/opt/railzway/deployments/
```

### 2. **SQL Migrations (if needed)**
```bash
# Create directory
ssh user@server "mkdir -p /opt/railzway/sql"

# Copy migrations
scp -r sql/* user@server:/opt/railzway/sql/
```

---

## 🔑 **GitHub Secrets (untuk CI/CD)**

Go to: https://github.com/railzwaylabs/railzway-cloud/settings/secrets/actions

Add/Update:
- [ ] **GCE_HOST_PROD_1**: `<your-server-ip>`
- [ ] **GCE_USERNAME_PROD_1**: `taufik_triantono`
- [ ] **GCE_SSH_KEY_PROD_1**: 
  ```
  -----BEGIN OPENSSH PRIVATE KEY-----
  (paste full private key dari ~/.ssh/railzway-deploy)
  -----END OPENSSH PRIVATE KEY-----
  ```

---

## 🚀 **Initial Deployment**

### 1. **Manual First Deployment**
```bash
# SSH to server
ssh taufik_triantono@<server-ip>

# Navigate to deployments
cd /opt/railzway/deployments

# Run Nomad job
nomad job run -var="version=v1.2.0" railzway-cloud.nomad

# Check status
nomad job status railzway-cloud

# Get allocation ID
ALLOC_ID=$(nomad job allocs railzway-cloud | grep running | head -1 | awk '{print $1}')

# View logs
nomad alloc logs $ALLOC_ID

# Check health
curl http://localhost:8080/health
```

### 2. **Verify Service Discovery**
```bash
# Check Consul registration
consul catalog services | grep railzway-cloud

# Check service details
consul catalog service railzway-cloud
```

### 3. **Verify Traefik Routing (if configured)**
```bash
# Check if Traefik picked up the service
curl -H "Host: cloud.railzway.com" http://localhost

# Or via HTTPS (if DNS configured)
curl https://cloud.railzway.com/health
```

---

## 🔍 **Verification Checklist**

- [ ] Nomad job status = "running"
- [ ] Health check passing in Consul
- [ ] Application logs show no errors
- [ ] Database connection successful
- [ ] `/health` endpoint returns 200 OK
- [ ] Traefik routing working (if configured)
- [ ] SSL certificate issued (if using Let's Encrypt)

---

## 🎯 **Next Steps After Initial Deployment**

### 1. **Test Automated Deployment**
```bash
# Merge PR to trigger CI/CD
# Or manually trigger workflow
```

### 2. **Monitor**
```bash
# Watch job status
watch -n 2 'nomad job status railzway-cloud'

# Stream logs
nomad alloc logs -f $ALLOC_ID
```

### 3. **Scale (Optional)**
```bash
# Scale to 3 replicas for HA
nomad job scale railzway-cloud 3
```

---

## 🆘 **Troubleshooting**

### Nomad Job Won't Start
```bash
# Check allocation events
nomad alloc status $ALLOC_ID | grep Events -A 20

# Check Docker image pull
docker pull ghcr.io/railzwaylabs/railzway-cloud:v1.2.0
```

### Health Check Failing
```bash
# Check application logs
nomad alloc logs $ALLOC_ID

# Check if port is listening
netstat -tlnp | grep 8080

# Manual health check
curl -v http://localhost:8080/health
```

### Consul KV Issues
```bash
# Verify all keys exist
consul kv get -recurse railzway-cloud/

# Re-run setup script
./setup-consul-kv.sh /opt/railzway/.env
```

### Traefik Not Routing
```bash
# Check Traefik logs
journalctl -u traefik -f

# Verify service tags in Consul
consul catalog service railzway-cloud -detailed
```

---

## 📋 **Quick Command Reference**

```bash
# Deploy
nomad job run -var="version=v1.2.0" railzway-cloud.nomad

# Status
nomad job status railzway-cloud

# Logs
nomad alloc logs $(nomad job allocs railzway-cloud | grep running | head -1 | awk '{print $1}')

# Restart
nomad job restart railzway-cloud

# Stop
nomad job stop railzway-cloud

# Rollback
nomad job revert railzway-cloud 0
```
