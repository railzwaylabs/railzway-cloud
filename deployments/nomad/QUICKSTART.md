# Ready to Deploy! 🚀

## ✅ Infrastructure Status

Semua sudah ready:
- ✅ Nomad cluster
- ✅ Consul
- ✅ Traefik
- ✅ Docker runtime
- ✅ PostgreSQL (Nomad job)
- ✅ Redis (Nomad job)

---

## 🎯 Tinggal 2 Langkah!

### **Step 1: Setup Environment Variables**

```bash
# SSH ke server
ssh taufik_triantono@<server-ip>

# Buat .env file
sudo mkdir -p /opt/railzway
sudo nano /opt/railzway/.env
```

**Paste ini** (sesuaikan values):

```bash
# Database (dari Nomad job PostgreSQL)
DB_HOST=postgres.service.consul  # atau IP allocation PostgreSQL
DB_PORT=5432
DB_NAME=cloud
DB_USER=postgres
DB_PASSWORD=<your_postgres_password>
DB_SSL_MODE=disable

# Provision DB (sama dengan DB utama)
PROVISION_DB_HOST=postgres.service.consul
PROVISION_DB_PORT=5432
PROVISION_DB_NAME=cloud
PROVISION_DB_USER=postgres
PROVISION_DB_PASSWORD=<your_postgres_password>
PROVISION_DB_SSL_MODE=disable

# Redis (dari Nomad job Redis)
PROVISION_RATE_LIMIT_REDIS_ADDR=redis.service.consul:6379
PROVISION_RATE_LIMIT_REDIS_PASSWORD=
PROVISION_RATE_LIMIT_REDIS_DB=0

# OAuth2 - SESUAIKAN INI!
OAUTH2_CLIENT_ID=<your_oauth_client_id>
OAUTH2_CLIENT_SECRET=<your_oauth_client_secret>
OAUTH2_URI=<your_auth_provider_url>
OAUTH2_CALLBACK_URL=https://cloud.railzway.com/auth/callback

# Tenant OAuth - SESUAIKAN INI!
TENANT_OAUTH2_CLIENT_ID=<tenant_client_id>
TENANT_OAUTH2_CLIENT_SECRET=<tenant_client_secret>
TENANT_AUTH_JWT_SECRET_KEY=$(openssl rand -base64 32)

# Security - GENERATE RANDOM!
AUTH_JWT_SECRET=$(openssl rand -base64 32)
ADMIN_API_TOKEN=$(openssl rand -hex 32)

# Application
APP_ROOT_DOMAIN=railzway.com
APP_ROOT_SCHEME=https
```

**Generate secrets:**
```bash
# Generate random secrets
openssl rand -base64 32  # untuk JWT secrets
openssl rand -hex 32     # untuk API token
```

---

### **Step 2: Deploy!**

```bash
# Copy deployment files
scp deployments/nomad/* taufik_triantono@<server-ip>:/opt/railzway/deployments/

# SSH ke server
ssh taufik_triantono@<server-ip>

# Run deployment
cd /opt/railzway/deployments
./deploy.sh v1.2.0
```

**Script akan otomatis:**
1. Verify prerequisites ✅
2. Populate Consul KV dari .env ✅
3. Deploy Nomad job ✅
4. Health check ✅

---

## 📊 Verify Deployment

```bash
# Check job status
nomad job status railzway-cloud

# Check logs
nomad alloc logs $(nomad job allocs railzway-cloud | grep running | head -1 | awk '{print $1}')

# Health check
curl http://localhost:8080/health

# Check via Traefik
curl https://cloud.railzway.com/health
```

---

## 🔧 Jika Ada Masalah

### PostgreSQL connection error
```bash
# Check PostgreSQL service
consul catalog service postgres

# Get PostgreSQL IP
nomad job allocs postgres | grep running

# Test connection
psql -h postgres.service.consul -U postgres -d cloud -c "SELECT 1"
```

### Redis connection error
```bash
# Check Redis service
consul catalog service redis

# Test connection
redis-cli -h redis.service.consul ping
```

### Nomad job tidak start
```bash
# Check allocation
ALLOC_ID=$(nomad job allocs railzway-cloud | grep -E 'running|pending' | head -1 | awk '{print $1}')
nomad alloc status $ALLOC_ID

# Check events
nomad alloc status $ALLOC_ID | grep Events -A 20
```

---

## 🎉 Setelah Deploy Sukses

### Setup GitHub Actions (Auto-deploy)
Update GitHub Secrets:
- `GCE_HOST_PROD_1` = `<server-ip>`
- `GCE_USERNAME_PROD_1` = `taufik_triantono`
- `GCE_SSH_KEY_PROD_1` = (private key)

Setelah itu, setiap merge ke `main` → auto deploy! 🚀

### Monitor
```bash
# Watch logs
nomad alloc logs -f <alloc-id>

# Watch metrics
consul catalog service railzway-cloud -detailed
```

---

**Ready? Merge PR dan deploy!** 🎯
